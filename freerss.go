package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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

	xhtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/gorilla/feeds"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mmcdole/gofeed"
	"golang.org/x/crypto/bcrypt"
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
type User struct {
	Userid    int64
	Username  string
	HashedPwd string
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

func runtesthash(args []string) error {
	if len(args) == 0 {
		return errors.New("Please specify a username")
	}
	if len(args) == 1 {
		username := args[0]
		shash := genHash(username)
		fmt.Printf("%s\n", shash)
		return nil
	}

	username := args[0]
	shash := args[1]
	if validateHash(shash, username) {
		fmt.Printf("validate ok\n")
	} else {
		fmt.Printf("not validated\n")
	}
	return nil
}
func runtestsignup(args []string) error {
	sw, parms := parseArgs(args)
	// [-i new_file]  Create and initialize db file
	if sw["i"] != "" {
		dbfile := sw["i"]
		if fileExists(dbfile) {
			return fmt.Errorf("File '%s' already exists. Can't initialize it.\n", dbfile)
		}
		createTables(dbfile)
		return nil
	}

	// Need to specify a db file as first parameter.
	if len(parms) == 0 {
		return errors.New("Specify a db file")
	}

	// Exit if db file doesn't exist.
	dbfile := parms[0]
	if !fileExists(dbfile) {
		return fmt.Errorf(`Database file '%s' doesn't exist. Create one using:
	freerss -i <notes.db>
   `, dbfile)
	}

	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return fmt.Errorf("Error opening '%s' (%s)\n", dbfile, err)
	}

	if len(parms) < 3 {
		return fmt.Errorf("Specify a username and password")
	}
	username := parms[1]
	pwd := parms[2]
	err = signup(db, username, pwd)
	if err != nil {
		return err
	}
	return nil
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
	sw, parms := parseArgs(args)

	// [-i new_file]  Create and initialize db file
	if sw["i"] != "" {
		dbfile := sw["i"]
		if fileExists(dbfile) {
			return fmt.Errorf("File '%s' already exists. Can't initialize it.\n", dbfile)
		}
		createTables(dbfile)
		return nil
	}

	// Need to specify a db file as first parameter.
	if len(parms) == 0 {
		s := `Usage:

   Start webservice using database file:
	freerss <db file> [port]

   Initialize new database file:
	freerss -i <new db file>

`
		fmt.Printf(s)
		return nil
	}

	// Exit if db file doesn't exist.
	dbfile := parms[0]
	if !fileExists(dbfile) {
		return fmt.Errorf(`Database file '%s' doesn't exist. Create one using:
	freerss -i <instance.db>
   `, dbfile)
	}

	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return fmt.Errorf("Error opening '%s' (%s)\n", dbfile, err)
	}

	gfparser := gofeed.NewParser()

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "./static/radio.ico") })
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./"))))
	http.HandleFunc("/api/feed/", feedHandler(nil, gfparser))
	http.HandleFunc("/api/discoverfeed/", discoverfeedHandler(nil, gfparser))
	http.HandleFunc("/api/login/", loginHandler(db))
	http.HandleFunc("/api/signup/", signupHandler(db))
	http.HandleFunc("/api/edituser/", edituserHandler(db))
	http.HandleFunc("/api/deluser/", deluserHandler(db))
	http.HandleFunc("/api/savegrid/", savegridHandler(db))
	http.HandleFunc("/api/loadgrid/", loadgridHandler(db))
	http.HandleFunc("/api/testfeed/", testfeedHandler(db))

	port := "8000"
	if len(parms) > 1 {
		port = parms[1]
	}
	fmt.Printf("Listening on %s...\n", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
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
		"CREATE TABLE user (user_id INTEGER PRIMARY KEY NOT NULL, username TEXT UNIQUE, password TEXT);",
		"INSERT INTO user (user_id, username, password) VALUES (1, 'admin', '');",
		"CREATE TABLE savedgrid (user_id INTEGER PRIMARY KEY NOT NULL, gridjson TEXT);",
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

func genHash(sinput string) string {
	bsHash, err := bcrypt.GenerateFromPassword([]byte(sinput), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(bsHash)
}
func validateHash(shash, sinput string) bool {
	if shash == "" && sinput == "" {
		return true
	}
	err := bcrypt.CompareHashAndPassword([]byte(shash), []byte(sinput))
	if err != nil {
		return false
	}
	return true
}

func findUser(db *sql.DB, username string) *User {
	s := "SELECT user_id, username, password FROM user WHERE username = ?"
	row := db.QueryRow(s, username)
	var u User
	err := row.Scan(&u.Userid, &u.Username, &u.HashedPwd)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return nil
	}
	return &u
}
func isUsernameExists(db *sql.DB, username string) bool {
	if findUser(db, username) == nil {
		return false
	}
	return true
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

func genTok(u *User) string {
	tok := genHash(fmt.Sprintf("%s_%s", u.Username, u.HashedPwd))
	return tok
}
func validateTok(tok string, u *User) bool {
	return validateHash(tok, fmt.Sprintf("%s_%s", u.Username, u.HashedPwd))
}

var ErrLoginIncorrect = errors.New("Incorrect username or password")

func login(db *sql.DB, username, pwd string) (string, error) {
	var u User
	s := "SELECT user_id, username, password FROM user WHERE username = ?"
	row := db.QueryRow(s, username)
	err := row.Scan(&u.Userid, &u.Username, &u.HashedPwd)
	if err == sql.ErrNoRows {
		return "", ErrLoginIncorrect
	}
	if err != nil {
		return "", err
	}
	if !validateHash(u.HashedPwd, pwd) {
		return "", ErrLoginIncorrect
	}

	// Return session token, this will be used to authenticate username
	// on every request by calling validateTok()
	tok := genTok(&u)
	return tok, nil
}

type LoginResult struct {
	Tok   string `json:"tok"`
	Error string `json:"error"`
}

func loginHandler(db *sql.DB) http.HandlerFunc {
	type LoginReq struct {
		Username string `json:"username"`
		Pwd      string `json:"pwd"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Use POST method", 401)
			return
		}
		bs, err := ioutil.ReadAll(r.Body)
		if err != nil {
			handleErr(w, err, "loginHandler")
			return
		}
		var loginreq LoginReq
		err = json.Unmarshal(bs, &loginreq)
		if err != nil {
			handleErr(w, err, "loginHandler")
			return
		}

		var result LoginResult
		tok, err := login(db, loginreq.Username, loginreq.Pwd)
		if err != nil {
			result.Error = fmt.Sprintf("%s", err)
		}
		result.Tok = tok

		w.Header().Set("Content-Type", "application/json")
		P := makeFprintf(w)
		bs, _ = json.MarshalIndent(result, "", "\t")
		P("%s\n", string(bs))
	}
}

func signup(db *sql.DB, username, pwd string) error {
	if isUsernameExists(db, username) {
		return fmt.Errorf("username '%s' already exists", username)
	}

	hashedPwd := genHash(pwd)
	s := "INSERT INTO user (username, password) VALUES (?, ?);"
	_, err := sqlexec(db, s, username, hashedPwd)
	if err != nil {
		return fmt.Errorf("DB error creating user: %s", err)
	}
	return nil
}
func signupHandler(db *sql.DB) http.HandlerFunc {
	type SignupReq struct {
		Username string `json:"username"`
		Pwd      string `json:"pwd"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Use POST method", 401)
			return
		}

		bs, err := ioutil.ReadAll(r.Body)
		if err != nil {
			handleErr(w, err, "signupHandler")
			return
		}
		var signupreq SignupReq
		err = json.Unmarshal(bs, &signupreq)
		if err != nil {
			handleErr(w, err, "signupHandler")
			return
		}
		if signupreq.Username == "" {
			http.Error(w, "username required", 401)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		P := makeFprintf(w)

		// Attempt to sign up new user.
		var result LoginResult
		err = signup(db, signupreq.Username, signupreq.Pwd)
		if err != nil {
			result.Error = fmt.Sprintf("%s", err)
			bs, _ := json.MarshalIndent(result, "", "\t")
			P("%s\n", string(bs))
			return
		}

		// Log in the newly signed up user.
		tok, err := login(db, signupreq.Username, signupreq.Pwd)
		result.Tok = tok
		if err != nil {
			result.Error = fmt.Sprintf("%s", err)
		}
		bs, _ = json.MarshalIndent(result, "", "\t")
		P("%s\n", string(bs))
	}
}

func edituser(db *sql.DB, username, pwd string, newpwd string) error {
	// Validate existing password
	_, err := login(db, username, pwd)
	if err != nil {
		return err
	}

	// Set new password
	hashedPwd := genHash(newpwd)
	s := "UPDATE user SET password = ? WHERE username = ?"
	_, err = sqlexec(db, s, hashedPwd, username)
	if err != nil {
		return fmt.Errorf("DB error updating user password: %s", err)
	}
	return nil
}
func edituserHandler(db *sql.DB) http.HandlerFunc {
	type EditUserReq struct {
		Username string `json:"username"`
		Pwd      string `json:"pwd"`
		NewPwd   string `json:"newpwd"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Use POST method", 401)
			return
		}

		bs, err := ioutil.ReadAll(r.Body)
		if err != nil {
			handleErr(w, err, "edituserHandler")
			return
		}
		var req EditUserReq
		err = json.Unmarshal(bs, &req)
		if err != nil {
			handleErr(w, err, "edituserHandler")
			return
		}
		if req.Username == "" {
			http.Error(w, "username required", 401)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		P := makeFprintf(w)

		// Attempt to edit user.
		var result LoginResult
		err = edituser(db, req.Username, req.Pwd, req.NewPwd)
		if err != nil {
			result.Error = fmt.Sprintf("%s", err)
			bs, _ := json.MarshalIndent(result, "", "\t")
			P("%s\n", string(bs))
			return
		}

		// Log in the newly edited user.
		tok, err := login(db, req.Username, req.NewPwd)
		result.Tok = tok
		if err != nil {
			result.Error = fmt.Sprintf("%s", err)
		}
		bs, _ = json.MarshalIndent(result, "", "\t")
		P("%s\n", string(bs))
	}
}

func deluser(db *sql.DB, username, pwd string) error {
	// Validate existing password
	_, err := login(db, username, pwd)
	if err != nil {
		return err
	}

	// Delete user
	s := "DELETE FROM user WHERE username = ?"
	_, err = sqlexec(db, s, username)
	if err != nil {
		return fmt.Errorf("DB error deleting user: %s", err)
	}
	return nil
}
func deluserHandler(db *sql.DB) http.HandlerFunc {
	type DelUserReq struct {
		Username string `json:"username"`
		Pwd      string `json:"pwd"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Use POST method", 401)
			return
		}

		bs, err := ioutil.ReadAll(r.Body)
		if err != nil {
			handleErr(w, err, "deluserHandler")
			return
		}
		var req DelUserReq
		err = json.Unmarshal(bs, &req)
		if err != nil {
			handleErr(w, err, "deluserHandler")
			return
		}
		if req.Username == "" {
			http.Error(w, "username required", 401)
			return
		}

		// Attempt to delete user.
		var result LoginResult
		err = deluser(db, req.Username, req.Pwd)
		if err != nil {
			result.Error = fmt.Sprintf("%s", err)
		}

		w.Header().Set("Content-Type", "application/json")
		P := makeFprintf(w)
		bs, _ = json.MarshalIndent(result, "", "\t")
		P("%s\n", string(bs))
	}
}

func saveGrid(db *sql.DB, userid int64, gridjson string) error {
	s := "INSERT OR REPLACE INTO savedgrid (user_id, gridjson) VALUES (?, ?);"
	_, err := sqlexec(db, s, userid, gridjson)
	if err != nil {
		return fmt.Errorf("DB error saving grid: %s", err)
	}
	return nil
}
func savegridHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Use POST method", 401)
			return
		}

		q := r.URL.Query()
		username := q.Get("username")
		tok := q.Get("tok")
		if username == "" {
			http.Error(w, "username required", 401)
			return
		}
		if tok == "" {
			http.Error(w, "tok required", 401)
			return
		}
		bs, err := ioutil.ReadAll(r.Body)
		if err != nil {
			handleErr(w, err, "saveGrid")
			return
		}
		gridjson := string(bs)

		u := findUser(db, username)
		if u == nil {
			http.Error(w, fmt.Sprintf("No user '%s'", username), 401)
			return
		}
		if !validateTok(tok, u) {
			http.Error(w, fmt.Sprintf("Token not validated for '%s' ", u.Username), 401)
			return
		}
		err = saveGrid(db, u.Userid, gridjson)
		if err != nil {
			handleErr(w, err, "saveGrid")
			return
		}
	}
}

func loadGrid(db *sql.DB, userid int64) string {
	s := "SELECT gridjson FROM savedgrid WHERE user_id = ?"
	row := db.QueryRow(s, userid)
	var gridjson string
	err := row.Scan(&gridjson)
	if err == sql.ErrNoRows {
		return ""
	}
	if err != nil {
		fmt.Printf("loadGrid() err: %s\n", err)
		return ""
	}
	return gridjson
}
func loadgridHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("username")
		tok := r.FormValue("tok")
		if username == "" {
			http.Error(w, "username required", 401)
			return
		}
		if tok == "" {
			http.Error(w, "tok required", 401)
			return
		}

		u := findUser(db, username)
		if u == nil {
			http.Error(w, fmt.Sprintf("No user '%s'", username), 401)
			return
		}
		if !validateTok(tok, u) {
			http.Error(w, fmt.Sprintf("Token not validated for '%s' ", u.Username), 401)
			return
		}
		gridjson := loadGrid(db, u.Userid)
		if gridjson == "" {
			gridjson = "[]"
		}

		w.Header().Set("Content-Type", "application/json")
		P := makeFprintf(w)
		P("%s\n", gridjson)
	}
}

func testfeedHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from all sites.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/rss+xml")
		P := makeFprintf(w)

		now := time.Now()
		f := feeds.Feed{
			Title:       "FreeRSS test rss feed",
			Link:        &feeds.Link{Href: "http://freerss.robdelacruz.com/api/testfeed"},
			Description: "Test RSS feed with inline script",
			Author:      &feeds.Author{},
			Created:     now,
		}
		item := feeds.Item{
			Title: "test item with inline script",
			Link:  &feeds.Link{Href: "http://freerss.robdelacruz.com/api/testfeed"},
			Description: `<p>This is <em>markup</em> containing an inline script. 
			<a href="#" onclick="alert('click')">click me</a>
			</p>
			<script defer type="application/javascript">console.log("inline script"); alert("inline script!");</script>`,
			Author:  &feeds.Author{},
			Created: now,
		}
		f.Items = []*feeds.Item{
			&item,
		}
		rss, err := f.ToRss()
		if err != nil {
			handleErr(w, err, "testfeedHandler")
			return
		}
		P("%s\n", rss)
	}
}
