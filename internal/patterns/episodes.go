package patterns

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	reEpisodeOfSeparator = regexp.MustCompile(`(?i)[ ._-]?of[ ._-]?`)
)

func registerEpisodePatterns(m *Matcher) {
	reNum := regexp.MustCompile(`\d+`)
	intFromRaw := func(raw string) any {
		n := reNum.FindString(raw)
		if n == "" {
			return raw
		}
		v, err := strconv.Atoi(n)
		if err != nil {
			return raw
		}
		return v
	}

	AddPatternSpec(m, "season_episode_pair", `(?i)\bS(\d{1,3})[ ._-]?E(\d{1,3})\b`, PatternSpec{})
	AddPatternSpec(m, "season_episode_pair", `\b(\d{1,2})x(\d{1,3})\b`, PatternSpec{})
	AddPatternSpec(m, "season_episode_pair", `(?i)\bS(\d{1,3})xE?(\d{1,3})\b`, PatternSpec{})

	// e.g. "E23", "Ep 23", "Episode 23"
	AddPatternSpec(m, "episode", `(?i)\b(?:e|ep|epi|episode|cap)[ ._-]?\d{1,3}\b`, PatternSpec{
		Formatter: intFromRaw,
	})

	// e.g. "Season 6", "Saison 12", "Temporada 1", "Stagione 6", "S03"
	AddPatternSpec(m, "season", `(?i)\b(?:season|saison|temporada|stagione|s)[ ._-]?\d{1,3}\b`, PatternSpec{
		Formatter: intFromRaw,
	})

	// e.g. "1of4", "3 of 9" => episode 1 / 3
	AddPatternSpec(m, "episode", `(?i)\b\d{1,3}[ ._-]?of[ ._-]?\d{1,3}\b`, PatternSpec{
		Formatter: func(raw string) any {
			parts := reEpisodeOfSeparator.Split(raw, 2)
			if len(parts) != 2 {
				return intFromRaw(raw)
			}
			v, convErr := strconv.Atoi(reNum.FindString(parts[0]))
			if convErr != nil {
				return intFromRaw(raw)
			}
			return v
		},
	})

	// compact style e.g. "117" => S01E17, "208" => S02E08
	AddPatternSpec(m, "season_episode_pair_compact", `\b\d{3}\b`, PatternSpec{
		Validator: func(input string, start, end int) bool {
			raw := input[start:end]
			// avoid matching obvious years (4-digit, kept for safety)
			if len(raw) == 4 && (strings.HasPrefix(raw, "19") || strings.HasPrefix(raw, "20")) {
				return false
			}
			return true
		},
		Formatter: func(raw string) any {
			if raw[1:] == "00" {
				return raw
			}
			return raw[0:1] + "x" + raw[1:]
		},
	})
	AddPatternSpec(m, "season_episode_pair_compact", `\b\d{4}\b`, PatternSpec{
		Validator: func(input string, start, end int) bool {
			raw := input[start:end]
			// Avoid obvious years; keep compact season/episode like 2619.
			if strings.HasPrefix(raw, "19") || strings.HasPrefix(raw, "20") {
				return false
			}
			return true
		},
		Formatter: func(raw string) any {
			return raw[0:2] + "x" + raw[2:]
		},
	})
}
