<div data-wid={wid} draggable="true" class="widget w-full" on:click={onwidgetclick}>
    <div class="flex flex-row justify-between w-full border-b border-gray-500 pb-1 mb-2">
        <h1 class="text-sm font-bold">
            {#if ui.feed}
                {#if ui.feed.url != ""}
                    <a href="{ui.feed.url}" class="" target="_blank">{ui.feed.title}</a>
                {:else}
                    <a href="#a" class="" target="_blank">{ui.feed.title}</a>
                {/if}
            {:else}
                Select Feed
            {/if}
        </h1>
        <div class="relative">
            <button class="menubutton h-4 w-4" on:click={onrefresh}>
                <img class="" src="refresh.svg" alt="refresh">
            </button>
            <button class="menubutton h-4 w-4" on:click={onmenu}>
                <img class="" src="cheveron-down.svg" alt="settings">
            </button>
            {#if ui.showmenu}
            <div class="absolute top-auto right-0 py-1 bg-gray-200 text-gray-800 w-20 border border-gray-500 shadow-xs">
                <a href="#a" class="block leading-none px-2 py-1 hover:bg-gray-400 hover:text-gray-900" role="menuitem" on:click={onsettings}>Settings</a>
                <a href="#a" class="block leading-none px-2 py-1 hover:bg-gray-400 hover:text-gray-900" role="menuitem" on:click={ondelete}>Delete</a>
            </div>
            {/if}
        </div>
    </div>

{#if ui.mode == "loading"}
    <p>Loading...</p>
{:else if ui.mode == "display"}
    {#if ui.err}
        <p>Error ({ui.err})</p>
    {:else if ui.feed}
        <ul class="linklist">
        {#each ui.feed.entries as entry}
            <li>
            {#if preview}
                <div class="mb-3">
                    <a class="block text-gray-200 text-xs mb-1" href="{entry.url}" target="_blank">{entry.title}</a>
                    <div class="content text-gray-500">
                        {@html entry.desc}
                    </div>
                </div>
            {:else}
                <a class="block" href="{entry.url}" target="_blank">{entry.title}</a>
            {/if}
            </li>
        {/each}
        </ul>
    {/if}
{:else if ui.mode == "settings"}
    <form class="">
        <div class="mb-2">
            <label class="block" for="feedurl">Website/Feed url</label>
            <input class="block border border-gray-500 bg-gray-200 text-gray-800 py-0 px-2 w-full" id="feedurl" name="feedurl" type="text" bind:value={settingsform.feedurl}>
        </div>
        <div class="mb-2">
            <label class="block" for="maxitems"># links to display</label>
            <input class="block border border-gray-500 py-0 px-2 bg-gray-200 text-gray-800 w-10" id="maxitems" name="maxitems" maxlength="2" type="text" bind:value={settingsform.maxitems}>
        </div>
        <div class="mb-2">
            <label class="inline-flex items-center" for="preview">
                <input class="mr-1" id="preview" name="preview" type="checkbox" bind:checked={settingsform.preview}>
                <span class="">show preview</span>
            </label>
        </div>
    {#if settingsform.status != ""}
        <div class="mb-2">
            <p class="font-bold">{settingsform.status}</p>
        </div>
    {/if}
        <div class="flex flex-row justify-center">
            <div>
    {#if settingsform.mode == "loading"}
                <button disabled on:click={onformupdate} class="inline mx-auto py-1 px-2 bg-gray-200 text-gray-800 mr-2">Update</button>
    {:else}
                <button on:click={onformupdate} class="inline mx-auto py-1 px-2 bg-gray-200 text-gray-800 mr-2">Update</button>
    {/if}
                <button on:click={onformcancel} class="inline mx-auto py-1 px-2 bg-gray-200 text-gray-800">Cancel</button>
            </div>
        </div>
    </form>
{:else if ui.mode == "delete"}
    <p>delete</p>
{/if}
</div>

<script>
import {onMount, createEventDispatcher} from "svelte";
let dispatch = createEventDispatcher();
let svcurl = "http://localhost:8000/api";

export let feedurl = "";
export let maxitems = 10;
export let preview = true;
export let wid = 0;
let container;

let ui = {};
ui.feed = null;
ui.err = null;
ui.showmenu = false;

if (feedurl != "") {
    ui.mode = "loading";
} else {
    ui.mode = "settings";
}

let settingsform = {};
settingsform.mode = "";
settingsform.status = "";
settingsform.feedurl = feedurl;
settingsform.maxitems = maxitems;
settingsform.preview = preview;

onMount(function() {
    reloadDisplay();
});

function reloadDisplay() {
    if (feedurl == "") {
        ui.mode = "settings";
        return;
    }

    let qmaxitems = maxitems;
    if (qmaxitems == 0) {
        qmaxitems = 5;
    }
    let sreq = `${svcurl}/feed?url=${encodeURIComponent(feedurl)}&maxitems=${qmaxitems}`
    fetch(sreq, {method: "GET"})
    .then(res => {
        if (!res.ok) {
            return res.text().then(text => Promise.reject(text));
        }
        return res.json();
    })
    .then(feed => {
        ui.feed = feed;
        ui.mode = "display";
        ui.err = null;
    })
    .catch(err => {
        ui.err = err;
        ui.mode = "display";
        ui.feed = null;
    });
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

function onrefresh(e) {
    e.preventDefault();
    e.stopPropagation();
    ui.showmenu = false;
    reloadDisplay();
}
function onmenu(e) {
    e.preventDefault();
    e.stopPropagation();
    ui.showmenu = !ui.showmenu;
}
function onwidgetclick(e) {
    ui.showmenu = false;
}
function onsettings(e) {
    e.preventDefault();
    ui.mode = "settings";
    settingsform.status = "";
    settingsform.feedurl = feedurl;
    settingsform.maxitems = maxitems;
    settingsform.preview = preview;
}
function ondelete(e) {
    e.preventDefault();
    ui.mode = "delete";

    let widget = {
        wid: wid,
        feedurl: feedurl,
        maxitems: maxitems,
        preview: preview,
    };
    dispatch("deleted", widget);
}

async function onformupdate(e) {
    e.preventDefault();
    settingsform.status = "";

    if (settingsform.feedurl == "") {
        ui.mode = "display";
        return;
    }

    settingsform.feedurl = completeurl(settingsform.feedurl);
    if (settingsform.feedurl != feedurl) {
        try {
            settingsform.mode = "loading";
            settingsform.status = "Finding feeds...";
            let feeds = await discoverfeeds(settingsform.feedurl);
            settingsform.mode = "";
            settingsform.status = "";

            if (feeds.length == 0) {
                settingsform.status = "No feed found.";
                return;
            }

            feedurl = feeds[0];
            settingsform.feedurl = feedurl;
        } catch(err) {
            console.log(err);
            settingsform.mode = "";
            settingsform.status = "server error: try again later";
            return;
        }
    }

    maxitems = settingsform.maxitems;
    preview = settingsform.preview;
    reloadDisplay();

    let widget = {
        wid: wid,
        feedurl: feedurl,
        maxitems: maxitems,
        preview: preview,
    };
    dispatch("updated", widget);
}
function onformcancel(e) {
    e.preventDefault();

    // restore previous settings
    settingsform.feedurl = feedurl;
    settingsform.maxitems = maxitems;
    settingsform.preview = preview;
    ui.mode = "display";
}

async function discoverfeeds(qurl) {
    let sreq = `${svcurl}/discoverfeed?url=${encodeURIComponent(qurl)}`
    let res = await fetch(sreq, {method: "GET"});
    if (!res.ok) {
        let err = await res.text();
        return Promise.reject(err);
    }
    let feeds = await res.json();
    return feeds;
}

let _suggestions = [
    "http://www.everything2.org",
    "http://www.slashdot.org",
    "http://news.ycombinator.com",
    "http://www.lwn.net",
    "http://www.lewrockwell.com",
    "http://www.zerohedge.com",
    "http://www.naturalnews.com",
];
function suggesturl() {
    let s = _suggestions[Math.floor(Math.random()*_suggestions.length)];
    return s;
}

function completeurl(surl) {
    surl = surl.trim();
    if (!surl.startsWith("http://") && !surl.startsWith("https://")) {
        return `http://${surl}`;
    }
    if (surl.startsWith("//")) {
        return `http:${surl}`;
    }
    return surl;
}
</script>

