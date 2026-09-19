//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"
	"testing"

	"github.com/odwrtw/whatsthis"
)

func TestVideo(t *testing.T) {
	encoded, ok := video(js.Undefined(), []js.Value{
		js.ValueOf("Big.Buck.Bunny.2008.1080p.BluRay.x264-YIFY.mkv"),
	}).(string)
	if !ok {
		t.Fatal("video did not return a string")
	}

	var got whatsthis.Info
	if err := json.Unmarshal([]byte(encoded), &got); err != nil {
		t.Fatalf("video returned invalid JSON: %v", err)
	}

	if got.Title != "Big Buck Bunny" || got.Year != 2008 || got.ReleaseGroup != "YIFY" {
		t.Fatalf("video returned unexpected metadata: %+v", got)
	}
}

func TestVideoRejectsInvalidArguments(t *testing.T) {
	if got := video(js.Undefined(), nil); got != "" {
		t.Fatalf("video with no arguments returned %#v", got)
	}
	if got := video(js.Undefined(), []js.Value{js.ValueOf(42)}); got != "" {
		t.Fatalf("video with a non-string argument returned %#v", got)
	}
}
