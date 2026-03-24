// Package whatsthis extracts structured metadata (title, year, quality, codec,
// release group, etc.) from video filenames.
package whatsthis

// Type represents the type of media (episode, movie, etc.).
type Type string

const (
	// Episode indicates a TV show episode.
	Episode Type = "episode"
	// Movie indicates a movie.
	Movie Type = "movie"
)

// Info holds the structured metadata extracted from a video filename.
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

// Video parses a video filename and returns structured metadata.
func Video(input string) Info {
	return parse(input)
}
