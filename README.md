# whatsthis

[![Test](https://github.com/odwrtw/whatsthis/actions/workflows/test.yml/badge.svg)](https://github.com/odwrtw/whatsthis/actions/workflows/test.yml)

Pure Go library that guesses video metadata (title, year, codec, resolution, release group, etc.) from a filename. Zero dependencies. Inspired by the Python [guessit](https://github.com/guessit-io/guessit) project.

## Installation

```sh
go get github.com/odwrtw/whatsthis
```

## Usage

```go
info := whatsthis.Video("Big.Buck.Bunny.2008.1080p.BluRay.x264-YIFY.mkv")
fmt.Println(info.Type)         // movie
fmt.Println(info.Title)        // Big Buck Bunny
fmt.Println(info.Year)         // 2008
fmt.Println(info.ScreenSize)   // 1080p
fmt.Println(info.VideoCodec)   // H.264
fmt.Println(info.ReleaseGroup) // YIFY
fmt.Println(info.Container)    // mkv
```

## WebAssembly demo

The static web interface runs the parser entirely in the browser; filenames are not uploaded anywhere.

Try it at [odwrtw.github.io/whatsthis](https://odwrtw.github.io/whatsthis/).

Build it with the repository's Go version:

```sh
./scripts/build-web.sh
```

Then serve `dist` with any static HTTP server and open `index.html`. The generated directory contains the site assets, the WebAssembly binary, and the matching `wasm_exec.js` from the local Go toolchain.

Pushes to `master` deploy the site through GitHub Actions. To enable deployment, select **GitHub Actions** as the source under **Settings → Pages** in the repository.
