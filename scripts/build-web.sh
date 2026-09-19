#!/bin/sh
set -eu

cd "$(dirname "$0")/.."

site=dist
rm -rf "$site"
mkdir -p "$site"

cp web/index.html web/app.js web/style.css "$site/"
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$site/"
GOOS=js GOARCH=wasm go build -trimpath -ldflags='-s -w' -o "$site/whatsthis.wasm" ./cmd/wasm
: >"$site/.nojekyll"
