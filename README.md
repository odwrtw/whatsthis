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
