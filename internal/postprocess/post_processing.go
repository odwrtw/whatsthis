package postprocess

import (
	"regexp"
	"strconv"
	"strings"
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
	releaseGroupTechHints     = []string{"web-dl", "webdl", "webrip", "hdtv", "bluray", "x264", "h264", "x265", "h265", "dd", "ddp", "1080p", "720p", "2160p"}
	releaseGroupTagHints      = []string{"web-dl", "webdl", "webrip", "hdtv", "bluray", "x264", "h264", "x265", "h265", "dd", "ddp", "aac", "ac3", "dts", "1080p", "720p", "2160p"}
	blockedReleaseGroup       = map[string]struct{}{
		"h264": {}, "x264": {}, "h265": {}, "x265": {}, "1080p": {}, "720p": {},
		"hdtv": {}, "webrip": {}, "webdl": {}, "bluray": {}, "dvdrip": {}, "proper": {},
	}
	reSeasonEpisodeSE = regexp.MustCompile(`(?i)^s(\d{1,3})[ ._-]?e(\d{1,3})$`)
	reSeasonEpisodeX  = regexp.MustCompile(`(?i)^s?(\d{1,3})x(?:e)?(\d{1,3})$`)
	reTokenSplit      = regexp.MustCompile(`[ ._]+`)
)

type Match struct {
	Name  string
	Value any
	Start int
	End   int
	Raw   string
	Tags  []string
}

type Result map[string]any

func (r Result) Get(name string) (any, bool) {
	v, ok := r[name]
	return v, ok
}

func (r Result) Set(name string, value any) {
	r[name] = value
}

type Options struct {
	TypeHint      string
	ExpectedTitle string
	ExpectedGroup string
	Includes      map[string]struct{}
	Excludes      map[string]struct{}
}

func propertyEnabled(name string, opts *Options) bool {
	if len(opts.Includes) > 0 {
		if _, ok := opts.Includes[name]; !ok {
			return false
		}
	}
	if _, denied := opts.Excludes[name]; denied {
		return false
	}
	return true
}

func containsAnyHint(input string, hints []string) bool {
	for _, hint := range hints {
		if strings.Contains(input, hint) {
			return true
		}
	}
	return false
}

func PostProcessResult(input string, matches []Match, result Result, opts *Options) {
	deriveSeasonEpisode(result)
	deriveProperCount(result)
	if opts.ExpectedGroup != "" {
		result.Set("release_group", opts.ExpectedGroup)
	}
	if propertyEnabled("release_group", opts) {
		deriveReleaseGroup(input, result)
	}
	if propertyEnabled("mimetype", opts) {
		deriveMIMEType(result)
	}
	if propertyEnabled("title", opts) {
		deriveTitle(input, matches, result, opts)
	}
	if propertyEnabled("episode_title", opts) {
		deriveEpisodeTitle(input, result)
	}
	if propertyEnabled("type", opts) {
		deriveType(input, result, opts)
	}
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

func deriveReleaseGroup(input string, result Result) {
	if _, exists := result.Get("release_group"); exists {
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
				result.Set("release_group", prefix)
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
	candidate = strings.TrimFunc(candidate, IsSeparator)
	candidate = refineReleaseGroupCandidate(candidate)
	if candidate == "" {
		return
	}
	if strings.Contains(candidate, "_") {
		if _, hasSeason := result.Get("season"); hasSeason {
			return
		}
		if _, hasEpisode := result.Get("episode"); hasEpisode {
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
	result.Set("release_group", candidate)
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

func deriveSeasonEpisode(result Result) {
	pair, ok := result.Get("season_episode_pair")
	if !ok {
		// Fallback to compact style parser only when explicit pair is not available.
		pair, ok = result.Get("season_episode_pair_compact")
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
	result.Set("season", season)
	result.Set("episode", episode)
}

func deriveProperCount(result Result) {
	v, ok := result.Get("other")
	if !ok {
		return
	}
	if s, ok := v.(string); ok && s == "Proper" {
		result.Set("proper_count", 1)
	}
}

func deriveMIMEType(result Result) {
	c, ok := result.Get("container")
	if !ok {
		return
	}
	switch strings.ToLower(c.(string)) {
	case "mkv":
		result.Set("mimetype", "video/x-matroska")
	case "mp4", "m4v":
		result.Set("mimetype", "video/mp4")
	case "avi":
		result.Set("mimetype", "video/x-msvideo")
	}
}

func deriveType(input string, result Result, opts *Options) {
	if opts.TypeHint != "" {
		result.Set("type", strings.ToLower(opts.TypeHint))
		return
	}
	if _, ok := result.Get("season"); ok {
		result.Set("type", "episode")
		return
	}
	if _, ok := result.Get("episode"); ok {
		result.Set("type", "episode")
		return
	}
	if reTypeEpisode.MatchString(input) {
		result.Set("type", "episode")
		return
	}
	if reTypeSeasonToken.MatchString(input) {
		result.Set("type", "episode")
		return
	}
	if reTypeTSeasonToken.MatchString(input) && reTypeEpisodeHints.MatchString(input) {
		result.Set("type", "episode")
		return
	}
	if reTypeCompact3Digits.MatchString(input) && !reTypeYear.MatchString(input) {
		result.Set("type", "episode")
		return
	}
	result.Set("type", "movie")
}
