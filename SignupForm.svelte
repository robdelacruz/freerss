<script>
import {onMount, createEventDispatcher} from "svelte";
let dispatch = createEventDispatcher();
let svcurl = "http://localhost:8000/api";

export let username = "";
let pwd = "";
let pwd2 = "";
let frm = {};
frm.mode = "";
frm.status = "";

// Post signup username/pwd and async returns loginresult:
// {tok: "", error: ""}
async function signup(username, pwd) {
    let sreq = `${svcurl}/signup/`;
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

async function onsignup(e) {
    e.preventDefault();
    if (pwd != pwd2) {
        frm.mode = "";
        frm.status = "re-entered password doesn't match";
        return;
    }

    frm.mode = "loading";
    frm.status = "Creating account...";

    let result = await signup(username, pwd);
    frm.mode = "";
    frm.status = "";
    if (result.error != "") {
        frm.status = result.error;
        return;
    }
    dispatch("signup", {username: username, tok: result.tok});
}
function oncancel(e) {
    username = "";
    pwd = "";
    dispatch("cancel");
}
</script>

<form class="bg-gray-800 text-gray-200 p-2">
    <div class="mb-2">
        <label class="block bg-gray-800 text-gray-200" for="username">username</label>
        <input class="block border border-gray-500 bg-gray-200 text-gray-800 py-0 px-2 w-full" id="username" name="username" type="text" bind:value={username}>
    </div>
    <div class="mb-4">
        <label class="block bg-gray-800 text-gray-200" for="pwd">password</label>
        <input class="block border border-gray-500 bg-gray-200 text-gray-800 py-0 px-2 w-full" id="pwd" name="pwd" type="password" bind:value={pwd}>
    </div>
    <div class="mb-4">
        <label class="block bg-gray-800 text-gray-200" for="pwd2">re-enter password</label>
        <input class="block border border-gray-500 bg-gray-200 text-gray-800 py-0 px-2 w-full" id="pwd2" name="pwd2" type="password" bind:value={pwd2}>
    </div>
{#if frm.status != ""}
    <div class="mb-2">
        <p class="font-bold bg-gray-800 text-gray-200">{frm.status}</p>
    </div>
{/if}
    <div class="flex flex-row justify-center mb-4">
        <div>
{#if frm.mode == "loading"}
            <button disabled on:click={onsignup} class="inline mx-auto py-1 px-2 bg-gray-200 text-gray-800 mr-2">Sign Up</button>
{:else}
            <button on:click={onsignup} class="inline mx-auto py-1 px-2 bg-gray-200 text-gray-800 mr-2">Sign Up</button>
{/if}
            <button on:click={oncancel} class="inline mx-auto py-1 px-2 bg-gray-200 text-gray-800">Cancel</button>
        </div>
    </div>
</form>

