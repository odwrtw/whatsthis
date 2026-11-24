package whatsthis

import (
	"regexp"
	"strconv"
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

// A compact trailing credit after technical metadata is not episode numbering.
func isEpisodeCredit(name string, start int) bool {
	if start <= 0 || !strings.ContainsRune("-[(", rune(name[start-1])) || !hasPrecedingMetadata(name[:start]) {
		return false
	}
	prefix := name[:start]
	if episodeMetadataStart(prefix) >= 0 {
		return true
	}
	// A bare resolution is weaker evidence: it must immediately precede the
	// credit, not occur inside a title such as "The 720 Show".
	tokens := strings.FieldsFunc(prefix, isMetaSep)
	return len(tokens) > 0 && reScreenSizeNumberToken.MatchString(tokens[len(tokens)-1])
}

// extractEpisodes reads multi-episode continuations attached to an explicit
// SxxEyy or NxMM marker: lists (S01E02E03, S01E02.S01E03, 1x02x03) and ranges
// (S01E02-03, S01E02-E04, S01E02 - E04, 1x02-1x04). It returns the sorted,
// unique episodes, or nil when fewer than two are named or the continuation
// names another season or a descending range, and the continuation end.
func extractEpisodes(name string, season, episode, start, end int) ([]int, int) {
	if start < 0 || episode > 999 {
		return nil, end
	}
	// Replacing ASCII separators preserves byte offsets into the original name.
	name = strings.ReplaceAll(name, "_", " ")
	marker := reEpisodeMarker.FindStringSubmatch(name[start:end])
	if marker == nil {
		return nil, end
	}
	cross := marker[1] != ""
	var named [1000]bool
	named[episode] = true
	previous, valid := episode, true
	for {
		number, isRange, sameSeason, next, ok := episodeContinuation(name, end, cross, season)
		if !ok {
			break
		}
		if !sameSeason || isRange && number < previous {
			valid = false
		}
		first := number
		if isRange && number >= previous {
			first = previous
		}
		for i := first; i <= number; i++ {
			named[i] = true
		}
		previous, end = number, next
	}
	var episodes []int
	for i, ok := range named {
		if ok {
			episodes = append(episodes, i)
		}
	}
	if !valid || len(episodes) < 2 {
		return nil, end
	}
	return episodes, end
}

// episodeContinuation parses one continuation at pos. Bare numbers need an
// attached dash and must not be resolutions or decimals (S01E02-720, -5.1);
// spaced dashes need an E label, so " - 05 - Title" remains a title.
func episodeContinuation(name string, pos int, cross bool, season int) (number int, isRange, sameSeason bool, end int, ok bool) {
	tail := name[pos:]
	m := reEpisodeContinuation.FindStringSubmatchIndex(tail)
	if m == nil {
		return 0, false, false, 0, false
	}
	separator, label := "", ""
	if m[2] >= 0 {
		separator = tail[m[2]:m[3]]
	}
	if m[4] >= 0 {
		label = strings.ToLower(tail[m[4]:m[5]])
	}
	digits := tail[m[10]:m[11]]
	end = m[11]
	labelSeason := -1
	switch {
	case m[6] >= 0 && !cross:
		labelSeason, _ = strconv.Atoi(tail[m[6]:m[7]])
	case m[8] >= 0 && cross:
		labelSeason, _ = strconv.Atoi(tail[m[8]:m[9]])
	case label == "e" && !cross, label == "x" && cross && digits != "264" && digits != "265":
	case label == "" && separator == "-":
		if reScreenSizeNumberToken.MatchString(digits) || reChannelSuffix.MatchString(tail[m[10]:]) {
			return 0, false, false, 0, false
		}
	default:
		return 0, false, false, 0, false
	}
	if separator == "" && label == "" || separator == "." && m[6] < 0 ||
		len(separator) > 1 && label != "e" && m[6] < 0 {
		return 0, false, false, 0, false
	}
	// The number must end at a boundary or at the next labelled continuation.
	if end < len(tail) && isAlnum(tail[end]) {
		if next := reEpisodeContinuation.FindStringSubmatchIndex(tail[end:]); next == nil || next[4] < 0 || tail[end] <= '9' {
			return 0, false, false, 0, false
		}
	}
	number, _ = strconv.Atoi(digits)
	return number, strings.Contains(separator, "-"), labelSeason < 0 || labelSeason == season, pos + end, true
}

func isAlnum(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
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
