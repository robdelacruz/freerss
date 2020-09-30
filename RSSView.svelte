<script>
import {onMount} from "svelte";

export let url = "";
export let maxitems = 10;
let svcurl = "http://localhost:8000/api/feed/";

let _feed = null;
let _err = null;

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
</script>

<div data-icol="0" data-iwidget="0" draggable="true" class="widget w-full">
{#if _feed}
    <h1 class="text-sm font-bold border-b border-gray-500 pb-1 mb-2">{_feed.title}</h1>
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

