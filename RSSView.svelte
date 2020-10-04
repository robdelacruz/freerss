<script>
import {onMount} from "svelte";

export let feedurl = "";
export let maxitems = 10;
let svcurl = "http://localhost:8000/api/feed/";

let ui = {};
ui.feed = null;
ui.err = null;
ui.showmenu = false;
ui.mode = "loading";

let settingsform = {};
settingsform.feedurl = feedurl;
settingsform.maxitems = maxitems;

onMount(function() {
    reloadDisplay();
});

function reloadDisplay() {
    ui.mode = "loading";

    let sreq = `${svcurl}?url=${encodeURIComponent(feedurl)}&maxitems=${maxitems}`
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

function onmenu(e) {
    e.preventDefault();
    e.stopPropagation();
    ui.showmenu = !ui.showmenu;
}
function onwidgetclick(e) {
    ui.showmenu = false;
}
function onsettings(e) {
    ui.mode = "settings";
}
function ondelete(e) {
    ui.mode = "delete";
}

function onformupdate(e) {
    e.preventDefault();

    feedurl = settingsform.feedurl;
    maxitems = settingsform.maxitems;
    reloadDisplay();
}
function onformcancel(e) {
    // restore previous settings
    settingsform.feedurl = feedurl;
    settingsform.maxitems = maxitems;
    ui.mode = "display";
}
</script>

<div data-icol="0" data-iwidget="0" draggable="true" class="widget w-full" on:click={onwidgetclick}>
    <div class="flex flex-row justify-between">
        <h1 class="text-sm font-bold border-b border-gray-500 pb-1 mb-2">
            {#if ui.feed}
                {ui.feed.title}
            {:else}
                Select Feed
            {/if}
        </h1>
        <div class="relative">
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
                <a class="block" href="{entry.url}">{entry.title}</a>
            </li>
        {/each}
        </ul>
    {/if}
{:else if ui.mode == "settings"}
    <form class="">
        <div class="mb-2">
            <label class="block" for="feedurl">feed url</label>
            <input class="block border border-gray-500 bg-gray-200 text-gray-800 py-1 px-2" id="feedurl" name="feedurl" size="50" type="text" bind:value={settingsform.feedurl}>
        </div>
        <div class="mb-2">
            <label class="block" for="maxitems">max items</label>
            <input class="block border border-gray-500 py-1 px-2 bg-gray-200 text-gray-800" id="maxitems" name="maxitems" size="3" maxlength="3" type="text" bind:value={settingsform.maxitems}>
        </div>
        <div class="">
            <button on:click={onformupdate} class="inline mx-auto py-1 px-2 bg-gray-200 text-gray-800 mr-4">Update</button>
            <button on:click={onformcancel} class="inline mx-auto py-1 px-2 bg-gray-200 text-gray-800">Cancel</button>
        </div>
    </form>
{:else if ui.mode == "delete"}
    <p>delete</p>
{/if}
</div>

