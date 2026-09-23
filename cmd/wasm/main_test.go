//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"
	"testing"

	"github.com/odwrtw/whatsthis"
)

func TestFile(t *testing.T) {
	encoded, ok := file(js.Undefined(), []js.Value{
		js.ValueOf("Velvet Meridian 2026 Season 1 Complete 1080p WEB x264 [theta_group]"),
	}).(string)
	if !ok {
		t.Fatal("file did not return a string")
	}

	var got whatsthis.Info
	if err := json.Unmarshal([]byte(encoded), &got); err != nil {
		t.Fatalf("file returned invalid JSON: %v", err)
	}

	if got.Type != whatsthis.ShowSeason || got.Title != "Velvet Meridian" || got.Year != 2026 || got.Season != 1 || got.Episode != 0 || got.ReleaseGroup != "theta_group" {
		t.Fatalf("file returned unexpected metadata: %+v", got)
	}
}

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

func TestFileRejectsInvalidArguments(t *testing.T) {
	if got := file(js.Undefined(), nil); got != "" {
		t.Fatalf("file with no arguments returned %#v", got)
	}
	if got := file(js.Undefined(), []js.Value{js.ValueOf(42)}); got != "" {
		t.Fatalf("file with a non-string argument returned %#v", got)
	}
}
