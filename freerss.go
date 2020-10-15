package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	xhtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"html"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mmcdole/gofeed"
)

type PrintFunc func(format string, a ...interface{}) (n int, err error)

type Feed struct {
	Title   string    `json:"title"`
	Url     string    `json:"url"`
	Desc    string    `json:"desc"`
	Pubdate string    `json:"pubdate"`
	Pubtime time.Time `json:"-"`
	Entries []*Entry  `json:"entries"`
}
type Entry struct {
	Title   string    `json:"title"`
	Url     string    `json:"url"`
	Desc    string    `json:"desc"`
	Body    string    `json:"body"`
	Author  string    `json:"author"`
	Pubdate string    `json:"pubdate"`
	Pubtime time.Time `json:"-"`
}

func (f *Feed) String() string {
	bs, err := json.MarshalIndent(f, "", "\t")
	if err != nil {
		return ""
	}
	return string(bs)
}
func (e *Entry) String() string {
	bs, err := json.MarshalIndent(e, "", "\t")
	if err != nil {
		return ""
	}
	return string(bs)
}
func parseFeed(gfparser *gofeed.Parser, body string, maxitems int) (*Feed, error) {
	gf, err := gfparser.ParseString(body)
	if err != nil {
		return nil, err
	}

	f := Feed{}
	f.Title = gf.Title
	f.Url = gf.Link
	f.Desc = gf.Description
	convdate(gf.PublishedParsed, &f.Pubtime, &f.Pubdate)

	if maxitems == 0 {
		maxitems = len(gf.Items)
	}

	for i, it := range gf.Items {
		e := Entry{}
		e.Title = it.Title
		e.Url = it.Link
		e.Desc = it.Description
		e.Body = it.Content
		convdate(it.PublishedParsed, &e.Pubtime, &e.Pubdate)

		f.Entries = append(f.Entries, &e)

		if i >= maxitems-1 {
			break
		}
	}
	return &f, nil
}
func convdate(t *time.Time, dt *time.Time, sdt *string) {
	if t != nil {
		*dt = *t
		*sdt = dt.Format(time.RFC3339)
	}
}

func main() {
	err := run(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func rundiscoverrss(args []string) error {
	if len(args) == 0 {
		return errors.New("Please specify a feed url")
	}
	qurl := args[0]
	feeds, err := discoverfeeds(qurl)
	if err != nil {
		return err
	}

	if len(feeds) == 0 {
		fmt.Println("No feeds found.")
		return nil
	}

	fmt.Println("Found feeds:")
	for _, feed := range feeds {
		fmt.Println(feed)
	}
	return nil
}

func run(args []string) error {
	sw, _ := parseArgs(args)

	// [-i new_file]  Create and initialize db file
	if sw["i"] != "" {
		dbfile := sw["i"]
		if fileExists(dbfile) {
			return fmt.Errorf("File '%s' already exists. Can't initialize it.\n", dbfile)
		}
		createTables(dbfile)
		return nil
	}

	/*
	   	// Need to specify a db file as first parameter.
	   	if len(parms) == 0 {
	   		s := `Usage:

	   Start webservice using database file:
	   	t5 <db file>

	   Initialize new database file:
	   	t5 -i <new db file>
	   `
	   		fmt.Printf(s)
	   		return nil
	   	}

	   	// Exit if db file doesn't exist.
	   	dbfile := parms[0]
	   	if !fileExists(dbfile) {
	   		return fmt.Errorf(`Database file '%s' doesn't exist. Create one using:
	   	wb -i <notes.db>
	   `, dbfile)
	   	}

	   	db, err := sql.Open("sqlite3", dbfile)
	   	if err != nil {
	   		return fmt.Errorf("Error opening '%s' (%s)\n", dbfile, err)
	   	}
	*/

	gfparser := gofeed.NewParser()

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "./static/radio.ico") })
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./"))))
	http.HandleFunc("/api/feed/", feedHandler(nil, gfparser))
	http.HandleFunc("/api/discoverfeed/", discoverfeedHandler(nil, gfparser))

	port := "8000"
	fmt.Printf("Listening on %s...\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	return err
}

func createTables(newfile string) {
	if fileExists(newfile) {
		s := fmt.Sprintf("File '%s' already exists. Can't initialize it.\n", newfile)
		fmt.Printf(s)
		os.Exit(1)
	}

	db, err := sql.Open("sqlite3", newfile)
	if err != nil {
		fmt.Printf("Error opening '%s' (%s)\n", newfile, err)
		os.Exit(1)
	}

	ss := []string{
		"CREATE TABLE user (user_id INTEGER PRIMARY KEY NOT NULL, username TEXT UNIQUE, password TEXT, active INTEGER NOT NULL, email TEXT);",
		"INSERT INTO user (user_id, username, password, active, email) VALUES (1, 'admin', '', 1, '');",
	}

	tx, err := db.Begin()
	if err != nil {
		log.Printf("DB error (%s)\n", err)
		os.Exit(1)
	}
	for _, s := range ss {
		_, err := txexec(tx, s)
		if err != nil {
			tx.Rollback()
			log.Printf("DB error (%s)\n", err)
			os.Exit(1)
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Printf("DB error (%s)\n", err)
		os.Exit(1)
	}
}

//*** DB functions ***
func sqlstmt(db *sql.DB, s string) *sql.Stmt {
	stmt, err := db.Prepare(s)
	if err != nil {
		log.Fatalf("db.Prepare() sql: '%s'\nerror: '%s'", s, err)
	}
	return stmt
}
func sqlexec(db *sql.DB, s string, pp ...interface{}) (sql.Result, error) {
	stmt := sqlstmt(db, s)
	defer stmt.Close()
	return stmt.Exec(pp...)
}
func txstmt(tx *sql.Tx, s string) *sql.Stmt {
	stmt, err := tx.Prepare(s)
	if err != nil {
		log.Fatalf("tx.Prepare() sql: '%s'\nerror: '%s'", s, err)
	}
	return stmt
}
func txexec(tx *sql.Tx, s string, pp ...interface{}) (sql.Result, error) {
	stmt := txstmt(tx, s)
	defer stmt.Close()
	return stmt.Exec(pp...)
}

//*** Helper functions ***

// Helper function to make fmt.Fprintf(w, ...) calls shorter.
// Ex.
// Replace:
//   fmt.Fprintf(w, "<p>Some text %s.</p>", str)
//   fmt.Fprintf(w, "<p>Some other text %s.</p>", str)
// with the shorter version:
//   P := makeFprintf(w)
//   P("<p>Some text %s.</p>", str)
//   P("<p>Some other text %s.</p>", str)
func makeFprintf(w io.Writer) func(format string, a ...interface{}) (n int, err error) {
	return func(format string, a ...interface{}) (n int, err error) {
		return fmt.Fprintf(w, format, a...)
	}
}
func listContains(ss []string, v string) bool {
	for _, s := range ss {
		if v == s {
			return true
		}
	}
	return false
}
func fileExists(file string) bool {
	_, err := os.Stat(file)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
func makePrintFunc(w io.Writer) func(format string, a ...interface{}) (n int, err error) {
	// Return closure enclosing io.Writer.
	return func(format string, a ...interface{}) (n int, err error) {
		return fmt.Fprintf(w, format, a...)
	}
}
func atoi(s string) int {
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}
func idtoi(sid string) int64 {
	return int64(atoi(sid))
}
func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}
func atof(s string) float64 {
	if s == "" {
		return 0.0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0
	}
	return f
}

func unescapeUrl(qurl string) string {
	returl := "/"
	if qurl != "" {
		returl, _ = url.QueryUnescape(qurl)
	}
	return returl
}
func escape(s string) string {
	return html.EscapeString(s)
}

func parseArgs(args []string) (map[string]string, []string) {
	switches := map[string]string{}
	parms := []string{}

	standaloneSwitches := []string{}
	definitionSwitches := []string{"i"}
	fNoMoreSwitches := false
	curKey := ""

	for _, arg := range args {
		if fNoMoreSwitches {
			// any arg after "--" is a standalone parameter
			parms = append(parms, arg)
		} else if arg == "--" {
			// "--" means no more switches to come
			fNoMoreSwitches = true
		} else if strings.HasPrefix(arg, "--") {
			switches[arg[2:]] = "y"
			curKey = ""
		} else if strings.HasPrefix(arg, "-") {
			if listContains(definitionSwitches, arg[1:]) {
				// -a "val"
				curKey = arg[1:]
				continue
			}
			for _, ch := range arg[1:] {
				// -a, -b, -ab
				sch := string(ch)
				if listContains(standaloneSwitches, sch) {
					switches[sch] = "y"
				}
			}
		} else if curKey != "" {
			switches[curKey] = arg
			curKey = ""
		} else {
			// standalone parameter
			parms = append(parms, arg)
		}
	}

	return switches, parms
}

func handleErr(w http.ResponseWriter, err error, sfunc string) {
	log.Printf("%s: server error (%s)\n", sfunc, err)
	http.Error(w, fmt.Sprintf("%s", err), 500)
}
func handleDbErr(w http.ResponseWriter, err error, sfunc string) bool {
	if err == sql.ErrNoRows {
		http.Error(w, "Not found.", 404)
		return true
	}
	if err != nil {
		log.Printf("%s: database error (%s)\n", sfunc, err)
		http.Error(w, "Server database error.", 500)
		return true
	}
	return false
}
func handleTxErr(tx *sql.Tx, err error) bool {
	if err != nil {
		tx.Rollback()
		return true
	}
	return false
}

func feedHandler(db *sql.DB, gfparser *gofeed.Parser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		qurl := unescapeUrl(r.FormValue("url"))
		if qurl == "" {
			http.Error(w, "?url=<feedurl> required", 401)
			return
		}
		qmaxitems := atoi(r.FormValue("maxitems"))

		res, err := http.Get(qurl)
		if err != nil {
			http.Error(w, fmt.Sprintf("Not found: %s", qurl), 404)
			return
		}
		defer res.Body.Close()
		bs, err := ioutil.ReadAll(res.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("error reading feed (%s)", err), 404)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		P := makeFprintf(w)
		f, err := parseFeed(gfparser, string(bs), qmaxitems)
		if err != nil {
			handleErr(w, err, "feedHandler")
			return
		}
		P("%s\n", f)
	}
}

func discoverfeedHandler(db *sql.DB, gfparser *gofeed.Parser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		qurl := unescapeUrl(r.FormValue("url"))
		if qurl == "" {
			http.Error(w, "?url=<feedurl> required", 401)
			return
		}

		feeds, err := discoverfeeds(qurl)
		if err != nil {
			handleErr(w, err, "discoverfeedHandler")
			return
		}
		bs, err := json.MarshalIndent(feeds, "", "\t")
		if err != nil {
			handleErr(w, err, "discoverfeedHandler")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		P := makeFprintf(w)
		P("%s\n", string(bs))
	}
}
func discoverfeeds(qurl string) ([]string, error) {
	res, err := http.Get(qurl)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bs, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	feeds := []string{}

	// Check if url is already an rss feed.
	gfparser := gofeed.NewParser()
	if isValidFeed(gfparser, bs) {
		feeds = append(feeds, qurl)
	}
	ubase, _ := url.Parse(qurl)

	surls := getFeedLinks(bs)
	for _, surl := range surls {
		surl = completeFeedUrl(ubase, surl)
		feeds = append(feeds, surl)
	}

	return feeds, nil
}
func getAttr(tok xhtml.Token, k string) string {
	for _, attr := range tok.Attr {
		if attr.Key == k {
			return attr.Val
		}
	}
	return ""
}
func isValidFeed(gfparser *gofeed.Parser, bs []byte) bool {
	_, err := parseFeed(gfparser, string(bs), 0)
	if err != nil {
		return false
	}
	return true
}
func getFeedLinks(bs []byte) []string {
	var feeds []string

	z := xhtml.NewTokenizer(bytes.NewReader(bs))
	for {
		tt := z.Next()
		if tt == xhtml.ErrorToken {
			break
		}

		tok := z.Token()
		if tok.DataAtom != atom.Link {
			continue
		}
		stype := getAttr(tok, "type")
		if stype != "application/rss+xml" && stype != "application/atom+xml" {
			continue
		}
		href := getAttr(tok, "href")
		if href == "" {
			continue
		}
		feeds = append(feeds, href)
	}

	return feeds
}
func completeFeedUrl(ubase *url.URL, sfeedurl string) string {
	ufeed, _ := url.Parse(sfeedurl)
	if ufeed.Scheme == "" {
		ufeed.Scheme = ubase.Scheme
	}
	if ufeed.Host == "" {
		ufeed.Host = ubase.Host
	}
	// if feed is relative to baseurl
	if !strings.HasPrefix(ufeed.Path, "/") {
		ufeed.Path = path.Join(ubase.Path, ufeed.Path)
	}
	return ufeed.String()
}
