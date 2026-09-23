// Package whatsthis extracts structured metadata (title, year, quality, codec,
// release group, etc.) from video filenames and release names.
package whatsthis

import (
	"path/filepath"
	"strings"
)

// Type represents the type of media (episode, movie, etc.).
type Type string

const (
	// Episode indicates a TV show episode.
	Episode Type = "episode"
	// ShowSeason indicates a complete TV show season.
	ShowSeason Type = "show_season"
	// Movie indicates a movie.
	Movie Type = "movie"
)

// Info holds structured metadata extracted from a filename or release name.
type Info struct {
	Type         Type   `json:"type"`
	Title        string `json:"title"`
	Episode      int    `json:"episode"`
	Season       int    `json:"season"`
	Year         int    `json:"year"`
	ScreenSize   string `json:"screen_size"`
	ReleaseGroup string `json:"release_group"`
	AudioCodec   string `json:"audio_codec"`
	VideoCodec   string `json:"video_codec"`
	Container    string `json:"container"`
	MimeType     string `json:"mimetype"`
}

// File parses a file or extensionless release name and returns structured
// metadata. Whole-season names are recognized before falling back to Video.
func File(input string) Info {
	if hasFileExtension(input) {
		return Video(input)
	}
	if info, ok := parseSeasonName(input); ok {
		return info
	}
	return Video(input)
}

// hasFileExtension distinguishes a filename suffix from a dot-separated
// release-name token such as .1080p, .x265 or .x265-GROUP.
func hasFileExtension(input string) bool {
	ext := strings.TrimPrefix(filepath.Ext(input), ".")
	if ext == "" || isSeasonMetadataBracket(ext) ||
		!((ext[0] >= 'a' && ext[0] <= 'z') || (ext[0] >= 'A' && ext[0] <= 'Z')) ||
		strings.ContainsAny(ext, "-[]_ ()") {
		return false
	}
	return true
}

// Video parses a video filename and returns structured metadata.
func Video(input string) Info {
	return parse(input)
}
