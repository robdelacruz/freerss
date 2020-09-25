import App from "./App.svelte";
const app = new App({
    target: document.querySelector("#hello1"),
    props: {
        name: "rob",
    }
});

export default app;

