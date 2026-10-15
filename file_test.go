package whatsthis

import (
	"reflect"
	"testing"
)

func TestFileWholeSeasons(t *testing.T) {
	tests := []struct {
		name     string
		expected Info
	}{
		{
			name: "Velvet Meridian S01 Season 1 2026 1080p WEBRip DD5.1 10bits x265-ALPHA",
			expected: Info{Type: ShowSeason, Title: "Velvet Meridian", Season: 1, Year: 2026, ScreenSize: "1080p",
				ReleaseGroup: "ALPHA", AudioCodec: "Dolby Digital", VideoCodec: "H.265"},
		},
		{
			name:     "VELVET MERIDIAN S01 - BETA",
			expected: Info{Type: ShowSeason, Title: "VELVET MERIDIAN", Season: 1, ReleaseGroup: "BETA"},
		},
		{
			name: "VELVET MERIDIAN 2026 S01 1080p WEB-DL HEVC x265 5.1 GAMMA",
			expected: Info{Type: ShowSeason, Title: "VELVET MERIDIAN", Season: 1, Year: 2026, ScreenSize: "1080p",
				ReleaseGroup: "GAMMA", VideoCodec: "H.265"},
		},
		{
			name: "VELVET MERIDIAN (2026) S01 (1080p AMZN WEB-DL x265 10bit EAC3 5.1 EncoderDelta) [DELTA]",
			expected: Info{Type: ShowSeason, Title: "VELVET MERIDIAN", Season: 1, Year: 2026, ScreenSize: "1080p",
				ReleaseGroup: "DELTA", AudioCodec: "Dolby Digital Plus", VideoCodec: "H.265"},
		},
		{
			name: "Velvet Meridian - Season 1 - Mp4 x264 AC3 1080p",
			expected: Info{Type: ShowSeason, Title: "Velvet Meridian", Season: 1, ScreenSize: "1080p",
				AudioCodec: "Dolby Digital", VideoCodec: "H.264"},
		},
		{
			name: "VELVET.MERIDIAN.S01.1080p.WEBRip.x265-EPSILON",
			expected: Info{Type: ShowSeason, Title: "VELVET MERIDIAN", Season: 1, ScreenSize: "1080p",
				ReleaseGroup: "EPSILON", VideoCodec: "H.265"},
		},
		{
			name: "VELVET.MERIDIAN.COMPLETE.Season.2.S02.720p.WEB-DL.x265-KAPPA",
			expected: Info{Type: ShowSeason, Title: "VELVET MERIDIAN", Season: 2, ScreenSize: "720p",
				ReleaseGroup: "KAPPA", VideoCodec: "H.265"},
		},
		{
			name: "VELVET.MERIDIAN.S01.1080p.WEBRip.x265.AAC-ZETA7",
			expected: Info{Type: ShowSeason, Title: "VELVET MERIDIAN", Season: 1, ScreenSize: "1080p",
				ReleaseGroup: "ZETA7", AudioCodec: "AAC", VideoCodec: "H.265"},
		},
		{
			name: "VELVET MERIDIAN (2026) Season 1 S01 (2160p STREAM WEB-DL x265 HEVC 10bit DDP 5.1 EncoderEta)",
			expected: Info{Type: ShowSeason, Title: "VELVET MERIDIAN", Season: 1, Year: 2026, ScreenSize: "2160p",
				ReleaseGroup: "EncoderEta", AudioCodec: "Dolby Digital Plus", VideoCodec: "H.265"},
		},
		{
			name: "Velvet Meridian 2026 Season 1 Complete 1080p WEB x264 [theta_group]",
			expected: Info{Type: ShowSeason, Title: "Velvet Meridian", Season: 1, Year: 2026, ScreenSize: "1080p",
				ReleaseGroup: "theta_group", VideoCodec: "H.264"},
		},
		{
			name: "Velvet Meridian 2026 Season 1 Complete 720p WEB x264 [theta_group]",
			expected: Info{Type: ShowSeason, Title: "Velvet Meridian", Season: 1, Year: 2026, ScreenSize: "720p",
				ReleaseGroup: "theta_group", VideoCodec: "H.264"},
		},
		{
			name: "Velvet Meridian - Season 01 (2026)[COMPLETE] 1080p DLMux x265 AAC[IOTA8]",
			expected: Info{Type: ShowSeason, Title: "Velvet Meridian", Season: 1, Year: 2026, ScreenSize: "1080p",
				ReleaseGroup: "IOTA8", AudioCodec: "AAC", VideoCodec: "H.265"},
		},
		{
			name: "13 Lanterns S02 COMPLETE 720p WEBRip x264 [4.9GB] [MP4] [Season 2 Full]",
			expected: Info{Type: ShowSeason, Title: "13 Lanterns", Season: 2, ScreenSize: "720p",
				VideoCodec: "H.264"},
		},
		{
			name: "VELVET.MERIDIAN.S01E01-08.1080p.WEB-DL.DDP5.1.H.265",
			expected: Info{Type: ShowSeason, Title: "VELVET MERIDIAN", Season: 1, ScreenSize: "1080p",
				AudioCodec: "Dolby Digital Plus", VideoCodec: "H.265"},
		},
		{
			name: "VELVET.MERIDIAN.S01E01-08.1080p.WEB-DL.DDP5.1.H.264",
			expected: Info{Type: ShowSeason, Title: "VELVET MERIDIAN", Season: 1, ScreenSize: "1080p",
				AudioCodec: "Dolby Digital Plus", VideoCodec: "H.264"},
		},
		{
			name:     "Show S01 [1080p]",
			expected: Info{Type: ShowSeason, Title: "Show", Season: 1, ScreenSize: "1080p"},
		},
		{
			name:     "Show S01 [x265]",
			expected: Info{Type: ShowSeason, Title: "Show", Season: 1, VideoCodec: "H.265"},
		},
		{
			name:     "Show S01 - [720p] [AAC]",
			expected: Info{Type: ShowSeason, Title: "Show", Season: 1, ScreenSize: "720p", AudioCodec: "AAC"},
		},
		{
			name:     "Show S01 [DTS]",
			expected: Info{Type: ShowSeason, Title: "Show", Season: 1, AudioCodec: "DTS"},
		},
		{
			name:     "Show S01 [GROUP] [1080p] [x265]",
			expected: Info{Type: ShowSeason, Title: "Show", Season: 1, ReleaseGroup: "GROUP", ScreenSize: "1080p", VideoCodec: "H.265"},
		},
		{
			name:     "Show.S01.1080p.WEBRip.x265",
			expected: Info{Type: ShowSeason, Title: "Show", Season: 1, ScreenSize: "1080p", VideoCodec: "H.265"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := File(test.name); !reflect.DeepEqual(got, test.expected) {
				t.Fatalf("got:      %+v\nexpected: %+v", got, test.expected)
			}
		})
	}
}

func TestFileMultiSeasonSeries(t *testing.T) {
	tests := []struct {
		name  string
		input string
		title string
	}{
		{
			name:  "Example Show seasons 1 through 21",
			input: "Example.Show.Uncensored.S01-S21.720p.BluRay.x265-ExampleTeam",
			title: "Example Show Uncensored",
		},
		{
			name:  "Season word range",
			input: "Example.Show.Season.1-4.S01-S04.1080p.WEB-DL.x265-GROUP",
			title: "Example Show",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := File(test.input)
			if got.Type != Show || got.Title != test.title || got.Season != 0 || got.Episode != 0 {
				t.Fatalf("File(%q) = %+v, want multi-season series %q", test.input, got, test.title)
			}
		})
	}
}

func TestFileRepeatedSeasonRanges(t *testing.T) {
	for _, markers := range []string{
		"Season.1-4.S01-S04",
		"S01-S04.Season.1-4",
		"Season.1-4.Season.1-4",
	} {
		for _, group := range []string{"", "GROUP"} {
			input := "Example.Show." + markers + ".1080p.x265"
			if group != "" {
				input += "-" + group
			}
			t.Run(input, func(t *testing.T) {
				want := Info{
					Type: Show, Title: "Example Show", ScreenSize: "1080p",
					VideoCodec: "H.265", ReleaseGroup: group,
				}
				if got := File(input); !reflect.DeepEqual(got, want) {
					t.Fatalf("File(%q) = %+v, want %+v", input, got, want)
				}
			})
		}
	}
}

func TestFileSeasonPackWithWebsiteSuffix(t *testing.T) {
	name := "Example.Show.s01.HD720p.WEBRip.Rus.ExampleSite.tv"
	got := File(name)
	if got.Type != ShowSeason || got.Title != "Example Show" || got.Season != 1 {
		t.Fatalf("File(%q) = %+v, want a season 1 pack for Example Show", name, got)
	}
}

func TestFileFallsBackToVideo(t *testing.T) {
	tests := []string{
		"Big.Buck.Bunny.2008.1080p.BluRay.x264-YIFY",
		"VELVET.MERIDIAN.S01E01.1080p.WEB-DL.H.265",
		"VELVET.MERIDIAN.S01E02-03.1080p.WEB-DL.H.265",
		"VELVET.MERIDIAN.S01E01.1080p.WEB-DL.H.265.mkv",
		"Show S01.txt",
		"Show.S01.nfo",
		"Show S01.mkv",
	}

	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			if got, expected := File(name), Video(name); !reflect.DeepEqual(got, expected) {
				t.Fatalf("got:      %+v\nexpected: %+v", got, expected)
			}
			if got := File(name); got.Type == ShowSeason {
				t.Fatalf("filename classified as a whole season: %+v", got)
			}
		})
	}
}
