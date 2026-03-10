package guessit

// Result is the parsed metadata.
type Result struct {
	Type         string `json:"type,omitempty"`
	Title        string `json:"title,omitempty"`
	Episode      int    `json:"episode,omitempty"`
	Season       int    `json:"season,omitempty"`
	Year         int    `json:"year,omitempty"`
	ScreenSize   string `json:"screen_size,omitempty"`
	ReleaseGroup string `json:"release_group,omitempty"`
	AudioCodec   string `json:"audio_codec,omitempty"`
	VideoCodec   string `json:"video_codec,omitempty"`
	Container    string `json:"container,omitempty"`
	MIMEType     string `json:"mimetype,omitempty"`
}
