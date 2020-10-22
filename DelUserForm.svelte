<form class="bg-gray-800 text-gray-200 p-2">
    <h1 class="text-sm font-bold mb-2">Delete User</h1>
    <div class="mb-2">
        <label class="block bg-gray-800 text-gray-200" for="username">username</label>
        <input disabled class="block border border-gray-500 bg-gray-500 text-gray-800 py-0 px-2 w-full" id="username" name="username" type="text" bind:value={username}>
    </div>
    <div class="mb-4">
        <label class="block bg-gray-800 text-gray-200" for="pwd">current password</label>
        <input class="block border border-gray-500 bg-gray-200 text-gray-800 py-0 px-2 w-full" id="pwd" name="pwd" type="password" bind:value={pwd}>
    </div>
{#if frm.status != ""}
    <div class="mb-2">
        <p class="font-bold bg-gray-800 text-gray-200">{frm.status}</p>
    </div>
{/if}
    <div class="flex flex-row justify-center mb-4">
        <div>
{#if frm.mode == "loading"}
            <button disabled on:click={ondelete} class="inline mx-auto py-1 px-2 bg-gray-200 text-gray-800 mr-2">Delete User</button>
{:else}
            <button on:click={ondelete} class="inline mx-auto py-1 px-2 bg-gray-200 text-gray-800 mr-2">Delete User</button>
{/if}
            <button on:click={oncancel} class="inline mx-auto py-1 px-2 bg-gray-200 text-gray-800">Cancel</button>
        </div>
    </div>
</form>

<script>
import {createEventDispatcher} from "svelte";
let dispatch = createEventDispatcher();
let svcurl = "http://localhost:8000/api";

export let username = "";
let pwd = "";
let newpwd = "";
let newpwd2 = "";
let frm = {};
frm.mode = "";
frm.status = "";

// Post del username/pwd and async returns error text if any:
// {tok: "", error: ""}
async function deluser(username, pwd) {
    let sreq = `${svcurl}/deluser/`;
    let reqbody = {
        username: username,
        pwd: pwd,
    };
    try {
        let res = await fetch(sreq, {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify(reqbody),
        });
        if (!res.ok) {
            let err = await res.text();
            return {tok: "", error: err};
        }
        let result = await res.json();
        return result;
    } catch(err) {
        console.error(err);
        return {tok: "", error: "server/network error"};
    }
}

async function ondelete(e) {
    e.preventDefault();

    frm.mode = "loading";
    frm.status = "Deleting user...";

    let result = await deluser(username, pwd);
    frm.mode = "";
    frm.status = "";
    if (result.error != "") {
        frm.status = result.error;
        return;
    }
    dispatch("del");
}
function oncancel(e) {
    e.preventDefault();
    username = "";
    pwd = "";
    dispatch("cancel");
}
</script>

