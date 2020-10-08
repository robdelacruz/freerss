<script>
import {onMount} from "svelte";
import RSSView from "./RSSView.svelte";

let cols = [];

onMount(function() {
    cols = loadCols();
    console.log("onMount");
    console.log(cols);
});

function loadCols() {
    // Restore from localStorage if present.
    let jsoncols = localStorage.getItem("cols");
    if (jsoncols != null) {
        console.log("loadCols");
        console.log(jsoncols);
        return JSON.parse(jsoncols);
    }

    // Default widgets, if first time page was accessed.
    let nitems = 5;
    let initcols = [
        [
            newWidget("http://rss.slashdot.org/Slashdot/slashdotMain", nitems),
            newWidget("https://www.lewrockwell.com/feed/", nitems),
            newWidget("https://feeds.feedburner.com/zerohedge/feed", nitems),
        ],
        [
            newWidget("https://news.ycombinator.com/rss", nitems),
            newWidget("http://feeds.twit.tv/twit.xml", nitems),
        ],
        [
            newWidget("https://feeds.feedburner.com/breitbart", nitems),
            newWidget("", nitems),
        ],
    ];
    return initcols;
}
function saveCols(cols) {
    console.log("saveCols");
    console.log(cols);
    localStorage.setItem("cols", JSON.stringify(cols));
    console.log("saveCols");
    console.log(cols);
}

let _wid = 0;
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
                return cols;
            }
        }
    }
    return cols;
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
    removeWidget(cols, widget.wid);

    if (target.classList.contains("widget")) {
        let l = findWidgetLoc(cols, target);
        if (l == null) {
            return;
        }
        cols[l.col].splice(l.row, 0, widget);
        cols[l.col] = cols[l.col];
        console.log("ondrop -> widget");
        console.log(`l.col: ${l.col}`);
        console.log(`length: ${cols[l.col].length}`);
        for (let i=0; i < cols[l.col].length; i++) {
            console.log(cols[l.col][i]);
        }
    } else {
        // "dropzone" column
        let icol = target.getAttribute("data-icol");
        cols[icol].push(widget);
        cols[icol] = cols[icol];

        console.log("ondrop -> dropzone");
        console.log(`icol: ${icol}`);
        console.log(`length: ${cols[icol].length}`);
        for (let i=0; i < cols[icol].length; i++) {
            console.log(cols[icol][i]);
        }
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
</script>

<style>
</style>

<div class="flex flex-row justify-center">
{#each cols as col, icol}
    <div data-icol={icol} class="dropzone w-widget mx-2 pb-32">
    {#each cols[icol] as w, irow (w.wid)}
        <RSSView bind:wid={cols[icol][irow].wid} bind:feedurl={cols[icol][irow].feedurl} bind:maxitems={cols[icol][irow].maxitems} />
    {/each}
    </div>
{/each}
</div>

<svelte:body
    on:dragstart={ondragstart}
    on:dragenter={ondragenter}
    on:dragover={ondragover}
    on:drop={ondrop}
    on:dragend={ondragend}
/>

