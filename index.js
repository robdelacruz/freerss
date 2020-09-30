import RSSView from "./RSSView.svelte";
let rssview = new RSSView({
    target: document.querySelector("#container"),
    props: {
        url: "http://rss.slashdot.org/Slashdot/slashdotMain",
        maxitems: 15,
    },
});

