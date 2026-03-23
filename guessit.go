package guessit

// Type represents the type of media (episode, movie, etc.).
type Type string

const (
	// Episode indicates a TV show episode.
	Episode Type = "episode"
	// Movie indicates a movie.
	Movie Type = "movie"
)

// Guess holds the structured metadata extracted from a video filename.
type Guess struct {
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
	Format       string `json:"format"`
	MimeType     string `json:"mimetype"`
}

// GuessIt parses a video filename and returns structured metadata.
func GuessIt(input string) Guess {
	return parse(input)
}
