package whatsthis

import (
	"reflect"
	"testing"
)

func TestNumberingDoesNotConsumeMovieMetadata(t *testing.T) {
	for _, input := range []string{
		"Copperlight (2017) - UpScaled - 720p - SCR-Rip - Hindi - x264 - AC3 - 5.1 - ExampleTeam - ExampleTV",
		"2031 (2011) 1080p BrRip x264 - 1.8GB - GROUP",
		"Example.2024.1080p.x264.AAC-2.3GB.mkv",
		"Example.2024.1920x1080.x264.mkv",
		"Example.2024.1080p.x264.AC3-S56",
		"Example.2024.[1080p].[x264].[E-AC3-S78].mkv",
		"Show.Episode.1.1999.mkv",
	} {
		t.Run(input, func(t *testing.T) {
			got := File(input)
			if got.Type != Movie || got.Season != 0 || got.Episode != 0 {
				t.Fatalf("metadata or a movie-title number treated as episode numbering: %+v", got)
			}
		})
	}
}

func TestMetadataWordsRemainInTitles(t *testing.T) {
	for input, title := range map[string]string{
		"The.French.Beacon.1080p.x264.mkv":     "The French Beacon",
		"The.Uncut.Story.1080p.x264.mkv":       "The Uncut Story",
		"La.French.Voyage.2017.1080p.x264.mkv": "La French Voyage",
		"37.Copper.Bridges.1080p.x264.mkv":     "37 Copper Bridges",
	} {
		t.Run(input, func(t *testing.T) {
			got := File(input)
			if got.Type != Movie || got.Title != title {
				t.Fatalf("got %+v, want movie title %q", got, title)
			}
		})
	}
}

func TestSeasonNumberingDisambiguation(t *testing.T) {
	tests := []struct {
		input   string
		typ     Type
		season  int
		episode int
	}{
		{"Show.S4-11.1080p.x264", Show, 0, 0},
		{"Show.S04-S11.1080p.x264", Show, 0, 0},
		{"Show.Season.4-11.1080p.x264", Show, 0, 0},
		{"Show.S01.720p.WEB.H.264-GROUP", ShowSeason, 1, 0},
		{"Show.S00E00.1080p.x264.mkv", Episode, 0, 0},
		{"[720p] Show Season 1", ShowSeason, 1, 0},
		{"Silver Compass 02 - 13 Season 2 Complete [720p]", Episode, 2, 13},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got := File(test.input)
			if got.Type != test.typ || got.Season != test.season || got.Episode != test.episode {
				t.Fatalf("got %+v, want %s S%d E%d", got, test.typ, test.season, test.episode)
			}
		})
	}
}

func TestSeasonPackBareResolutions(t *testing.T) {
	for _, size := range []string{"480", "540", "720", "1080", "2160"} {
		for _, marker := range []string{"S01.", "Season.1.", "S01.-.", "S01.(", "S01.["} {
			for _, prefix := range []string{"", "[GROUP] "} {
				input := prefix + "Show." + marker + size
				if marker == "S01.(" {
					input += ")"
				} else if marker == "S01.[" {
					input += "]"
				}
				input += ".WEB.x264"
				t.Run(input, func(t *testing.T) {
					got := File(input)
					if got.Type != ShowSeason || got.Title != "Show" || got.Season != 1 ||
						got.Episode != 0 || got.ScreenSize != size+"p" || got.VideoCodec != "H.264" {
						t.Fatalf("bare resolution treated as episode numbering: %+v", got)
					}
				})
			}
		}
	}

	for _, input := range []string{
		"Show.S01E720.WEB.x264",
		"Show.1x720.WEB.x264",
		"Show.Episode 1.720.WEB.x264",
	} {
		t.Run(input, func(t *testing.T) {
			got := File(input)
			if got.Type != Episode || got.Season != 1 || got.Episode != 720 {
				t.Fatalf("explicit episode number discarded: %+v", got)
			}
		})
	}
}

func TestMovieTitleWithLeadingMetadataBrackets(t *testing.T) {
	for _, test := range []struct {
		input              string
		metadataAfterTitle bool
	}{
		{"[1080p] Example.Movie.2024.x264.mkv", false},
		{"[GROUP] [1080p] [AAC] Example.Movie.2024.x264.mkv", false},
		{"[1080p] Example.Movie.[720p].MULTi.2024.x264.mkv", true},
	} {
		t.Run(test.input, func(t *testing.T) {
			got := File(test.input)
			if got.Type != Movie || got.Title != "Example Movie" || got.Year != 2024 || got.VideoCodec != "H.264" {
				t.Fatalf("leading metadata brackets consumed movie title: %+v", got)
			}
			if test.metadataAfterTitle && (!got.Multilingual || got.ScreenSize != "720p") {
				t.Fatalf("metadata bracket after title was ignored: %+v", got)
			}
		})
	}
}

func TestNumericSeriesTitleWithLeadingTags(t *testing.T) {
	input := "[GROUP] 1923.S02E03.1080p.x264-GROUP.mkv"
	want := Info{Type: Episode, Title: "1923", Season: 2, Episode: 3,
		ScreenSize: "1080p", VideoCodec: "H.264", ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska"}
	if got := File(input); !reflect.DeepEqual(got, want) {
		t.Fatalf("got: %+v\nwant: %+v", got, want)
	}
}

func TestVideoSeasonEpisodeShorthand(t *testing.T) {
	// File reserves the ambiguous S4-11 shape for a collection; Video can
	// interpret it as an episode when the caller knows it is a video name.
	got := Video("Show.S4-11.1080p.x264")
	if got.Type != Episode || got.Season != 4 || got.Episode != 11 {
		t.Fatalf("got %+v, want episode S4 E11", got)
	}
}

func TestTechnicalMetadataDoesNotBecomeNumbering(t *testing.T) {
	for _, test := range []struct {
		input string
		want  Info
	}{
		{
			"Movie.2024.1080p.x264.[S01-S04]",
			Info{Type: Movie, Title: "Movie", Year: 2024, ScreenSize: "1080p", VideoCodec: "H.264", ReleaseGroup: "S01-S04"},
		},
		{
			"Movie.2024.1080p.x264-E03.mkv",
			Info{Type: Movie, Title: "Movie", Year: 2024, ScreenSize: "1080p", VideoCodec: "H.264", ReleaseGroup: "E03", Container: "mkv", MimeType: "video/x-matroska"},
		},
		{
			"[GROUP] Movie.1080p.x264.720.mkv",
			Info{Type: Movie, Title: "Movie", ScreenSize: "720p", VideoCodec: "H.264", Container: "mkv", MimeType: "video/x-matroska"},
		},
		{
			"Movie.1080i-GROUP.mkv",
			Info{Type: Movie, Title: "Movie", ScreenSize: "1080i", ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska"},
		},
		{
			"Show.S01.1440.WEB.x264",
			Info{Type: ShowSeason, Title: "Show", Season: 1, ScreenSize: "1440p", VideoCodec: "H.264"},
		},
	} {
		t.Run(test.input, func(t *testing.T) {
			if got := File(test.input); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("got: %+v\nwant: %+v", got, test.want)
			}
		})
	}
}

func TestUnderscoreSeasonPacks(t *testing.T) {
	for _, marker := range []string{"S01", "Season_1", "1st_Season", "S01E01-08", "S01_Season_1"} {
		input := "Show_" + marker + "_1080p_x264-[my_group]"
		t.Run(input, func(t *testing.T) {
			want := Info{Type: ShowSeason, Title: "Show", Season: 1, ScreenSize: "1080p", VideoCodec: "H.264", ReleaseGroup: "my_group"}
			if got := File(input); !reflect.DeepEqual(got, want) {
				t.Fatalf("got: %+v\nwant: %+v", got, want)
			}
		})
	}
}

func FuzzFile(f *testing.F) {
	for _, input := range []string{
		"", "\xff", "Show。S01E01.1080p.x264-GROUP.mkv",
		"Show_S2025E1056_1080i_x264.mkv", "[Group] Show - 1001 [1080p].mkv",
		"Movie.2024.[1080p].[x264].[E-AC3-S78].mkv", "Show.S01-S04.1080p.x264",
		"Show_S00-S03_Season_0-3_1080p_x264", "Show.S07-S08.S02-S03",
		"Show.S999-S000", "Show.S000-S999", "[ENG] Show.S02E03.VOSTFR.WEB-DL.DL",
	} {
		f.Add(input)
	}
	f.Fuzz(func(t *testing.T, input string) {
		for _, got := range []Info{File(input), Video(input)} {
			if got.Season < 0 || got.Episode < 0 || got.Year < 0 {
				t.Fatalf("negative metadata: %+v", got)
			}
			if got.Type == Show {
				if len(got.Seasons) < 2 || got.Season != 0 || got.Episode != 0 {
					t.Fatalf("invalid collection metadata: %+v", got)
				}
				for i, season := range got.Seasons {
					if season < 0 || i > 0 && season <= got.Seasons[i-1] {
						t.Fatalf("invalid season list: %+v", got)
					}
				}
			} else if got.Seasons != nil {
				t.Fatalf("season list on a non-collection: %+v", got)
			}
			switch got.Type {
			case Movie, Episode, ShowSeason, Show:
			default:
				t.Fatalf("invalid media type: %+v", got)
			}
		}
	})
}
