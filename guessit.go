package guessit

type Type string

const (
	Movie Type = "movie"
	Show  Type = "show"
)

type Guess struct {
	Type       Type   `json:"type"`
	Title      string `json:"title"`
	Episode    int    `json:"episode"`
	Season     int    `json:"season"`
	Year       int    `json:"year"`
	ScreenSize string `json:"screen_size"`
	Release    string `json:"release"`
	AudioCodec string `json:"audio_codec"`
	VideoCodec string `json:"video_codec"`
	Container  string `json:"container"`
	Format     string `json:"format"`
	MimeType   string `json:"mime_type"`
}

func GuessIt(input string) Guess {
	return Guess{}
}
