package guessit

import (
	"regexp"
	"strconv"
	"strings"
)

type parserSetter func(name string, value any, start, end int, raw string)

type parserRule struct {
	re    *regexp.Regexp
	value func(string) any
}

func runSequentialParsers(input string, set parserSetter) {
	parseContainer(input, set)
	parseSource(input, set)
	parseVideoCodec(input, set)
	parseAudioCodec(input, set)
	parseScreenSize(input, set)
	parseEdition(input, set)
	parseStreamingService(input, set)
	parseCRC32(input, set)
	parseSize(input, set)
	parseBitRate(input, set)
	parseCD(input, set)
	parsePart(input, set)
	parseDate(input, set)
	parseWebsite(input, set)
	parseLanguage(input, set)
	parseCountry(input, set)
	parseOther(input, set)
	parseBonus(input, set)
	parseFilm(input, set)
	parseEpisode(input, set)
	parseYear(input, set)
}

func parseContainer(input string, set parserSetter) {
	applyFirst(input, "container", set, []parserRule{{
		re: regexp.MustCompile(`(?i)\.(mkv|avi|mp4|m4v|mov|wmv|flv|mpg|mpeg|ts|m2ts|ogm|webm)\b`),
		value: func(raw string) any {
			return strings.ToLower(strings.TrimPrefix(raw, "."))
		},
	}})
}

func parseSource(input string, set parserSetter) {
	applyFirst(input, "source", set, []parserRule{
		{re: regexp.MustCompile(`(?i)\b(?:blu[ .-]?ray|bdrip|bdremux)\b`), value: constValue("Blu-ray")},
		{re: regexp.MustCompile(`(?i)\b(?:hddvd|hd[ .-]?dvd)\b`), value: constValue("HD-DVD")},
		{re: regexp.MustCompile(`(?i)\b(?:web[ ._-]?dl|web[ ._-]?rip|webrip)\b`), value: constValue("Web")},
		{re: regexp.MustCompile(`(?i)\b(?:hdtv)\b`), value: constValue("HDTV")},
		{re: regexp.MustCompile(`(?i)\b(?:dvdrip|dvd)\b`), value: constValue("DVD")},
		{re: regexp.MustCompile(`(?i)\b(?:dvb|pdtv|sdtv)\b`), value: constValue("Digital TV")},
		{re: regexp.MustCompile(`(?i)\b(?:vhs)\b`), value: constValue("VHS")},
		{re: regexp.MustCompile(`(?i)\b(?:cam)\b`), value: constValue("CAM")},
		{re: regexp.MustCompile(`(?i)\b(?:hd[ .-]?tv)\b`), value: constValue("HDTV")},
	})
}

func parseVideoCodec(input string, set parserSetter) {
	applyFirst(input, "video_codec", set, []parserRule{
		{re: regexp.MustCompile(`(?i)\b(?:x[ ._-]?264|h[ ._-]?264|avc)\b`), value: constValue("H.264")},
		{re: regexp.MustCompile(`(?i)\b(?:x[ ._-]?265|h[ ._-]?265|hevc)\b`), value: constValue("H.265")},
		{re: regexp.MustCompile(`(?i)\b(?:xvid|x\.vid)\b`), value: constValue("Xvid")},
		{re: regexp.MustCompile(`(?i)\b(?:divx)\b`), value: constValue("DivX")},
		{re: regexp.MustCompile(`(?i)\b(?:vp9)\b`), value: constValue("VP9")},
		{re: regexp.MustCompile(`(?i)\b(?:av1)\b`), value: constValue("AV1")},
		{re: regexp.MustCompile(`(?i)\b(?:mpeg-?2)\b`), value: constValue("MPEG-2")},
	})
}

func parseAudioCodec(input string, set parserSetter) {
	applyFirst(input, "audio_codec", set, []parserRule{
		{re: regexp.MustCompile(`(?i)\b(?:dts[ .-]?hd)\b`), value: constValue("DTS-HD")},
		{re: regexp.MustCompile(`(?i)\b(?:dts)\b`), value: constValue("DTS")},
		{re: regexp.MustCompile(`(?i)\b(?:true[ .-]?hd)\b`), value: constValue("TrueHD")},
		{re: regexp.MustCompile(`(?i)\b(?:e-?ac-?3|ac-?3|dd(?:p)?(?:[ ._-]?\d(?:[ ._-]?\d)?)?)\b`), value: constValue("Dolby Digital")},
		{re: regexp.MustCompile(`(?i)\b(?:aac(?:[ ._-]?\d\.\d)?)\b`), value: constValue("AAC")},
		{re: regexp.MustCompile(`(?i)\b(?:flac)\b`), value: constValue("FLAC")},
		{re: regexp.MustCompile(`(?i)\b(?:mp3)\b`), value: constValue("MP3")},
		{re: regexp.MustCompile(`(?i)\b(?:opus)\b`), value: constValue("Opus")},
	})
	applyFirst(input, "audio_channels", set, []parserRule{{
		re:    regexp.MustCompile(`\b(?:7\.1|5\.1|2\.1|2\.0|1\.0)\b`),
		value: identityValue,
	}})
}

func parseScreenSize(input string, set parserSetter) {
	applyFirst(input, "screen_size", set, []parserRule{
		{re: regexp.MustCompile(`(?i)\b360p\b`), value: constValue("360p")},
		{re: regexp.MustCompile(`(?i)\b480p\b`), value: constValue("480p")},
		{re: regexp.MustCompile(`(?i)\b540p\b`), value: constValue("540p")},
		{re: regexp.MustCompile(`(?i)\b576p\b`), value: constValue("576p")},
		{re: regexp.MustCompile(`(?i)\b720p\b`), value: constValue("720p")},
		{re: regexp.MustCompile(`(?i)\b1080p\b`), value: constValue("1080p")},
		{re: regexp.MustCompile(`(?i)\b1080i\b`), value: constValue("1080i")},
		{re: regexp.MustCompile(`(?i)\b2160p\b`), value: constValue("2160p")},
		{re: regexp.MustCompile(`(?i)\b4k\b`), value: constValue("2160p")},
		{re: regexp.MustCompile(`(?i)\b8k\b`), value: constValue("4320p")},
	})
}

func parseEdition(input string, set parserSetter) {
	applyFirst(input, "edition", set, []parserRule{
		{re: regexp.MustCompile(`(?i)\bdirector'?s?[ ._-]?cut\b`), value: constValue("Director's Cut")},
		{re: regexp.MustCompile(`(?i)\bdc\b`), value: constValue("Director's Cut")},
		{re: regexp.MustCompile(`(?i)\balternat(?:e|ive)[ ._-]?cut\b`), value: constValue("Alternative Cut")},
		{re: regexp.MustCompile(`(?i)\bextended(?:[ ._-]?cut)?\b`), value: constValue("Extended")},
		{re: regexp.MustCompile(`(?i)\buncensored\b`), value: constValue("Uncensored")},
		{re: regexp.MustCompile(`(?i)\bunrated\b`), value: constValue("Unrated")},
		{re: regexp.MustCompile(`(?i)\btheatrical(?:[ ._-]?cut)?\b`), value: constValue("Theatrical")},
		{re: regexp.MustCompile(`(?i)\bspecial[ ._-]?edition\b`), value: constValue("Special")},
		{re: regexp.MustCompile(`(?i)\bedition[ ._-]?collector\b`), value: constValue("Collector")},
		{re: regexp.MustCompile(`(?i)\bcollector'?s?(?:[ ._-]?edition)?\b`), value: constValue("Collector")},
		{re: regexp.MustCompile(`(?i)criterion(?:[ ._-]?(?:collection|edition))?`), value: constValue("Criterion")},
		{re: regexp.MustCompile(`(?i)\blimited\b`), value: constValue("Limited")},
	})
}

func parseStreamingService(input string, set parserSetter) {
	applyFirst(input, "streaming_service", set, []parserRule{
		{re: regexp.MustCompile(`(?i)\bnetflix\b`), value: constValue("Netflix")},
		{re: regexp.MustCompile(`(?i)\bnetflixuhd\b`), value: constValue("Netflix")},
		{re: regexp.MustCompile(`(?i)\bnf\b`), value: constValue("Netflix")},
		{re: regexp.MustCompile(`(?i)\bamazon[ ._-]?prime\b`), value: constValue("Amazon Prime")},
		{re: regexp.MustCompile(`(?i)\bamazon\b`), value: constValue("Amazon Prime")},
		{re: regexp.MustCompile(`(?i)\bamzn\b`), value: constValue("Amazon Prime")},
		{re: regexp.MustCompile(`(?i)\bhulu\b`), value: constValue("Hulu")},
		{re: regexp.MustCompile(`(?i)\bitunes(?:hd)?\b`), value: constValue("iTunes")},
		{re: regexp.MustCompile(`(?i)\bhditunes\b`), value: constValue("iTunes")},
		{re: regexp.MustCompile(`(?i)(?:^|[ ._-])ae(?:[ ._-]|$)`), value: constValue("A&E")},
		{re: regexp.MustCompile(`(?i)\bdisney[ ._-]?plus\b`), value: constValue("Disney+")},
		{re: regexp.MustCompile(`(?i)\bhbo(?:[ ._-]?max)?\b`), value: constValue("HBO")},
		{re: regexp.MustCompile(`(?i)\bapple[ ._-]?tv(?:[ ._-]?plus)?\b`), value: constValue("Apple TV+")},
		{re: regexp.MustCompile(`(?i)\bparamount[ ._-]?plus\b`), value: constValue("Paramount+")},
		{re: regexp.MustCompile(`(?i)\batvp\b`), value: constValue("Apple TV+")},
		{re: regexp.MustCompile(`(?i)\bbbc\b`), value: constValue("BBC")},
		{re: regexp.MustCompile(`(?i)\bcbs\b`), value: constValue("CBS")},
		{re: regexp.MustCompile(`(?i)\babc\b`), value: constValue("ABC")},
		{re: regexp.MustCompile(`(?i)\bdsnp\b|\bdsny\b`), value: constValue("Disney+")},
		{re: regexp.MustCompile(`(?i)\bfox\b`), value: constValue("FOX")},
		{re: regexp.MustCompile(`(?i)\bhmax\b`), value: constValue("HBO")},
		{re: regexp.MustCompile(`(?i)\bip\b`), value: constValue("iTunes")},
		{re: regexp.MustCompile(`(?i)\bmtv\b`), value: constValue("MTV")},
		{re: regexp.MustCompile(`(?i)\bnick\b`), value: constValue("Nick")},
		{re: regexp.MustCompile(`(?i)\bsho\b`), value: constValue("Showtime")},
	})
}

func parseCRC32(input string, set parserSetter) {
	applyFirst(input, "crc32", set, []parserRule{{
		re: regexp.MustCompile(`(?i)[\[\(]([0-9a-f]{8})[\]\)]`),
		value: func(raw string) any {
			raw = strings.TrimPrefix(raw, "[")
			raw = strings.TrimPrefix(raw, "(")
			raw = strings.TrimSuffix(raw, "]")
			raw = strings.TrimSuffix(raw, ")")
			return strings.ToUpper(raw)
		},
	}})
}

func parseSize(input string, set parserSetter) {
	reSpace := regexp.MustCompile(`\s+`)
	applyFirst(input, "size", set, []parserRule{{
		re: regexp.MustCompile(`(?i)\b\d+(?:\.\d+)?[ ]?(?:kb|mb|gb|tb)\b`),
		value: func(raw string) any {
			raw = strings.ToUpper(strings.TrimSpace(raw))
			raw = reSpace.ReplaceAllString(raw, "")
			return raw
		},
	}})
}

func parseBitRate(input string, set parserSetter) {
	reSpace := regexp.MustCompile(`\s+`)
	applyFirst(input, "bit_rate", set, []parserRule{{
		re: regexp.MustCompile(`(?i)\b\d+(?:\.\d+)?[ ]?(?:k|m)(?:b(?:it)?s?)?(?:ps)?\b`),
		value: func(raw string) any {
			raw = strings.ToLower(strings.TrimSpace(raw))
			raw = reSpace.ReplaceAllString(raw, "")
			raw = strings.ReplaceAll(raw, "bits", "b")
			raw = strings.ReplaceAll(raw, "bit", "b")
			if strings.HasSuffix(raw, "b") {
				raw += "ps"
			}
			return strings.ToUpper(raw)
		},
	}})
}

func parseCD(input string, set parserSetter) {
	reNum := regexp.MustCompile(`\d+`)
	parseFirstInt := func(raw string) any {
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
	reOf := regexp.MustCompile(`(?i)(\d+)[ ._-]?of[ ._-]?(\d+)`)
	applyFirst(input, "cd", set, []parserRule{
		{re: regexp.MustCompile(`(?i)\b(?:cd|disc)[ ._-]?\d{1,2}\b`), value: parseFirstInt},
		{re: regexp.MustCompile(`(?i)\bcd[ ._-]?\d{1,2}[ ._-]?of[ ._-]?\d{1,2}\b`), value: parseFirstInt},
	})
	applyFirst(input, "cd_count", set, []parserRule{{
		re: regexp.MustCompile(`(?i)\b(?:cd[ ._-]?\d{1,2}[ ._-]?of[ ._-]?\d{1,2}|\d{1,2}[ ._-]?cds?)\b`),
		value: func(raw string) any {
			raw = strings.ToLower(raw)
			if strings.Contains(raw, "of") {
				m := reOf.FindStringSubmatch(raw)
				if len(m) == 3 {
					if v, convErr := strconv.Atoi(m[2]); convErr == nil {
						return v
					}
				}
			}
			n := reNum.FindString(raw)
			if n == "" {
				return raw
			}
			v, convErr := strconv.Atoi(n)
			if convErr != nil {
				return raw
			}
			return v
		},
	}})
}

func parsePart(input string, set parserSetter) {
	reNum := regexp.MustCompile(`\d+`)
	applyFirst(input, "part", set, []parserRule{{
		re: regexp.MustCompile(`(?i)\b(?:part|pt)[ ._-]?\d{1,2}\b`),
		value: func(raw string) any {
			n := reNum.FindString(raw)
			if n == "" {
				return raw
			}
			v, convErr := strconv.Atoi(n)
			if convErr != nil {
				return raw
			}
			return v
		},
	}})
}

func parseDate(input string, set parserSetter) {
	applyFirst(input, "date", set, []parserRule{
		{re: regexp.MustCompile(`\b(20\d{2})[.\-](\d{2})[.\-](\d{2})\b`), value: func(raw string) any {
			return strings.ReplaceAll(raw, ".", "-")
		}},
		{re: regexp.MustCompile(`\b(\d{2})[.\-](\d{2})[.\-](20\d{2})\b`), value: func(raw string) any {
			if len(raw) != 10 {
				return raw
			}
			return raw[6:10] + "-" + raw[3:5] + "-" + raw[0:2]
		}},
	})
}

func parseWebsite(input string, set parserSetter) {
	reWebsiteEpisodeLike := regexp.MustCompile(`(?i)\bs\d{1,3}(?:[ ._-]?e|x(?:e)?)[ ._-]?\d{1,3}\b|\b\d{1,2}x\d{1,3}\b`)
	applyFirst(input, "website", set, []parserRule{{
		re: regexp.MustCompile(`(?i)\b(?:www\.)?[a-z0-9-]+(?:\.[a-z0-9-]+)*\.(?:com|org|net|io|ru|to)\b`),
		value: func(raw string) any {
			return strings.ToLower(raw)
		},
	}}, func(start, end int, raw string) bool {
		return !reWebsiteEpisodeLike.MatchString(input[start:end])
	})
}

func parseLanguage(input string, set parserSetter) {
	applyFirst(input, "language", set, []parserRule{
		{re: regexp.MustCompile(`(?i)\b(english|eng|en)\b`), value: constValue("English")},
		{re: regexp.MustCompile(`(?i)\b(french|francais|fra|fr)\b`), value: constValue("French")},
		{re: regexp.MustCompile(`(?i)\b(spanish|esp|es)\b`), value: constValue("Spanish")},
		{re: regexp.MustCompile(`(?i)\b(german|deu|de)\b`), value: constValue("German")},
		{re: regexp.MustCompile(`(?i)\bcatalan\b`), value: constValue("Catalan")},
	})
	applyFirst(input, "subtitle_language", set, []parserRule{{
		re: regexp.MustCompile(`(?i)\b(?:sub|subs|subtitle|vostfr)[ ._-]?(en|eng|english|fr|fre?nch|spanish|german|it|ita)\b`),
		value: func(raw string) any {
			rawLower := strings.ToLower(raw)
			switch {
			case strings.Contains(rawLower, "eng") || strings.Contains(rawLower, " en"):
				return "English"
			case strings.Contains(rawLower, "fre") || strings.Contains(rawLower, "fr") || strings.Contains(rawLower, "vostfr"):
				return "French"
			case strings.Contains(rawLower, "span"):
				return "Spanish"
			case strings.Contains(rawLower, "ger"):
				return "German"
			case strings.Contains(rawLower, "ita") || strings.Contains(rawLower, "it"):
				return "it"
			default:
				return raw
			}
		},
	}})
}

func parseCountry(input string, set parserSetter) {
	applyFirst(input, "country", set, []parserRule{
		{re: regexp.MustCompile(`(?i)\bUS\b|\bUSA\b`), value: constValue("US")},
		{re: regexp.MustCompile(`(?i)\bUK\b`), value: constValue("UK")},
		{re: regexp.MustCompile(`(?i)\bFR\b|\bFrance\b`), value: constValue("FR")},
		{re: regexp.MustCompile(`(?i)\bCA\b|\bCanada\b|\bQu(?:e|\x{00E9})bec\b`), value: constValue("CA")},
	})
}

func parseOther(input string, set parserSetter) {
	applyFirst(input, "other", set, []parserRule{
		{re: regexp.MustCompile(`(?i)\bproper\b`), value: constValue("Proper")},
		{re: regexp.MustCompile(`(?i)\brepack\b`), value: constValue("Proper")},
		{re: regexp.MustCompile(`(?i)\brerip\b`), value: constValue("Proper")},
		{re: regexp.MustCompile(`(?i)\brip\b`), value: constValue("Rip")},
		{re: regexp.MustCompile(`(?i)\bre-?encoded\b|\breenc\b`), value: constValue("Reencoded")},
		{re: regexp.MustCompile(`(?i)\bcomplete\b`), value: constValue("Complete")},
		{re: regexp.MustCompile(`(?i)\bdual[ ._-]?audio\b`), value: constValue("Dual Audio")},
	})
}

func parseBonus(input string, set parserSetter) {
	applyFirst(input, "bonus", set, []parserRule{{
		re: regexp.MustCompile(`(?i)\bx(\d{1,2})\b`),
		value: func(raw string) any {
			if len(raw) < 2 {
				return raw
			}
			v, convErr := strconv.Atoi(raw[1:])
			if convErr != nil {
				return raw
			}
			return v
		},
	}})
}

func parseFilm(input string, set parserSetter) {
	applyFirst(input, "film", set, []parserRule{{
		re: regexp.MustCompile(`(?i)\bf(\d{1,2})\b`),
		value: func(raw string) any {
			if len(raw) < 2 {
				return raw
			}
			v, convErr := strconv.Atoi(raw[1:])
			if convErr != nil {
				return raw
			}
			return v
		},
	}})
}

func parseEpisode(input string, set parserSetter) {
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

	applyFirst(input, "season_episode_pair", set, []parserRule{
		{re: regexp.MustCompile(`(?i)\bS(\d{1,3})[ ._-]?E(\d{1,3})\b`), value: identityValue},
		{re: regexp.MustCompile(`\b(\d{1,2})x(\d{1,3})\b`), value: identityValue},
		{re: regexp.MustCompile(`(?i)\bS(\d{1,3})xE?(\d{1,3})\b`), value: identityValue},
	})

	applyFirst(input, "episode", set, []parserRule{
		{re: regexp.MustCompile(`(?i)\b(?:e|ep|epi|episode|cap)[ ._-]?\d{1,3}\b`), value: intFromRaw},
		{re: regexp.MustCompile(`(?i)\b\d{1,3}[ ._-]?of[ ._-]?\d{1,3}\b`), value: func(raw string) any {
			parts := regexp.MustCompile(`(?i)[ ._-]?of[ ._-]?`).Split(raw, 2)
			if len(parts) != 2 {
				return intFromRaw(raw)
			}
			v, convErr := strconv.Atoi(reNum.FindString(parts[0]))
			if convErr != nil {
				return intFromRaw(raw)
			}
			return v
		}},
	})

	applyFirst(input, "season", set, []parserRule{{
		re:    regexp.MustCompile(`(?i)\b(?:season|saison|temporada|stagione|s)[ ._-]?\d{1,3}\b`),
		value: intFromRaw,
	}})

	applyFirst(input, "season_episode_pair_compact", set, []parserRule{
		{re: regexp.MustCompile(`\b\d{3}\b`), value: func(raw string) any {
			if raw[1:] == "00" {
				return raw
			}
			return raw[0:1] + "x" + raw[1:]
		},
		},
		{re: regexp.MustCompile(`\b\d{4}\b`), value: func(raw string) any {
			return raw[0:2] + "x" + raw[2:]
		}},
	}, func(start, end int, raw string) bool {
		if len(raw) == 4 && (strings.HasPrefix(raw, "19") || strings.HasPrefix(raw, "20")) {
			return false
		}
		return true
	})
}

func parseYear(input string, set parserSetter) {
	applyFirst(input, "year", set, []parserRule{{
		re: regexp.MustCompile(`\b(19\d{2}|20[0-3]\d)\b`),
		value: func(raw string) any {
			v, convErr := strconv.Atoi(raw)
			if convErr != nil {
				return raw
			}
			return v
		},
	}})
}

func applyFirst(input, name string, set parserSetter, rules []parserRule, validator ...func(start, end int, raw string) bool) {
	for _, rule := range rules {
		idx := rule.re.FindStringIndex(input)
		if len(idx) != 2 {
			continue
		}
		raw := input[idx[0]:idx[1]]
		if len(validator) > 0 && validator[0] != nil && !validator[0](idx[0], idx[1], raw) {
			continue
		}
		value := any(raw)
		if rule.value != nil {
			value = rule.value(raw)
		}
		set(name, value, idx[0], idx[1], raw)
		return
	}
}

func constValue(v any) func(string) any {
	return func(string) any { return v }
}

func identityValue(raw string) any {
	return raw
}
