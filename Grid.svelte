<script>
import {onMount} from "svelte";
import RSSView from "./RSSView.svelte";
import LoginForm from "./LoginForm.svelte";
let svcurl = "http://localhost:8000/api";
let cols = [];
let _wid = 0;

let ui = {};
ui.mode = "";

function getHighestWid(cols) {
    let highestwid = 0;
    for (let icol=0; icol < cols.length; icol++) {
        for (let irow=0; irow < cols[icol].length; irow++) {
            if (cols[icol][irow].wid > highestwid) {
                highestwid = cols[icol][irow].wid;
            }
        }
    }
    return highestwid;
}

onMount(function() {
    loadCols().then(resultcols => {
        if (resultcols == null) {
            return;
        }
        cols = resultcols;
    });
    console.log(cols);
    _wid = getHighestWid(cols);
});

function rssview_updated(e) {
    saveCols(cols);
}

function rssview_deleted(e) {
    removeWidget(cols, e.detail.wid);
    cols = cols;
    saveCols(cols);
}

function currentSession() {
    let cookies = document.cookie.split(";");
    for (let i=0; i < cookies.length; i++) {
        let cookie = cookies[i].trim();
        let [k,v] = cookie.split("=");
        if (k != "usernametok") {
            continue;
        }
        if (v == undefined) {
            v = "";
        }
        let [username, tok] = v.split("|");
        if (tok == undefined) {
            tok = "";
        }
        return {username: username, tok: tok};
    }
    return null;
}

async function loadGrid(username, tok) {
    let sreq = `${svcurl}/loadgrid/?username=${username}&tok=${tok}`;
    try {
        let res = await fetch(sreq, {
            method: "GET",
        });
        if (!res.ok) {
            let err = await res.text();
            console.error(err);
            return null;
        }
        let result = await res.json();
        return result;
    } catch(err) {
        console.error(err);
        return null;
    }
}
async function loadCols() {
    let session = currentSession();
    if (session == null) {
        // Restore from localStorage if present.
        let jsoncols = localStorage.getItem("cols");
        if (jsoncols != null) {
            return JSON.parse(jsoncols);
        }

        // Default widgets, if first time page was accessed.
        let nitems = 5;
        let initcols = [
            [
                newWidget("http://rss.slashdot.org/Slashdot/slashdotMain", 8),
                newWidget("https://news.ycombinator.com/rss", 10),
            ],
            [
                newWidget("https://www.lewrockwell.com/feed/", 8),
                newWidget("https://feeds.feedburner.com/zerohedge/feed", 8),
            ],
            [
                newWidget("", nitems),
            ],
        ];
        return initcols;
    }

    let sessioncols = await loadGrid(session.username, session.tok);
    return sessioncols;
}
async function saveGrid(username, tok, cols) {
    let sreq = `${svcurl}/savegrid/?username=${username}&tok=${tok}`;
    try {
        let res = await fetch(sreq, {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify(cols),
        });
        if (!res.ok) {
            let err = await res.text();
            console.error(err);
            return false;
        }
        return true;
    } catch(err) {
        console.error(err);
        return false;
    }
}
async function saveCols(cols) {
    let session = currentSession();
    if (session == null) {
        localStorage.setItem("cols", JSON.stringify(cols));
        return;
    }

    let wasSaved = await saveGrid(session.username, session.tok, cols);
    if (!wasSaved) {
        console.error("Error saving grid");
        return;
    }
}

function newWidget(feedurl, maxitems) {
    _wid++;

    return {
        wid: _wid,
        feedurl: feedurl,
        maxitems: maxitems,
    };
}

function getAttrWid(el) {
    return el.getAttribute("data-wid");
}
function findWidgetLocFromWid(cols, wid) {
    for (let icol=0; icol < cols.length; icol++) {
        for (let irow=0; irow < cols[icol].length; irow++) {
            let w = cols[icol][irow];
            if (w.wid == wid) {
                return {col: icol, row: irow};
            }
        }
    }
    return null;
}
function findWidgetLoc(cols, el) {
    let wid = getAttrWid(el);
    if (!wid) {
        return null;
    }
    return findWidgetLocFromWid(cols, wid);
}
function removeWidget(cols, wid) {
    for (let icol=0; icol < cols.length; icol++) {
        for (let irow=0; irow < cols[icol].length; irow++) {
            let w = cols[icol][irow];
            if (w.wid == wid) {
                cols[icol].splice(irow, 1);
                cols[icol] = cols[icol];
                return;
            }
        }
    }
    return;
}

function targetHasClass(e, ...cc) {
    for (let i=0; i < cc.length; i++) {
        let c = cc[i];
        if (e.target.classList.contains(c)) {
            return true;
        }
    }
    return false;
}
function ondragstart(e) {
    if (!targetHasClass(e, "widget")) {
        e.preventDefault();
        return;
    }
    let l = findWidgetLoc(cols, e.target);
    if (l == null) {
        e.preventDefault();
        return;
    }
    let jsonwidget = JSON.stringify(cols[l.col][l.row]);
    e.dataTransfer.setData("text/plain", jsonwidget);
}
function ondragenter(e) {
    let target = e.target.closest(".widget, .dropzone");
    if (target == null) {
        return;
    }
    e.preventDefault();
}
function ondragover(e) {
    let target = e.target.closest(".widget, .dropzone");
    if (target == null) {
        return;
    }
    e.preventDefault();
    e.dataTransfer.dropEffect = "move";
}
function ondrop(e) {
    let target = e.target.closest(".widget, .dropzone");
    if (target == null) {
        return;
    }

    e.preventDefault();
    let jsonwidget = e.dataTransfer.getData("text/plain");
    let widget = JSON.parse(jsonwidget);

    if (target.classList.contains("widget")) {
        // Don't do anything if widget was dragged to itself.
        if (getAttrWid(target) == widget.wid) {
            return;
        }

        removeWidget(cols, widget.wid);

        let l = findWidgetLoc(cols, target);
        if (l == null) {
            return;
        }
        cols[l.col].splice(l.row, 0, widget);
        cols[l.col] = cols[l.col];
    } else {
        // "dropzone" column
        removeWidget(cols, widget.wid);

        let icol = target.getAttribute("data-icol");
        cols[icol].push(widget);
        cols[icol] = cols[icol];
    }

    saveCols(cols);
}
function ondragend(e) {
    if (!targetHasClass(e, "widget")) {
        return;
    }

    // if drop was completed
//    if (e.dataTransfer.dropEffect != "none") {
//        let l = findWidgetLoc(cols, e.target);
//        if (l == null) {
//            return;
//        }
//        cols[l.col].splice(l.row, 1);
//        cols[l.col] = cols[l.col];
//    }
}

// Add empty widget to the upper leftmost corner.
function onaddwidget(e) {
    let ncolstoadd = 3 - cols.length;
    for (let i=0; i < ncolstoadd; i++) {
        cols.push([]);
    }
    cols[0].splice(0, 0, newWidget("", 0));
    cols[0] = cols[0];
}

function onlogin(e) {
    ui.mode = "login";
}
function loginform_login(e) {
    let username = e.detail.username;
    let tok = e.detail.tok;
    document.cookie = `usernametok=${username}|${tok};path=/`;
    ui.mode = "";

    loadCols().then(resultcols => {
        if (resultcols == null) {
            return;
        }
        cols = resultcols;
    });
}
function loginform_cancel(e) {
    console.log(`loginform_cancel()`);
    ui.mode = "";
}

</script>

<div class="flex flex-row justify-between border-b border-gray-500 text-gray-200 pb-1 mb-2">
    <div>
        <h1 class="inline self-end text-sm ml-1 mr-2">FreeRSS</h1>
        <a href="about.html" class="self-end mr-2">About</a>
    </div>
    <div>
        <a href="#a" class="text-xs bg-gray-400 text-gray-800 self-center rounded px-2 mr-2" on:click={onaddwidget}>Add Widget</a>
        <a href="#a" class="self-end mr-1" on:click={onlogin}>Login</a>
    </div>
</div>

{#if ui.mode == ""}
    <div class="flex flex-row justify-center">
    {#each cols as col, icol}
        <div data-icol={icol} class="dropzone w-widget mx-2 pb-32">
        {#each cols[icol] as w, irow (w.wid)}
        <RSSView bind:wid={cols[icol][irow].wid} bind:feedurl={cols[icol][irow].feedurl} bind:maxitems={cols[icol][irow].maxitems} on:updated={rssview_updated} on:deleted={rssview_deleted} />
        {/each}
        </div>
    {/each}
    </div>
{:else if ui.mode == "login"}
    <div class="flex flex-row w-full h-screen justify-center items-center">
        <div class="widget">
            <LoginForm username="" pwd="" on:login={loginform_login} on:cancel={loginform_cancel} />
        </div>
    </div>
{/if}

<svelte:body
    on:dragstart={ondragstart}
    on:dragenter={ondragenter}
    on:dragover={ondragover}
    on:drop={ondrop}
    on:dragend={ondragend}
/>

