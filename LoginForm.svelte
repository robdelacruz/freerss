<script>
import {onMount, createEventDispatcher} from "svelte";
let dispatch = createEventDispatcher();
let svcurl = "http://localhost:8000/api";

export let username = "";
export let pwd = "";
let frm = {};
frm.mode = "";
frm.status = "";

// Post login username/pwd and async returns loginresult:
// {tok: "", error: ""}
async function login(username, pwd) {
    let sreq = `${svcurl}/login/`;
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

async function onlogin(e) {
    frm.mode = "loading";
    frm.status = "Loging in...";

    let result = await login(username, pwd);
    frm.mode = "";
    frm.status = "";
    if (result.error != "") {
        frm.status = result.error;
        return;
    }
    dispatch("login", {username: username, tok: result.tok});
}
function oncancel(e) {
    username = "";
    pwd = "";
    dispatch("cancel");
}
</script>

<form class="bg-gray-800 p-2">
    <div class="mb-2">
        <label class="block bg-gray-800 text-gray-200" for="username">username</label>
        <input class="block border border-gray-500 bg-gray-200 text-gray-800 py-0 px-2 w-full" id="username" name="username" type="text" bind:value={username}>
    </div>
    <div class="mb-2">
        <label class="block bg-gray-800 text-gray-200" for="pwd">password</label>
        <input class="block border border-gray-500 bg-gray-200 text-gray-800 py-0 px-2 w-full" id="pwd" name="pwd" type="password" bind:value={pwd}>
    </div>
{#if frm.status != ""}
    <div class="mb-2">
        <p class="font-bold bg-gray-800 text-gray-200">{frm.status}</p>
    </div>
{/if}
    <div class="flex flex-row justify-center">
        <div>
{#if frm.mode == "loading"}
            <button disabled on:click={onlogin} class="inline mx-auto py-1 px-2 bg-gray-200 text-gray-800 mr-2">Login</button>
{:else}
            <button on:click={onlogin} class="inline mx-auto py-1 px-2 bg-gray-200 text-gray-800 mr-2">Login</button>
{/if}
            <button on:click={oncancel} class="inline mx-auto py-1 px-2 bg-gray-200 text-gray-800">Cancel</button>
        </div>
    </div>
</form>

