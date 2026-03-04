package guessit

import (
	"encoding/json"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type outputRow struct {
	FileName string         `json:"filename"`
	Expected map[string]any `json:"expected"`
}

type fieldMap struct {
	name    string
	guessit string
	ours    string
}

type mismatchRow struct {
	fileName string
	expected string
	ours     string
}

func TestOutputJSONDiffSummary(t *testing.T) {
	content, err := os.ReadFile("output.json")
	if err != nil {
		t.Skip("output.json not available")
	}
	var rows []outputRow
	if err := json.Unmarshal(content, &rows); err != nil {
		t.Fatalf("invalid output.json: %v", err)
	}

	maps := []fieldMap{
		{"title", "title", "title"},
		{"season", "season", "season"},
		{"episode", "episode", "episode"},
		{"year", "year", "year"},
		{"quality", "screen_size", "screen_size"},
		{"video_codec", "video_codec", "video_codec"},
		{"audio_codec", "audio_codec", "audio_codec"},
		{"release_group", "release_group", "release_group"},
		{"container", "container", "container"},
		{"mimetype", "mimetype", "mimetype"},
		{"type", "type", "type"},
	}

	matchOursVsGuessit := map[string]int{}
	availableGuessit := map[string]int{}
	mismatches := map[string][]mismatchRow{}

	for _, row := range rows {
		if row.FileName == "" || row.Expected == nil {
			continue
		}
		ours := GuessIt(row.FileName)
		for _, m := range maps {
			gv, gok := row.Expected[m.guessit]
			ov, ook := ours[m.ours]
			ng := normalizeField(m.guessit, gv)
			no := normalizeField(m.guessit, ov)
			if gok && ng != "" {
				availableGuessit[m.name]++
			}
			if gok && ook && ng != "" && no != "" && ng == no {
				matchOursVsGuessit[m.name]++
				continue
			}
			if gok && ng != "" && no != "" && ng != no {
				mismatches[m.name] = append(mismatches[m.name], mismatchRow{
					fileName: row.FileName,
					expected: ng,
					ours:     no,
				})
			}
		}
	}

	t.Logf("rows=%d", len(rows))
	maxLogs := mismatchLogLimit()
	for _, key := range sortedFieldKeys(maps) {
		t.Logf("field=%s ours_vs_guessit_match=%d/%d (%.2f%%)",
			key,
			matchOursVsGuessit[key], availableGuessit[key],
			pct(matchOursVsGuessit[key], availableGuessit[key]),
		)
		for _, mm := range limitMismatches(mismatches[key], maxLogs) {
			t.Logf("mismatch field=%s filename=%q expected=%q ours=%q", key, mm.fileName, mm.expected, mm.ours)
		}
	}
}

func mismatchLogLimit() int {
	raw := strings.TrimSpace(os.Getenv("GO_GUESSIT_MAX_MISMATCH_LOG"))
	if raw == "" {
		return 20
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 20
	}
	return n
}

func limitMismatches(items []mismatchRow, max int) []mismatchRow {
	if max < 0 || len(items) <= max {
		return items
	}
	return items[:max]
}

func pct(n, d int) float64 {
	if d == 0 {
		return 0
	}
	return (float64(n) * 100) / float64(d)
}

func normalizeField(key string, v any) string {
	switch key {
	case "type":
		switch x := v.(type) {
		case bool:
			if x {
				return "movie"
			}
			return "episode"
		}
	}
	s := strings.ToLower(strings.TrimSpace(strings.NewReplacer(".", "", "-", "", "_", "", " ", "").Replace(toString(v))))
	switch key {
	case "year":
		if s == "0" {
			return ""
		}
	case "source":
		switch s {
		case "web", "webrip", "webdl":
			return "web"
		case "bluray", "bdrip":
			return "bluray"
		case "dvdrip", "dvd":
			return "dvd"
		case "digitaltv", "pdtv", "sdtv", "dvb":
			return "pdtv"
		}
	case "streaming_service":
		switch s {
		case "netflix", "nf":
			return "nf"
		case "amazonprime", "amazon", "amzn":
			return "amzn"
		case "itunes", "ip":
			return "ip"
		case "appletv+", "appletv", "atvp":
			return "atvp"
		case "disney+", "disneyplus", "dsnp", "dsny":
			return "dsnp"
		case "hbo", "hbomax", "hmax":
			return "hmax"
		case "showtime", "sho":
			return "sho"
		}
	case "audio_codec":
		if strings.Contains(s, "dd") || strings.Contains(s, "dolbydigital") || strings.Contains(s, "ac3") {
			return "dd"
		}
		if strings.Contains(s, "aac") {
			return "aac"
		}
		if strings.Contains(s, "dts") {
			return "dts"
		}
	case "video_codec":
		if s == "x264" || s == "h264" || s == "avc" {
			return "h264"
		}
		if s == "x265" || s == "h265" || s == "hevc" {
			return "h265"
		}
	}
	return s
}

func toString(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case float64:
		return strings.TrimSuffix(strings.TrimSuffix(strings.ToLower(strings.TrimSpace(toJSONNumber(x))), ".0"), ".")
	default:
		b, _ := json.Marshal(x)
		return string(b)
	}
}

func toJSONNumber(f float64) string {
	b, _ := json.Marshal(f)
	return string(b)
}

func sortedFieldKeys(maps []fieldMap) []string {
	seen := map[string]struct{}{}
	keys := make([]string, 0, len(maps))
	for _, m := range maps {
		if _, ok := seen[m.name]; ok {
			continue
		}
		seen[m.name] = struct{}{}
		keys = append(keys, m.name)
	}
	sort.Strings(keys)
	return keys
}

func TestGuessItStripsLeadingSeasonEpisodeFromTitle(t *testing.T) {
	got := GuessIt("s01e01 Les petits fantômes .mp4")
	title, ok := got["title"].(string)
	if !ok {
		t.Fatalf("missing title in result: %#v", got)
	}
	if title != "Les petits fantômes" {
		t.Fatalf("unexpected title: got %q want %q", title, "Les petits fantômes")
	}
}

func TestGuessItDerivesReleaseGroupFromWebDLTail(t *testing.T) {
	got := GuessIt("Twin.Peaks.S01E01.Pilot.720p.WEB-DL.x264.POOP.mp4")
	group, ok := got["release_group"].(string)
	if !ok {
		t.Fatalf("missing release_group in result: %#v", got)
	}
	if group != "POOP" {
		t.Fatalf("unexpected release_group: got %q want %q", group, "POOP")
	}
}
