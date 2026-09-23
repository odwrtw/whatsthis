//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/odwrtw/whatsthis"
)

var (
	fileFunc  js.Func
	videoFunc js.Func
)

func file(_ js.Value, args []js.Value) any {
	if len(args) != 1 || args[0].Type() != js.TypeString {
		return ""
	}

	result, err := json.Marshal(whatsthis.File(args[0].String()))
	if err != nil {
		return ""
	}

	return string(result)
}

func video(_ js.Value, args []js.Value) any {
	if len(args) != 1 || args[0].Type() != js.TypeString {
		return ""
	}

	result, err := json.Marshal(whatsthis.Video(args[0].String()))
	if err != nil {
		return ""
	}

	return string(result)
}

func main() {
	fileFunc = js.FuncOf(file)
	videoFunc = js.FuncOf(video)

	api := js.Global().Get("Object").New()
	api.Set("file", fileFunc)
	api.Set("video", videoFunc)
	js.Global().Set("whatsthis", api)

	select {}
}
