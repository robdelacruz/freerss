<script>
import RSSView from "./RSSView.svelte";
import {onMount, createEventDispatcher} from "svelte";
let dispatch = createEventDispatcher();
let svcurl = "/api";
export let username = "";
export let tok = "";

let ui = {};
ui.mode = "";
ui.cols = [];

ui.mode = "loading";
$: {
    loadCols(username, tok).then(resultcols => {
        ui.mode = "";
        if (resultcols == null) {
            return;
        }
        ui.cols = resultcols;
    });
}

function rssview_updated(e) {
    saveCols(ui.cols);
}

function rssview_deleted(e) {
    removeWidget(ui.cols, e.detail.wid);
    ui.cols = ui.cols;
    saveCols(ui.cols);
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

            if (res.status >= 400 && res.status < 500) {
                dispatch("invalidtok");
            }
            return null;
        }
        let result = await res.json();
        return result;
    } catch(err) {
        console.error(err);
        return null;
    }
}
async function loadCols(username, tok) {
    if (username == "") {
        // Restore from localStorage if present.
        let jsoncols = localStorage.getItem("cols");
        if (jsoncols != null) {
            return JSON.parse(jsoncols);
        }
        // Return default widgets, if first time page was accessed.
        return defaultCols();
    }

    let sessioncols = await loadGrid(username, tok);
    return sessioncols;
}
function defaultCols() {
    let w1 = newWidget("http://www.engadget.com/rss.xml", 5, true);
    w1.wid = 1;

    let w2 = newWidget("http://rss.slashdot.org/Slashdot/slashdotMain", 5, false);
    w2.wid = 2;

    let w3 = newWidget("https://news.ycombinator.com/rss", 10, false);
    w3.wid = 3;

    let w4 = newWidget("http://fortune2.robdelacruz.com/fortune/?outputfmt=rss&sw=ec", 5, false);
    w4.wid = 4;

    let w5 = newWidget("https://www.lewrockwell.com/feed/", 5, false);
    w5.wid = 5;

    let w6 = newWidget("https://feeds.feedburner.com/zerohedge/feed", 5, false);
    w6.wid = 6;

    return [[w1], [w2, w3], [w4, w5, w6]];
}
function adddefaultwidgets(e) {
    if (ui.cols.length > 0) {
        return;
    }
    ui.cols = defaultCols();
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

            if (res.status >= 400 && res.status < 500) {
                dispatch("invalidtok");
            }
            return false;
        }
        return true;
    } catch(err) {
        console.error(err);
        return false;
    }
}
async function saveCols(cols) {
    if (username == "") {
        localStorage.setItem("cols", JSON.stringify(cols));
        return;
    }

    let wasSaved = await saveGrid(username, tok, cols);
    if (!wasSaved) {
        console.error("Error saving grid");
        return;
    }
}

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
function newWidget(feedurl, maxitems, preview) {
    return {
        wid: getHighestWid(ui.cols) + 1,
        feedurl: feedurl,
        maxitems: maxitems,
        preview: preview,
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
    let l = findWidgetLoc(ui.cols, e.target);
    if (l == null) {
        e.preventDefault();
        return;
    }
    let jsonwidget = JSON.stringify(ui.cols[l.col][l.row]);
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

        removeWidget(ui.cols, widget.wid);

        let l = findWidgetLoc(ui.cols, target);
        if (l == null) {
            return;
        }
        ui.cols[l.col].splice(l.row, 0, widget);
        ui.cols[l.col] = ui.cols[l.col];
    } else {
        // "dropzone" column
        removeWidget(ui.cols, widget.wid);

        let icol = target.getAttribute("data-icol");
        ui.cols[icol].push(widget);
        ui.cols[icol] = ui.cols[icol];
    }

    saveCols(ui.cols);
}
function ondragend(e) {
    if (!targetHasClass(e, "widget")) {
        return;
    }

    // if drop was completed
//    if (e.dataTransfer.dropEffect != "none") {
//        let l = findWidgetLoc(ui.cols, e.target);
//        if (l == null) {
//            return;
//        }
//        ui.cols[l.col].splice(l.row, 1);
//        ui.cols[l.col] = ui.cols[l.col];
//    }
}

// Add empty widget to the upper leftmost corner.
export function addwidget() {
    let ncolstoadd = 3 - ui.cols.length;
    for (let i=0; i < ncolstoadd; i++) {
        ui.cols.push([]);
    }
    ui.cols[0].splice(0, 0, newWidget("", 5, false));
    ui.cols[0] = ui.cols[0];
}


</script>

<div class="flex flex-row justify-center">
{#if ui.mode == ""}
    {#each ui.cols as col, icol}
    <div data-icol={icol} class="dropzone w-widget mx-2 pb-32">
        {#each ui.cols[icol] as w, irow (w.wid)}
        <RSSView bind:wid={ui.cols[icol][irow].wid} bind:feedurl={ui.cols[icol][irow].feedurl} bind:maxitems={ui.cols[icol][irow].maxitems} bind:preview={ui.cols[icol][irow].preview} on:updated={rssview_updated} on:deleted={rssview_deleted} />
        {/each}
    </div>
    {:else}
    <p class="py-1 px-2 bg-gray-200 text-gray-800">You don't have any widgets yet.<br> 
        <a on:click={addwidget} href="#a" class="underline">Add a new widget</a> or 
        <a on:click={adddefaultwidgets} href="#a" class="underline">Add sample widgets</a> 
    </p>
    {/each}
{:else if ui.mode == "loading"}
    <p class="font-bold py-1 px-2 bg-gray-200 text-gray-800">Loading...</p>
{/if}
</div>

<svelte:body
    on:dragstart={ondragstart}
    on:dragenter={ondragenter}
    on:dragover={ondragover}
    on:drop={ondrop}
    on:dragend={ondragend}
/>

