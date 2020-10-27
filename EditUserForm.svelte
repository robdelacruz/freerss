<form class="bg-gray-800 text-gray-200 p-2">
    <div class="mb-2">
        <label class="block bg-gray-800 text-gray-200" for="username">username</label>
        <input disabled class="block border border-gray-500 bg-gray-500 text-gray-800 py-0 px-2 w-full" id="username" name="username" type="text" bind:value={username}>
    </div>
    <div class="mb-4">
        <label class="block bg-gray-800 text-gray-200" for="pwd">current password</label>
        <input class="block border border-gray-500 bg-gray-200 text-gray-800 py-0 px-2 w-full" id="pwd" name="pwd" type="password" bind:value={pwd}>
    </div>
    <div class="mb-4">
        <label class="block bg-gray-800 text-gray-200" for="pwd2">new password</label>
        <input class="block border border-gray-500 bg-gray-200 text-gray-800 py-0 px-2 w-full" id="newpwd" name="newpwd" type="password" bind:value={newpwd}>
    </div>
    <div class="mb-4">
        <label class="block bg-gray-800 text-gray-200" for="pwd2">re-enter new password</label>
        <input class="block border border-gray-500 bg-gray-200 text-gray-800 py-0 px-2 w-full" id="newpwd2" name="newpwd2" type="password" bind:value={newpwd2}>
    </div>
{#if frm.status != ""}
    <div class="mb-2">
        <p class="font-bold bg-gray-800 text-gray-200">{frm.status}</p>
    </div>
{/if}
    <div class="flex flex-row justify-center mb-4">
        <div>
{#if frm.mode == "loading"}
            <button disabled on:click={onupdate} class="inline mx-auto py-1 px-2 bg-gray-200 text-gray-800 mr-2">Update</button>
{:else}
            <button on:click={onupdate} class="inline mx-auto py-1 px-2 bg-gray-200 text-gray-800 mr-2">Update</button>
{/if}
            <button on:click={oncancel} class="inline mx-auto py-1 px-2 bg-gray-200 text-gray-800">Cancel</button>
        </div>
    </div>
</form>

<script>
import {createEventDispatcher} from "svelte";
let dispatch = createEventDispatcher();
let svcurl = "/api";

export let username = "";
let pwd = "";
let newpwd = "";
let newpwd2 = "";
let frm = {};
frm.mode = "";
frm.status = "";

// Post edit username/pwd and async returns loginresult:
// {tok: "", error: ""}
async function edituser(username, pwd, newpwd) {
    let sreq = `${svcurl}/edituser/`;
    let reqbody = {
        username: username,
        pwd: pwd,
        newpwd: newpwd,
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

async function onupdate(e) {
    e.preventDefault();
    if (newpwd != newpwd2) {
        frm.mode = "";
        frm.status = "re-entered password doesn't match";
        return;
    }

    frm.mode = "loading";
    frm.status = "Updating account...";

    let result = await edituser(username, pwd, newpwd);
    frm.mode = "";
    frm.status = "";
    if (result.error != "") {
        frm.status = result.error;
        return;
    }
    dispatch("update", {username: username, tok: result.tok});
}
function oncancel(e) {
    e.preventDefault();
    username = "";
    pwd = "";
    newpwd = "";
    newpwd2 = "";
    dispatch("cancel");
}
</script>

