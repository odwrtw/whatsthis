package guessit

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var (
	reReleaseGroupScenePrefix = regexp.MustCompile(`^[A-Za-z0-9]{3,8}$`)
	reEpisodeLike             = regexp.MustCompile(`(?i)\bs\d{1,3}(?:[ ._-]?e|x(?:e)?)[ ._-]?\d{1,3}\b|\b\d{1,2}x\d{1,3}\b`)
	reTrailingParenBlock      = regexp.MustCompile(`\(([^()]*)\)\s*$`)
	reTypeEpisode             = regexp.MustCompile(`(?i)\bS\d{1,3}[ ._-]?E\d{1,3}(?:[ ._-]?E\d{1,3})*\b|\b\d{1,2}x\d{1,3}(?:x\d{1,3})*\b`)
	reTypeSeasonToken         = regexp.MustCompile(`(?i)(?:^|[ ._-])S\d{1,3}(?:$|[ ._-])`)
	reTypeTSeasonToken        = regexp.MustCompile(`(?i)(?:^|[ ._-])T\d{1,2}(?:$|[ ._-])`)
	reTypeEpisodeHints        = regexp.MustCompile(`(?i)(?:^|[ ._-])(?:trailer|bonus|deleted[ ._-]?scenes?)(?:$|[ ._-])`)
	reTypeCompact3Digits      = regexp.MustCompile(`(?i)\b\d{3}\b`)
	reTypeYear                = regexp.MustCompile(`\b(19\d{2}|20\d{2})\b`)
	reSeasonEpisodeSE         = regexp.MustCompile(`(?i)^s(\d{1,3})[ ._-]?e(\d{1,3})$`)
	reSeasonEpisodeX          = regexp.MustCompile(`(?i)^s?(\d{1,3})x(?:e)?(\d{1,3})$`)
	reTokenSplit              = regexp.MustCompile(`[ ._]+`)

	reTitleSeasonEpisodeBoundary = regexp.MustCompile(`(?i)[ ._-]s\d{1,3}[ ._-]?e\d{1,3}\b`)
	reTitleLeadingSeasonEpisode  = regexp.MustCompile(`(?i)^s\d{1,3}[ ._-]?e\d{1,3}\b[ ._-]*`)
	reTitleYearBoundary          = regexp.MustCompile(`(?i)[ ._-](19\d{2}|20[0-3]\d)\b`)
	reTitleArticlePrefix         = regexp.MustCompile(`^([a-z]{3,6})[._-]+((?:the|a|an)[._-].+)$`)
	reTitleShortScenePrefix      = regexp.MustCompile(`^([a-z0-9]{3,4})-(.+)$`)
)

var (
	releaseGroupTechHints = []string{"web-dl", "webdl", "webrip", "hdtv", "bluray", "x264", "h264", "x265", "h265", "dd", "ddp", "1080p", "720p", "2160p"}
	releaseGroupTagHints  = []string{"web-dl", "webdl", "webrip", "hdtv", "bluray", "x264", "h264", "x265", "h265", "dd", "ddp", "aac", "ac3", "dts", "1080p", "720p", "2160p"}
	blockedReleaseGroup   = map[string]struct{}{
		"h264": {}, "x264": {}, "h265": {}, "x265": {}, "1080p": {}, "720p": {},
		"hdtv": {}, "webrip": {}, "webdl": {}, "bluray": {}, "dvdrip": {}, "proper": {},
	}
)

type match struct {
	Name  string
	Value any
	Start int
	End   int
	Raw   string
	Tags  []string
}

type parsedResult map[string]any

func (r parsedResult) get(name string) (any, bool) {
	v, ok := r[name]
	return v, ok
}

func (r parsedResult) set(name string, value any) {
	r[name] = value
}

func postProcessResult(input string, matches []match, result parsedResult) {
	deriveSeasonEpisode(result)
	deriveProperCount(result)
	deriveReleaseGroup(input, result)
	deriveMIMEType(result)
	deriveTitle(input, matches, result)
	deriveEpisodeTitle(input, result)
	deriveType(input, result)
}

func parseSeasonEpisode(raw string) (int, int, bool) {
	if m := reSeasonEpisodeSE.FindStringSubmatch(raw); len(m) == 3 {
		s, errS := strconv.Atoi(m[1])
		e, errE := strconv.Atoi(m[2])
		return s, e, errS == nil && errE == nil
	}
	if m := reSeasonEpisodeX.FindStringSubmatch(raw); len(m) == 3 {
		s, errS := strconv.Atoi(m[1])
		e, errE := strconv.Atoi(m[2])
		return s, e, errS == nil && errE == nil
	}
	return 0, 0, false
}

func deriveReleaseGroup(input string, result parsedResult) {
	if _, exists := result.get("release_group"); exists {
		return
	}
	base := input
	if i := strings.LastIndexAny(base, `/\`); i >= 0 && i+1 < len(base) {
		base = base[i+1:]
	}
	if dot := strings.LastIndex(base, "."); dot > 0 {
		base = base[:dot]
	}
	dash := strings.LastIndex(base, "-")
	if dash >= 0 && dash+1 < len(base) {
		if firstDash := strings.Index(base, "-"); firstDash == dash && firstDash > 0 {
			prefix := base[:firstDash]
			suffixLower := strings.ToLower(base[firstDash+1:])
			if reReleaseGroupScenePrefix.MatchString(prefix) && reEpisodeLike.MatchString(suffixLower) {
				result.set("release_group", prefix)
				return
			}
		}
	}
	candidate := ""
	if m := reTrailingParenBlock.FindStringSubmatch(base); len(m) == 2 {
		inner := strings.TrimSpace(m[1])
		innerLower := strings.ToLower(inner)
		if containsAnyHint(innerLower, releaseGroupTechHints) {
			fields := strings.Fields(inner)
			if len(fields) > 0 {
				candidate = strings.Trim(fields[len(fields)-1], "[](){}.,;")
			}
		}
	}
	if candidate == "" {
		if dash < 0 || dash+1 >= len(base) {
			return
		}
		candidate = strings.TrimSpace(base[dash+1:])
	}
	if lb := strings.LastIndex(candidate, "["); lb > 0 && strings.HasSuffix(candidate, "]") {
		prefix := strings.ToLower(candidate[:lb])
		tag := strings.TrimSpace(candidate[lb+1 : len(candidate)-1])
		if tag != "" {
			useTag := strings.Contains(prefix, ".mkv") || strings.Contains(prefix, ".mp4") || strings.Contains(prefix, ".avi") || strings.Contains(prefix, ".m4v")
			if !useTag && containsAnyHint(prefix, releaseGroupTagHints) {
				useTag = true
			}
			if useTag {
				candidate = tag
			}
		}
	}
	if len(candidate) >= 2 {
		pairs := [][2]byte{{'[', ']'}, {'(', ')'}, {'{', '}'}}
		for _, p := range pairs {
			if candidate[0] == p[0] && candidate[len(candidate)-1] == p[1] {
				candidate = strings.TrimSpace(candidate[1 : len(candidate)-1])
				break
			}
		}
	}
	candidate = strings.TrimFunc(candidate, isSeparator)
	candidate = refineReleaseGroupCandidate(candidate)
	if candidate == "" {
		return
	}
	if strings.Contains(candidate, "_") {
		if _, hasSeason := result.get("season"); hasSeason {
			return
		}
		if _, hasEpisode := result.get("episode"); hasEpisode {
			return
		}
	}
	lc := strings.ToLower(candidate)
	if reEpisodeLike.MatchString(lc) {
		return
	}
	if _, bad := blockedReleaseGroup[lc]; bad {
		return
	}
	result.set("release_group", candidate)
}

func refineReleaseGroupCandidate(candidate string) string {
	if candidate == "" {
		return ""
	}
	if strings.ContainsAny(candidate, "._ ") {
		parts := reTokenSplit.Split(candidate, -1)
		tokens := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(strings.Trim(p, "[](){}.,;"))
			if p == "" {
				continue
			}
			tokens = append(tokens, p)
		}
		if len(tokens) > 1 {
			first := strings.ToLower(tokens[0])
			if first == "dl" || strings.Contains(first, "subs") || containsAnyHint(first, releaseGroupTagHints) || strings.HasPrefix(first, "dd") {
				return tokens[len(tokens)-1]
			}
		}
	}
	return candidate
}

func deriveSeasonEpisode(result parsedResult) {
	pair, ok := result.get("season_episode_pair")
	if !ok {
		pair, ok = result.get("season_episode_pair_compact")
		if !ok {
			return
		}
	}
	raw, ok := pair.(string)
	if !ok {
		return
	}
	season, episode, ok := parseSeasonEpisode(raw)
	if !ok {
		return
	}
	result.set("season", season)
	result.set("episode", episode)
}

func deriveProperCount(result parsedResult) {
	v, ok := result.get("other")
	if !ok {
		return
	}
	if s, ok := v.(string); ok && s == "Proper" {
		result.set("proper_count", 1)
	}
}

func deriveMIMEType(result parsedResult) {
	c, ok := result.get("container")
	if !ok {
		return
	}
	s, ok := c.(string)
	if !ok {
		return
	}
	switch strings.ToLower(s) {
	case "mkv":
		result.set("mimetype", "video/x-matroska")
	case "mp4", "m4v":
		result.set("mimetype", "video/mp4")
	case "avi":
		result.set("mimetype", "video/x-msvideo")
	}
}

func deriveType(input string, result parsedResult) {
	if _, ok := result.get("season"); ok {
		result.set("type", "episode")
		return
	}
	if _, ok := result.get("episode"); ok {
		result.set("type", "episode")
		return
	}
	if reTypeEpisode.MatchString(input) {
		result.set("type", "episode")
		return
	}
	if reTypeSeasonToken.MatchString(input) {
		result.set("type", "episode")
		return
	}
	if reTypeTSeasonToken.MatchString(input) && reTypeEpisodeHints.MatchString(input) {
		result.set("type", "episode")
		return
	}
	if reTypeCompact3Digits.MatchString(input) && !reTypeYear.MatchString(input) {
		result.set("type", "episode")
		return
	}
	result.set("type", "movie")
}

func deriveTitle(input string, matches []match, result parsedResult) {
	base := filepath.Base(input)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	if i := reTitleLeadingSeasonEpisode.FindStringIndex(base); len(i) == 2 && i[0] == 0 {
		base = base[i[1]:]
	}
	minStart := -1
	for _, m := range matches {
		if m.Start <= 0 {
			continue
		}
		switch m.Name {
		case "title", "website", "release_group", "language", "subtitle_language", "country":
			continue
		}
		if m.Name == "season_episode_pair" {
			raw := strings.ToLower(m.Raw)
			hasX := strings.Contains(raw, "x")
			hasSeasonEpisodeLetters := strings.Contains(raw, "s") && strings.Contains(raw, "e")
			if !hasX && !hasSeasonEpisodeLetters {
				continue
			}
		}
		if m.Name == "season_episode_pair_compact" {
			if len(m.Raw) <= 3 && shouldIgnoreCompactBoundary(input, m) {
				continue
			}
		}
		if minStart < 0 || m.Start < minStart {
			minStart = m.Start
		}
	}
	if minStart > 0 && minStart <= len(base) {
		base = base[:minStart]
	}
	if i := reTitleSeasonEpisodeBoundary.FindStringIndex(base); len(i) == 2 {
		base = base[:i[0]]
	}
	if i := reTitleYearBoundary.FindStringIndex(base); len(i) == 2 {
		base = base[:i[0]]
	}
	if i := strings.Index(base, "("); i > 0 {
		base = base[:i]
	}
	if strings.HasPrefix(base, "[") {
		if end := strings.Index(base, "]"); end >= 0 && end+1 < len(base) {
			base = strings.TrimLeft(base[end+1:], " ._-")
		}
	}
	if m := reTitleArticlePrefix.FindStringSubmatch(base); len(m) == 3 {
		base = m[2]
	}
	if m := reTitleShortScenePrefix.FindStringSubmatch(base); len(m) == 3 {
		prefix := m[1]
		if prefix == strings.ToLower(prefix) && prefix != "the" && prefix != "and" && strings.ContainsAny(m[2], "._") &&
			reEpisodeLike.MatchString(input) {
			base = m[2]
		}
	}
	base = cleanup(base)
	base = strings.TrimSpace(base)
	if base != "" {
		result.set("title", base)
	}
}

func shouldIgnoreCompactBoundary(input string, m match) bool {
	if len(m.Raw) != 3 || m.Start <= 0 || m.Start > len(input) || m.Raw[0] != '1' {
		return false
	}
	prefix := strings.TrimRightFunc(input[:m.Start], isSeparator)
	if prefix == "" {
		return false
	}
	lastSep := -1
	for i := len(prefix) - 1; i >= 0; i-- {
		if isSeparator(rune(prefix[i])) {
			lastSep = i
			break
		}
	}
	lastToken := strings.ToLower(prefix[lastSep+1:])
	return lastToken == "the"
}

func deriveEpisodeTitle(input string, result parsedResult) {
	if _, ok := result.get("episode"); !ok {
		return
	}
	re := regexp.MustCompile(`(?i)(?:S\d{1,2}[ ._-]?E\d{1,3}|\d{1,2}x\d{1,3})[ ._-]+(.+)`)
	m := re.FindStringSubmatch(input)
	if len(m) != 2 {
		return
	}
	title := cleanup(m[1])
	title = regexp.MustCompile(`(?i)[ ._-](hdtv|webrip|web[-_. ]?dl|bluray|bdrip|dvdrip|x264|x265|h\.?264|h\.?265|aac|ac3|dts).*$`).ReplaceAllString(title, "")
	title = regexp.MustCompile(`(?i)[ ._-](mkv|mp4|avi|srt|nfo)$`).ReplaceAllString(title, "")
	title = strings.TrimSuffix(title, ".")
	if title != "" {
		result.set("episode_title", title)
	}
}

func containsAnyHint(input string, hints []string) bool {
	for _, hint := range hints {
		if strings.Contains(input, hint) {
			return true
		}
	}
	return false
}

func cleanup(input string) string {
	trimmed := strings.TrimFunc(input, isSeparator)
	trimmed = strings.ReplaceAll(trimmed, "_", " ")
	trimmed = strings.ReplaceAll(trimmed, ".", " ")
	return strings.Join(strings.Fields(trimmed), " ")
}

func isSeparator(r rune) bool {
	return r == '.' || r == '_' || r == '-' || unicode.IsSpace(r)
}
