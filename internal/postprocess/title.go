package postprocess

import (
	"path/filepath"
	"regexp"
	"strings"
)

var (
	reTitleSeasonEpisodeBoundary = regexp.MustCompile(`(?i)[ ._-]s\d{1,3}[ ._-]?e\d{1,3}\b`)
	reTitleLeadingSeasonEpisode = regexp.MustCompile(`(?i)^s\d{1,3}[ ._-]?e\d{1,3}\b[ ._-]*`)
	reTitleYearBoundary          = regexp.MustCompile(`(?i)[ ._-](19\d{2}|20[0-3]\d)\b`)
	reTitleArticlePrefix         = regexp.MustCompile(`^([a-z]{3,6})[._-]+((?:the|a|an)[._-].+)$`)
	reTitleShortScenePrefix      = regexp.MustCompile(`^([a-z0-9]{3,4})-(.+)$`)
)

func deriveTitle(input string, matches []Match, result Result, opts *Options) {
	if opts.ExpectedTitle != "" {
		result.Set("title", opts.ExpectedTitle)
		return
	}
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
			// Keep normal S01E02 / 2x03 as a title boundary, but ignore compact forms like "100".
			hasX := strings.Contains(raw, "x")
			hasSeasonEpisodeLetters := strings.Contains(raw, "s") && strings.Contains(raw, "e")
			if !hasX && !hasSeasonEpisodeLetters {
				continue
			}
		}
		if m.Name == "season_episode_pair_compact" {
			// Keep known title-number forms like "The 100" out of boundaries.
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
	// Drop leading bracketed site/tag prefixes: [TorrentCouch.com].Title...
	if strings.HasPrefix(base, "[") {
		if end := strings.Index(base, "]"); end >= 0 && end+1 < len(base) {
			base = strings.TrimLeft(base[end+1:], " ._-")
		}
	}
	// Drop scene-style lowercase prefix tags: rvkd-the.old.man -> the.old.man
	if m := reTitleArticlePrefix.FindStringSubmatch(base); len(m) == 3 {
		base = m[2]
	}
	// Drop short scene-style group prefixes: pfa-star.trek.picard -> star.trek.picard
	if m := reTitleShortScenePrefix.FindStringSubmatch(base); len(m) == 3 {
		prefix := m[1]
		if prefix == strings.ToLower(prefix) && prefix != "the" && prefix != "and" && strings.ContainsAny(m[2], "._") &&
			reEpisodeLike.MatchString(input) {
			base = m[2]
		}
	}
	base = Cleanup(base)
	base = strings.TrimSpace(base)
	if base != "" {
		result.Set("title", base)
	}
}

func shouldIgnoreCompactBoundary(input string, m Match) bool {
	if len(m.Raw) != 3 || m.Start <= 0 || m.Start > len(input) || m.Raw[0] != '1' {
		return false
	}
	prefix := strings.TrimRightFunc(input[:m.Start], IsSeparator)
	if prefix == "" {
		return false
	}
	lastSep := -1
	for i := len(prefix) - 1; i >= 0; i-- {
		if IsSeparator(rune(prefix[i])) {
			lastSep = i
			break
		}
	}
	lastToken := strings.ToLower(prefix[lastSep+1:])
	return lastToken == "the"
}
