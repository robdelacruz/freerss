<script>
import {onMount} from "svelte";

export let url = "";
export let maxitems = 10;
let svcurl = "http://localhost:8000/api/feed/";

let _feed = null;
let _err = null;
let _showmenu = false;

onMount(function() {
    let sreq = `${svcurl}?url=${encodeURIComponent(url)}&maxitems=${maxitems}`
    fetch(sreq, {method: "GET"})
    .then(res => {
        if (!res.ok) {
            return res.text().then(text => Promise.reject(text));
        }
        return res.json();
    })
    .then(feed => {
        _feed = feed;
        console.log(_feed);
    })
    .catch(err => {
        _err = err;
    });
});

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
    _showmenu = !_showmenu;
}

function onwidgetclick(e) {
    e.preventDefault();
    _showmenu = false;
}
</script>

<div data-icol="0" data-iwidget="0" draggable="true" class="widget w-full" on:click={onwidgetclick}>
{#if _feed}
    <div class="flex flex-row justify-between">
        <h1 class="text-sm font-bold border-b border-gray-500 pb-1 mb-2">{_feed.title}</h1>
        <div class="relative">
            <button class="menubutton h-4 w-4" on:click={onmenu}>
                <img class="" src="cheveron-down.svg" alt="settings">
            </button>
            {#if _showmenu}
            <div class="absolute top-auto right-0 py-1 bg-gray-200 text-gray-800 w-20 border border-gray-500 shadow-xs">
                <a href="#a" class="block leading-none px-2 py-1 hover:bg-gray-400 hover:text-gray-900" role="menuitem">Settings</a>
                <a href="#a" class="block leading-none px-2 py-1 hover:bg-gray-400 hover:text-gray-900" role="menuitem">Delete</a>
            </div>
            {/if}
        </div>
    </div>
    <ul class="linklist">
    {#each _feed.entries as entry}
        <li>
            <p><a href="{entry.url}">{entry.title}</a></p>
        </li>
    {/each}
    </ul>
{:else if _err}
    <p>Error: {_err}</p>
{:else}
    <p>Loading...</p>
{/if}
</div>

