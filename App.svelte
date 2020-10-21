<div class="flex flex-row justify-between border-b border-gray-500 text-gray-200 pb-1 mb-2">
    <div>
        <h1 class="inline self-end text-sm ml-1 mr-2">FreeRSS</h1>
        <a href="about.html" class="self-end mr-2">About</a>
        <a href="#a" class="text-xs bg-gray-400 text-gray-800 self-center rounded px-2 mr-1" on:click={onaddwidget}>Add Widget</a>
    </div>
    <div>
{#if ui.username != ""}
        <p class="inline mr-2">{ui.username}</p>
        <a href="#a" class="inline self-end mr-1" on:click={onlogout}>Logout</a>
{:else}
        <p class="inline mr-2">(no user)</p>
        <a href="#a" class="inline self-end mr-1" on:click={onlogin}>Login</a>
{/if}
    </div>
</div>
{#if ui.mode == ""}
    <Grid bind:this={grid} username={ui.username} tok={ui.tok} />
{:else if ui.mode == "login"}
    <div class="flex flex-row w-full">
        <div class="widget">
            <LoginForm username="" pwd="" on:login={loginform_login} on:cancel={loginform_cancel} on:createaccount={loginform_createaccount} />
        </div>
    </div>
{:else if ui.mode == "signup"}
    <div class="flex flex-row w-full">
        <div class="widget">
            <SignupForm username="" pwd="" on:signup={loginform_login} on:cancel={loginform_cancel} />
        </div>
    </div>
{/if}

<script>
import Grid from "./Grid.svelte";
import LoginForm from "./LoginForm.svelte";
import SignupForm from "./SignupForm.svelte";
let grid;
let ui = {};
ui.mode = "";

let session = currentSession();
ui.username = session.username;
ui.tok = session.tok;

function onaddwidget(e) {
    grid.addwidget();
}

function onlogin(e) {
    ui.mode = "login";
}
function onlogout(e) {
    ui.username = "";
    ui.tok = "";
    document.cookie = `usernametok=;path=/`;
}
function loginform_login(e) {
    console.error("loginform_login");
    ui.mode = "";
    let username = e.detail.username;
    let tok = e.detail.tok;
    document.cookie = `usernametok=${username}|${tok};path=/`;

    ui.username = username;
    ui.tok = tok;
}
function loginform_cancel(e) {
    ui.mode = "";
}
function loginform_createaccount(e) {
    ui.mode = "signup";
}

function currentSession() {
    let cookies = document.cookie.split(";");
    for (let i=0; i < cookies.length; i++) {
        let cookie = cookies[i].trim();
        let [k,v] = cookie.split("=");
        if (k != "usernametok") {
            continue;
        }
        if (v == undefined) {
            v = "";
        }
        let [username, tok] = v.split("|");
        if (tok == undefined) {
            tok = "";
        }
        return {username: username, tok: tok};
    }
    return {username: "", tok: ""};
}

</script>

