package whatsthis

import (
	"regexp"
	"strings"
)

// Anchor once at initialization: searching every remaining suffix for an
// unanchored codec would make a long episode title quadratic to parse.
var episodeCodecStarts = func() []*regexp.Regexp {
	patterns := []*regexp.Regexp{
		reVideoCodecH264, reVideoCodecH265, reVideoCodecH264Spaced, reVideoCodecH265Spaced,
		reVideoCodecXvid, reVideoCodecAV1, reVideoCodecVP9, reVideoCodecVC1,
		reAudioDDP, reAudioDD, reAudioEAC3, reAudioAAC, reAudioDTS, reAudioFLAC, reAudioOPUS, reAudioTrueHD, reAudioMP3,
	}
	for i, pattern := range patterns {
		patterns[i] = regexp.MustCompile("^(?:" + pattern.String() + ")")
	}
	return patterns
}()

func hasFoldPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix)
}

func startsWithCodec(s string) bool {
	for _, pattern := range episodeCodecStarts {
		if pattern.MatchString(s) {
			return true
		}
	}
	return false
}

// isEpisodeMetadataOnly rejects ordinary words before retaining tags or years.
// Weak tags alone still need a metadata anchor; callers establish that context.
func isEpisodeMetadataOnly(meta string) bool {
	for _, token := range strings.FieldsFunc(meta, isMetaSep) {
		if !isKnownToken(token) && !(len(token) == 4 && reYear.MatchString(token)) {
			return false
		}
	}
	return true
}

// Episode-title words are not metadata anchors merely because they include a
// qualifier such as Internal, Proper or French. Start at a technical tag, a
// source, or an explicit language marker instead.
func episodeMetadataStart(meta string) int {
	for start := 0; start < len(meta); {
		if isMetaSep(rune(meta[start])) {
			start++
			continue
		}
		end := start
		for end < len(meta) && !isMetaSep(rune(meta[end])) {
			end++
		}
		token := meta[start:end]
		lower := strings.ToLower(token)
		if startsWithCodec(meta[start:]) {
			return start
		}
		if extractScreenSize(token) != "" && !reScreenSizeNumberToken.MatchString(token) {
			return start
		}
		switch lower {
		case "hdtv", "pdtv", "webrip", "webdl", "brrip", "brip", "bdrip", "bluray", "dvdrip", "hdrip":
			return start
		case "web", "blu":
			if hasFoldPrefix(meta[start:], "web-dl") || hasFoldPrefix(meta[start:], "blu-ray") || token == strings.ToUpper(token) {
				return start
			}
		}
		if lang := lookupReleaseLanguage(lower); lang.valid() {
			tail := strings.TrimLeft(meta[end:], " ._-")
			next := tail
			if i := strings.IndexFunc(tail, isMetaSep); i >= 0 {
				next = tail[:i]
			}
			tagContext := tail == "" || meta[end] == ']' || next != "" && isKnownToken(next)
			if token == strings.ToUpper(token) && tagContext || len(lower) > 2 && !lang.fullName {
				return start
			}
		}
		if (lower == "multi" || lower == "dual" || lower == "dl") && token == strings.ToUpper(token) {
			return start
		}
		start = end
	}
	return -1
}
