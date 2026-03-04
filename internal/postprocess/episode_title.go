package postprocess

import (
	"regexp"
	"strings"
)

func deriveEpisodeTitle(input string, result Result) {
	if _, ok := result.Get("episode"); !ok {
		return
	}
	re := regexp.MustCompile(`(?i)(?:S\d{1,2}[ ._-]?E\d{1,3}|\d{1,2}x\d{1,3})[ ._-]+(.+)`)
	m := re.FindStringSubmatch(input)
	if len(m) != 2 {
		return
	}
	title := Cleanup(m[1])
	title = regexp.MustCompile(`(?i)[ ._-](hdtv|webrip|web[-_. ]?dl|bluray|bdrip|dvdrip|x264|x265|h\.?264|h\.?265|aac|ac3|dts).*$`).ReplaceAllString(title, "")
	title = regexp.MustCompile(`(?i)[ ._-](mkv|mp4|avi|srt|nfo)$`).ReplaceAllString(title, "")
	title = strings.TrimSuffix(title, ".")
	if title != "" {
		result.Set("episode_title", title)
	}
}
