<script>
let cols = [
    ["abc", "def", "ghi"],
    ["jkl", "mno"],
    ["pqr", "zzz"],
];

function onchangeit(e) {
    e.preventDefault();

    col3widgets.push("stu");
    col3widgets = col3widgets;

    col2widgets[0] = "JKL";
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
        return;
    }
    let icol = e.target.getAttribute("data-icol");
    let iwidget = e.target.getAttribute("data-iwidget");
    e.dataTransfer.setData("text/plain", cols[icol][iwidget]);
}
function ondragover(e) {
    if (!targetHasClass(e, "widget", "dropzone")) {
        return;
    }
    e.preventDefault();
    e.dataTransfer.dropEffect = "move";
}
function ondrop(e) {
    if (!targetHasClass(e, "widget", "dropzone")) {
        return;
    }

    e.preventDefault();
    let widgetcontent = e.dataTransfer.getData("text/plain");

    if (targetHasClass(e, "widget")) {
        let icol = e.target.getAttribute("data-icol");
        let iwidget = e.target.getAttribute("data-iwidget");
        cols[icol].splice(iwidget, 0, widgetcontent);
        cols[icol] = cols[icol];
        return;
    }

    // "dropzone" column
    let icol = e.target.getAttribute("data-icol");
    let numcolwidgets = cols[icol].length;
    cols[icol].push(widgetcontent);
    cols[icol] = cols[icol];
}
function ondragend(e) {
    if (!targetHasClass(e, "widget", "dropzone")) {
        return;
    }

    // if drop was completed
    if (e.dataTransfer.dropEffect != "none") {
        let icol = e.target.getAttribute("data-icol");
        let iwidget = e.target.getAttribute("data-iwidget");
        cols[icol].splice(iwidget, 1);
        cols[icol] = cols[icol];
    }
}
</script>

<style>
</style>

<div class="flex flex-row justify-center">
{#each cols as col, icol}
    <div data-icol={icol} class="dropzone w-widget mx-2 pb-32">
    {#each col as w, iwidget}
    <div id="{w}" data-icol={icol} data-iwidget={iwidget} draggable="true" class="widget bg-green-500 mb-2 p-8 w-full">{w}</div>
    {/each}
    </div>
{/each}
</div>

<button on:click={onchangeit} class="bg-gray-500 px-2 py-1 border rounded">change it</button>

<svelte:body
    on:dragstart={ondragstart}
    on:dragover={ondragover}
    on:drop={ondrop}
    on:dragend={ondragend}
/>

