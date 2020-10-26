<div>
    <div class="flex flex-row justify-between border-b border-gray-500 text-gray-200 pb-1 mb-2">
        <div>
            <h1 class="inline self-end text-sm ml-1 mr-2"><a href="/">FreeRSS</a></h1>
            <a href="about.html" class="self-end mr-2">About</a>
            <a href="#a" class="text-xs bg-gray-400 text-gray-800 self-center rounded px-2 mr-1" on:click={onaddwidget}>Add Widget</a>
        </div>
        <div>
{#if ui.username != ""}
            <div class="relative inline mr-2">
                <a class="mr-1" href="#a" on:click={onmenu}>
                    {ui.username}
                </a>
            {#if ui.showmenu}
                <div class="absolute top-auto right-0 py-1 bg-gray-200 text-gray-800 w-20 border border-gray-500 shadow-xs w-32">
                    <a href="#a" class="block leading-none px-2 py-1 hover:bg-gray-400 hover:text-gray-900" role="menuitem" on:click={onchangepassword}>Change Password</a>
                    {#if ui.username != "admin"}
                    <a href="#a" class="block leading-none px-2 py-1 hover:bg-gray-400 hover:text-gray-900" role="menuitem" on:click={ondeluser}>Delete Account</a>
                    {/if}
                    {#if ui.username == "admin"}
                    <a href="#a" class="block leading-none px-2 py-1 hover:bg-gray-400 hover:text-gray-900" role="menuitem" on:click={onresetlocalstorage}>Reset LocalStorage</a>
                    {/if}
                </div>
            {/if}
            </div>
            <a href="#a" class="inline self-end mr-1" on:click={onlogout}>Logout</a>
{:else}
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
{:else if ui.mode == "edituser"}
    <div class="flex flex-row w-full">
        <div class="widget">
            <EditUserForm username="{ui.username}" on:update={loginform_login} on:cancel={loginform_cancel} />
        </div>
    </div>
{:else if ui.mode == "deluser"}
    <div class="flex flex-row w-full">
        <div class="widget">
            <DelUserForm username="{ui.username}" on:del={onlogout} on:cancel={loginform_cancel} />
        </div>
    </div>
{/if}
</div>

<script>
import Grid from "./Grid.svelte";
import LoginForm from "./LoginForm.svelte";
import SignupForm from "./SignupForm.svelte";
import EditUserForm from "./EditUserForm.svelte";
import DelUserForm from "./DelUserForm.svelte";
let grid;
let ui = {};
ui.mode = "";
ui.showmenu = false;

let session = currentSession();
ui.username = session.username;
ui.tok = session.tok;

document.addEventListener("click", onappclick, false);

function onmenu(e) {
    e.preventDefault();
    e.stopPropagation();
    ui.showmenu = !ui.showmenu;
}
function onappclick(e) {
    ui.showmenu = false;

    // Send signal to close any open pop-up menus in RSSViews.
    let rssviews = document.querySelectorAll(".rssview");
    for (let i=0; i < rssviews.length; i++) {
        let rssview = rssviews[i];
        let e = new Event("appclick");
        rssview.dispatchEvent(e);
    }
}
function onchangepassword(e) {
    ui.showmenu = false;
    ui.mode = "edituser";
}
function ondeluser(e) {
    ui.showmenu = false;
    ui.mode = "deluser";
}
function onresetlocalstorage(e) {
    localStorage.removeItem("cols");
}

function onaddwidget(e) {
    grid.addwidget();
}

function onlogin(e) {
    ui.mode = "login";
}
function onlogout(e) {
    ui.mode = "";
    ui.showmenu = false;
    ui.username = "";
    ui.tok = "";
    document.cookie = `usernametok=;path=/`;
}
function loginform_login(e) {
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

