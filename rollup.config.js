import svelte from "rollup-plugin-svelte";
import resolve from "@rollup/plugin-node-resolve";

export default {
    input: "index.js",
    output: {
        file: "static/bundle.js",
        format: "iife",
        name: "app",
        sourcemap: true
    },
    plugins: [
        svelte({
            compilerOptions: {
                dev: true,
                css: false,
            }
        }),
        resolve({
            browser: true,
            dedupe: ["svelte"]
        })
    ]
};

