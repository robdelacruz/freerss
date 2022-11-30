# Usage:
# 'make dep' and 'make webtools' to install dependencies.
# 'make clean' to clear all work files
# 'make' to build css and js into static/
# 'make serve' to start dev webserver

NODE_VER = 17

JSFILES = index.js App.svelte Grid.svelte RSSView.svelte LoginForm.svelte SignupForm.svelte EditUserForm.svelte DelUserForm.svelte

all: freerss static/style.css static/bundle.js

nodejs:
	curl -fsSL https://deb.nodesource.com/setup_$(NODE_VER).x | sudo bash -
	sudo apt install nodejs
	sudo npm install -g npx

dep:
	go env -w GO111MODULE=auto
	go get github.com/mmcdole/gofeed
	go get github.com/gorilla/feeds

webtools:
	npm install --save-dev tailwindcss
	npm install --save-dev postcss
	npm install --save-dev postcss-cli
	npm install --save-dev cssnano
	npm install --save-dev svelte
	npm install --save-dev rollup
	npm install --save-dev rollup-plugin-svelte
	npm install --save-dev @rollup/plugin-node-resolve

static/style.css: twsrc.css tailwind.config.js
	npx tailwind -i twsrc.css -o twsrc.o 1>/dev/null
	npx postcss twsrc.o > static/style.css
	#npx tailwind -i twsrc.css -o static/style.css 1>/dev/null

freerss: freerss.go
	go build -o freerss freerss.go

static/bundle.js: $(JSFILES)
	npx rollup -c

clean:
	rm -rf freerss static/*.js static/*.css static/*.map

serve:
	python -m SimpleHTTPServer

