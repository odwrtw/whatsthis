# go-guessit

[![Test](https://github.com/odwrtw/go-guessit/actions/workflows/test.yml/badge.svg)](https://github.com/odwrtw/go-guessit/actions/workflows/test.yml)

Pure Go library that guesses video metadata (title, year, codec, resolution, release group, etc.) from a filename. Zero dependencies. Inspired by the Python [guessit](https://github.com/guessit-io/guessit) project.

## Installation

```sh
go get github.com/odwrtw/go-guessit
```

## Usage

```go
guess := guessit.GuessIt("Big.Buck.Bunny.2008.1080p.BluRay.x264-YIFY.mkv")
fmt.Println(guess.Type)         // movie
fmt.Println(guess.Title)        // Big Buck Bunny
fmt.Println(guess.Year)         // 2008
fmt.Println(guess.ScreenSize)   // 1080p
fmt.Println(guess.VideoCodec)   // H.264
fmt.Println(guess.ReleaseGroup) // YIFY
fmt.Println(guess.Container)    // mkv
```
