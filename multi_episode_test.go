package whatsthis

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestMultiEpisodes(t *testing.T) {
	for _, test := range []struct {
		marker   string
		season   int
		episodes []int
	}{
		// Common renamer multi-episode styles.
		{"S01E02E03E04", 1, []int{2, 3, 4}},
		{"S01E02-E03-E04", 1, []int{2, 3, 4}},
		{"S01E02-E04", 1, []int{2, 3, 4}},
		{"S01E02-04", 1, []int{2, 3, 4}},
		{"S01E02-03-04", 1, []int{2, 3, 4}},
		{"S01E02.S01E03", 1, []int{2, 3}},
		{"S01E02S01E03", 1, []int{2, 3}},
		{"S01E02-S01E04", 1, []int{2, 3, 4}},
		{"S01E02 - E04", 1, []int{2, 3, 4}},
		{"s01e02-e03", 1, []int{2, 3}},
		{"S01E02E04", 1, []int{2, 4}},
		{"S01E02-E04E06", 1, []int{2, 3, 4, 6}},
		{"S01E02E02E03", 1, []int{2, 3}},
		{"S00E00-E01", 0, []int{0, 1}},
		{"1x02x03", 1, []int{2, 3}},
		{"1X02X03X04", 1, []int{2, 3, 4}},
		{"1x02-04", 1, []int{2, 3, 4}},
		{"1x02-1x04", 1, []int{2, 3, 4}},
	} {
		t.Run(test.marker, func(t *testing.T) {
			assertVideoInfo(t, "Example."+test.marker+".1080p.x264.AAC-GROUP.mkv", Info{
				Type: Episode, Title: "Example", Season: test.season, Episode: test.episodes[0], Episodes: test.episodes,
				ScreenSize: "1080p", VideoCodec: "H.264", AudioCodec: "AAC", ReleaseGroup: "GROUP",
				Container: "mkv", MimeType: "video/x-matroska",
			})
		})
	}
	// Underscores are separators, and the first named episode stays in Episode.
	assertVideoInfo(t, "Example_S01E04E02_720p.mkv", Info{Type: Episode, Title: "Example", Season: 1, Episode: 4,
		Episodes: []int{2, 4}, ScreenSize: "720p", Container: "mkv", MimeType: "video/x-matroska"})
}

func TestMultiEpisodeFalsePositives(t *testing.T) {
	for _, marker := range []string{"S01E02", "1x02"} {
		for _, suffix := range []string{
			"", " - Part 2", " - Episode 3", " - The Return of E03", " - Chapter 1x03", "-#slipperyslope",
			" - 24 Hours", " - 03 Days", "-03Days", " - 2.0", " - 567 (1080p)",
			".720p", "-1080p", "-720", " - 2024", "-2024", "-1920x1080",
			"-5.1", "-2.0", "-1.8GB", "-2.3 GB",
			".x264", "-x265", "x264", ".H.265", ".AAC2.0", ".DDP5.1", ".EAC3", "-E-AC3",
			"a", "b", "v2", ".1080p.x264-E03", ".E03",
			// Other seasons and descending ranges are not single-season lists.
			"-S02E01", "-2x01", "-E01", "-01", "-E1000",
		} {
			if marker == "1x02" && suffix == "-S02E01" {
				continue // An SxxEyy marker takes precedence over NxMM.
			}
			input := "Example." + marker + suffix + ".1080p.x264.mkv"
			t.Run(input, func(t *testing.T) {
				for name, parse := range map[string]func(string) Info{"File": File, "Video": Video} {
					got := parse(input)
					if got.Episodes != nil || got.Type != Episode || got.Season != 1 || got.Episode != 2 {
						t.Fatalf("%s(%q) = %+v; want single episode S01E02", name, input, got)
					}
				}
			})
		}
	}
	for _, input := range []string{
		"Example.Ep02Ep03.mkv", "Example.Episodes 02 & 04.mkv", "Example.S01E02 1x03.mkv",
		"Example.Season 1 Episode 02-Episode 04.mkv", "Movie.2024.1080p.x264-E03-E04.mkv",
	} {
		t.Run(input, func(t *testing.T) {
			for name, parse := range map[string]func(string) Info{"File": File, "Video": Video} {
				if got := parse(input); got.Episodes != nil {
					t.Fatalf("%s(%q) = %+v; not an explicit single-season list", name, input, got)
				}
			}
		})
	}
}

func TestEpisodeLikeCreditsAfterTechnicalMetadata(t *testing.T) {
	for _, metadata := range []string{"720", "720p", "x264", "H.264", "DDP5.1", "720.x264"} {
		for _, credit := range []string{"-S01E02", ".[S01E02]", "-(S01E02)"} {
			input := "Movie.2024." + metadata + credit + ".mkv"
			for name, parse := range map[string]func(string) Info{"File": File, "Video": Video} {
				t.Run(name+"/"+input, func(t *testing.T) {
					got := parse(input)
					if got.Type != Movie || got.Title != "Movie" || got.Year != 2024 ||
						credit == "-S01E02" && got.ReleaseGroup != "S01E02" {
						t.Fatalf("episode-like credit became numbering: %+v", got)
					}
				})
			}
		}
	}
}

func TestMultiEpisodeSeasonPack(t *testing.T) {
	eight := []int{1, 2, 3, 4, 5, 6, 7, 8}
	for _, marker := range []string{"S01E01-08", "S01E01-E08", "S01E01 - E08"} {
		input := "Example." + marker + ".1080p.x264-GROUP"
		t.Run(input, func(t *testing.T) {
			want := Info{Type: ShowSeason, Title: "Example", Season: 1, Episodes: eight,
				ScreenSize: "1080p", VideoCodec: "H.264", ReleaseGroup: "GROUP"}
			if got := File(input); !reflect.DeepEqual(got, want) {
				t.Fatalf("File(%q) = %+v; want %+v", input, got, want)
			}
			want.Type, want.Episode = Episode, 1
			if got := Video(input); !reflect.DeepEqual(got, want) {
				t.Fatalf("Video(%q) = %+v; want %+v", input, got, want)
			}
		})
	}
	// Only a valid single-season range from E01 is a season directory.
	for _, input := range []string{
		"Example.S01E02-03.1080p.x264-GROUP", "Example.S01E01E02.1080p.x264-GROUP",
		"Example.S01E01-08S02E01.1080p.x264-GROUP", "Example.S01E01-08-S02E01.1080p.x264-GROUP",
		"Example.S01E01-720.1080p.x264-GROUP", "Example.S01E01 - 03 Days.1080p.x264-GROUP",
	} {
		t.Run(input, func(t *testing.T) {
			if got, want := File(input), Video(input); !reflect.DeepEqual(got, want) || got.Type != Episode {
				t.Fatalf("not a single-season directory: File=%+v Video=%+v", got, want)
			}
		})
	}
	for _, input := range []string{"Example.S01.1080p.x264", "Example.Season.1.Complete.1080p.x264", "Example.S01-S03.1080p.x264"} {
		if got := File(input); got.Episodes != nil {
			t.Fatalf("season label invents episodes: %s %+v", input, got)
		}
	}
}

func TestMultiEpisodeJSON(t *testing.T) {
	for _, test := range []struct {
		input    string
		episodes string
	}{
		{"Example.S01E02-E03.mkv", "[2,3]"},
		{"Example.S01E01-03.1080p.x264", "[1,2,3]"},
		{"Example.S01E02.mkv", ""},
		{"Example.S01.1080p.x264", ""},
		{"Movie.2024.mkv", ""},
	} {
		t.Run(test.input, func(t *testing.T) {
			encoded, err := json.Marshal(File(test.input))
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(encoded, &fields); err != nil {
				t.Fatal(err)
			}
			if string(fields["episodes"]) != test.episodes {
				t.Fatalf("JSON = %s; want episodes %q", encoded, test.episodes)
			}
		})
	}
}
