package guessit

// GuessIt parses a media filename and returns extracted metadata.
func GuessIt(input string) Result {
	parsed := parsedResult{}
	var matches []match
	set := func(name string, value any, start, end int, raw string) {
		setParsedField(parsed, name, value)
		if start >= 0 && end > start {
			matches = append(matches, match{
				Name:  name,
				Value: value,
				Start: start,
				End:   end,
				Raw:   raw,
			})
		}
	}
	runSequentialParsers(input, set)
	postProcessResult(input, matches, parsed)
	return resultFromParsed(parsed)
}

func setParsedField(parsed parsedResult, name string, value any) {
	switch name {
	case "type", "title", "screen_size", "release_group", "audio_codec", "video_codec", "container", "mimetype":
		if s, ok := value.(string); ok {
			parsed[name] = s
		}
	case "episode", "season", "year":
		if n, ok := asInt(value); ok {
			parsed[name] = n
		}
	default:
		parsed[name] = value
	}
}

func resultFromParsed(parsed parsedResult) Result {
	return Result{
		Type:         stringValue(parsed["type"]),
		Title:        stringValue(parsed["title"]),
		Episode:      intValue(parsed["episode"]),
		Season:       intValue(parsed["season"]),
		Year:         intValue(parsed["year"]),
		ScreenSize:   stringValue(parsed["screen_size"]),
		ReleaseGroup: stringValue(parsed["release_group"]),
		AudioCodec:   stringValue(parsed["audio_codec"]),
		VideoCodec:   stringValue(parsed["video_codec"]),
		Container:    stringValue(parsed["container"]),
		MIMEType:     stringValue(parsed["mimetype"]),
	}
}

func stringValue(v any) string {
	s, _ := v.(string)
	return s
}

func intValue(v any) int {
	n, _ := asInt(v)
	return n
}

func asInt(v any) (int, bool) {
	switch x := v.(type) {
	case int:
		return x, true
	case int64:
		return int(x), true
	case float64:
		return int(x), true
	default:
		return 0, false
	}
}
