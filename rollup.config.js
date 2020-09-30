import svelte from "rollup-plugin-svelte";
import resolve from "@rollup/plugin-node-resolve";

export default {
    input: "index.js",
    output: {
        file: "static/bundle.js",
        format: "iife",
        name: "app",
        sourcemap: false
    },
    plugins: [
        svelte({
            dev: true,
            css: function(css) {
                css.write("static/bundle.css");
            }
        }),
        resolve({
            browser: true,
            dedupe: ["svelte"]
        })
    ]
};

