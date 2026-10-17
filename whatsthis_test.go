package whatsthis

import (
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"testing"

	"golang.org/x/text/language"
)

type entry struct {
	FileName string `json:"filename"`
	Expected Info   `json:"expected"`
	Parser   string `json:"parser,omitempty"`
}

func testGuessFromFile(t *testing.T, filename, parser string) {
	t.Helper()

	parsers := map[string]func(string) Info{"video": Video, "file": File}
	parse, ok := parsers[parser]
	if !ok {
		t.Fatalf("unsupported parser %q", parser)
	}

	content, err := os.ReadFile(filename) //nolint:gosec // test data paths are not user-controlled
	if err != nil {
		t.Fatalf("read required fixture %s: %v", filename, err)
	}

	var entries []entry
	if err := json.Unmarshal(content, &entries); err != nil {
		t.Fatalf("invalid %s: %+v", filename, err)
	}

	for _, e := range entries {
		if e.Parser != "" {
			if _, ok := parsers[e.Parser]; !ok {
				t.Fatalf("unsupported parser %q for %s", e.Parser, e.FileName)
			}
			if e.Parser != parser {
				continue
			}
		}
		t.Run(e.FileName, func(t *testing.T) {
			if got := parse(e.FileName); !reflect.DeepEqual(got, e.Expected) {
				t.Errorf("got:      %+v\nexpected: %+v", got, e.Expected)
			}
		})
	}
}

func TestGuessMovies(t *testing.T) {
	t.Run("Video", func(t *testing.T) {
		testGuessFromFile(t, "testdata/movies-testdata.json", "video")
	})
	t.Run("File", func(t *testing.T) {
		testGuessFromFile(t, "testdata/movies-testdata.json", "file")
	})
}

func TestGuessTvShows(t *testing.T) {
	t.Run("Video", func(t *testing.T) {
		testGuessFromFile(t, "testdata/tvshows-testdata.json", "video")
	})
	t.Run("File", func(t *testing.T) {
		testGuessFromFile(t, "testdata/tvshows-testdata.json", "file")
	})
}

func TestGuessUnderscoreDelimitedEpisode(t *testing.T) {
	name := "Journey_to_the_Stars_DVDrip_S04E08_An_Evening_With_Alex.mp4"
	assertVideoInfo(t, name, Info{Type: Episode, Title: "Journey to the Stars", Season: 4, Episode: 8,
		Container: "mp4", MimeType: "video/mp4"})
}

func TestGuessEpisodeTitleBoundary(t *testing.T) {
	tests := []struct {
		input string
		title string
	}{
		{"Example.Show.(US)S01E01.1080p.x264-GROUP.mkv", "Example Show (US)"},
		{"Show!S01E01.1080p.x264-GROUP.mkv", "Show!"},
		{"Show。S01E01.1080p.x264-GROUP.mkv", "Show。"},
		{"S01E01.Show.1080p.x264-GROUP.mkv", "Show"},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			want := Info{
				Type: Episode, Title: test.title, Season: 1, Episode: 1,
				ScreenSize: "1080p", VideoCodec: "H.264", ReleaseGroup: "GROUP",
				Container: "mkv", MimeType: "video/x-matroska",
			}
			if got := File(test.input); !reflect.DeepEqual(got, want) {
				t.Fatalf("File(%q) = %+v, want %+v", test.input, got, want)
			}
		})
	}
}

func TestGuessReleaseGroupBeforeMetadataBrackets(t *testing.T) {
	for _, prefix := range []string{"Example.2024", "Example.S01E01"} {
		for _, suffix := range []string{
			".1080p.x264-GROUP.[ENG]",
			".x264-GROUP.[1080p].[ENG]",
			".[GROUP].[1080p].[x264].[ENG]",
			".[1080p].[x264].[ENG]",
		} {
			input := prefix + suffix + ".mkv"
			t.Run(input, func(t *testing.T) {
				want := Info{
					Type: Movie, Title: "Example", Year: 2024,
					ScreenSize: "1080p", VideoCodec: "H.264", Languages: []language.Tag{language.English},
					ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska",
				}
				if prefix == "Example.S01E01" {
					want.Type, want.Year, want.Season, want.Episode = Episode, 0, 1, 1
				}
				if suffix == ".[1080p].[x264].[ENG]" {
					want.ReleaseGroup = ""
				}
				if got := File(input); !reflect.DeepEqual(got, want) {
					t.Fatalf("File(%q) = %+v, want %+v", input, got, want)
				}
			})
		}
	}
}

func TestGuessTorrentWithWebsitePrefix(t *testing.T) {
	name := "[EXAMPLE-TORRENTS.PL] Example.S04E05.MULTi.1080p.AMZN.WEB-DL.H264.DDP5.1.Atmos-G42"
	want := Info{
		Type: Episode, Title: "Example", Season: 4, Episode: 5, ScreenSize: "1080p",
		ReleaseGroup: "G42", AudioCodec: "Dolby Digital Plus", VideoCodec: "H.264", Multilingual: true,
	}
	if got := File(name); !reflect.DeepEqual(got, want) {
		t.Fatalf("File(%q) = %+v, want %+v", name, got, want)
	}
}

func TestGuessOpusAudioCodec(t *testing.T) {
	name := "Example.Adventures.S01E48.FiNAL.VOSTFR.1080p.Bluray.AV1.OPUS-ExampleGroup.mkv"
	got := File(name)
	if got.AudioCodec != "Opus" {
		t.Fatalf("File(%q) = %+v, want Opus audio", name, got)
	}
}

func TestGuessTrueHDAudioCodec(t *testing.T) {
	name := "Explorers of the Cosmos 2014 1080p BluRay HDR10 TrueHD Atmos 7 1 x265-ExampleGroup.mkv"
	got := File(name)
	if got.AudioCodec != "Dolby TrueHD" {
		t.Fatalf("File(%q) = %+v, want Dolby TrueHD audio", name, got)
	}
}

func TestGuessFLACAudioCodec(t *testing.T) {
	name := "Echoes.of.the.Past.1963.LABEL.1080p.BluRay.x265.HEVC.FLAC-GROUP.mkv"
	got := File(name)
	if got.AudioCodec != "FLAC" {
		t.Fatalf("File(%q) = %+v, want FLAC audio", name, got)
	}
}

func TestGuessAV1Codec(t *testing.T) {
	name := "A.Traveler.of.the.Distant.Lands.S01E03.1080p.AV1.10bit-ExampleGroup.mkv"
	got := File(name)
	if got.VideoCodec != "AV1" {
		t.Fatalf("File(%q) = %+v, want AV1", name, got)
	}
}

func TestGuessAVCCodecLabel(t *testing.T) {
	name := "Example.S01E10.The.Path.of.Hope.WEB-DLRip-AVC.Example.Studio.-enc0der3.mkv"
	got := File(name)
	if got.VideoCodec != "H.264" {
		t.Fatalf("File(%q) = %+v, want H.264", name, got)
	}
}

func TestGuessVC1Codec(t *testing.T) {
	name := "Alex and Casey S02E05 New Horizons 1080p TrueHD 5 1 VC-1 REMUX-ExAmPlE.mkv"
	got := File(name)
	if got.VideoCodec != "VC-1" {
		t.Fatalf("File(%q) = %+v, want VC-1", name, got)
	}
}

func TestGuessVP9Codec(t *testing.T) {
	name := "Example.S01E01.The.Next.Adventure.720p.RED.WEB-DL.AAC5.1.VP9-GRP42.mkv"
	got := File(name)
	if got.VideoCodec != "VP9" {
		t.Fatalf("File(%q) = %+v, want VP9", name, got)
	}
}

func TestGuessUnderscoreCodecTag(t *testing.T) {
	name := "[A.B.C.alphaSub&Beta&Gamma-Studio] Example's Grand Adventure - Silver Travelers [01][Ma10p_1080p][x265_flac].mkv"
	assertVideoInfo(t, name, Info{Type: Episode, Title: "Example's Grand Adventure Silver Travelers", Episode: 1,
		ScreenSize: "1080p", VideoCodec: "H.265", AudioCodec: "FLAC", ReleaseGroup: "A.B.C.alphaSub&Beta&Gamma-Studio",
		Container: "mkv", MimeType: "video/x-matroska"})
}

func TestGuessWebDownloadRipMetadata(t *testing.T) {
	name := "Example.S01E01.WEB-DLRip.x264.exampleGroup.mkv"
	want := Info{
		Type: Episode, Title: "Example", Season: 1, Episode: 1,
		ReleaseGroup: "exampleGroup", VideoCodec: "H.264",
		Container: "mkv", MimeType: "video/x-matroska",
	}
	if got := File(name); !reflect.DeepEqual(got, want) {
		t.Fatalf("File(%q) = %+v, want %+v", name, got, want)
	}
}

func TestGuessPixelDimensionsResolution(t *testing.T) {
	name := "[ExampleGroup]_The_Journey_of_Alex_Morgan_(1920x1080_Blu-ray_FLAC)_[123ABC45].mkv"
	assertVideoInfo(t, name, Info{Type: Movie, Title: "The Journey of Alex Morgan", ScreenSize: "1080p", AudioCodec: "FLAC",
		ReleaseGroup: "ExampleGroup", Container: "mkv", MimeType: "video/x-matroska"})
}

func TestGuess4KResolution(t *testing.T) {
	name := "Distant Worlds II S01E01 Shores 4K HDR [BDRip by ExampleGroup].mkv"
	got := File(name)
	if got.Type != Episode || got.Season != 1 || got.Episode != 1 || got.ScreenSize != "2160p" {
		t.Fatalf("File(%q) = %+v, want episode S01E01 at 2160p", name, got)
	}
}

func TestGuess1440pResolution(t *testing.T) {
	name := "THE.REMARKABLE.SILVER.CARAVAN.S01E06.1440p.WEB.AAC.2.0.H.265-GROUP.mkv"
	got := File(name)
	if got.ScreenSize != "1440p" {
		t.Fatalf("File(%q) = %+v, want 1440p", name, got)
	}
}

func TestGuessHDResolution(t *testing.T) {
	name := "Alex.in.Wondervale.2020.S01E01.HD1080P.X264.AAC.Japanese.CHS.GRP4.mp4"
	got := File(name)
	if got.Type != Episode || got.Season != 1 || got.Episode != 1 || got.ScreenSize != "1080p" {
		t.Fatalf("File(%q) = %+v, want episode S01E01 at 1080p", name, got)
	}
}

func TestGuessBareResolution(t *testing.T) {
	name := "example.adventures.s02e01.hdtv.720.rus.eng.example.tv.mkv"
	got := File(name)
	if got.Type != Episode || got.Season != 2 || got.Episode != 1 || got.ScreenSize != "720p" {
		t.Fatalf("File(%q) = %+v, want episode S02E01 at 720p", name, got)
	}
}

func TestGuessSpaceSeparatedEpisodeMetadata(t *testing.T) {
	name := "42 Lanterns - S07E01 - 1080p WEBRip AMZN [EAC3 5.1 x264-GRp].mkv"
	want := Info{
		Type: Episode, Title: "42 Lanterns", Season: 7, Episode: 1,
		ScreenSize: "1080p", ReleaseGroup: "GRp",
		AudioCodec: "Dolby Digital Plus", VideoCodec: "H.264",
		Container: "mkv", MimeType: "video/x-matroska",
	}
	if got := File(name); !reflect.DeepEqual(got, want) {
		t.Fatalf("File(%q) = %+v, want %+v", name, got, want)
	}
}

func TestGuessLanguageMetadata(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		languages    []language.Tag
		multilingual bool
	}{
		{
			name:         "episode tags",
			input:        "The.Travelers.S05E06.Beyond.the.horizon.ITA.ENG.1080p.AMZN.WEB-DL.DDP5.1.H.264-GRP.EX.mkv",
			languages:    []language.Tag{language.Italian, language.English},
			multilingual: true,
		},
		{
			name:         "movie tags",
			input:        "Example.Movie.2024.ITA.ENG.1080p.WEB-DL.H.264-GROUP.mkv",
			languages:    []language.Tag{language.Italian, language.English},
			multilingual: true,
		},
		{
			name:         "whole season with generic marker",
			input:        "Example.Show.S01.MULTi.ITA.ENG.1080p.WEB-DL.H.264-GROUP",
			languages:    []language.Tag{language.Italian, language.English},
			multilingual: true,
		},
		{
			name:         "generic marker without language tags",
			input:        "Example.Show.S01E01.MULTi.1080p.WEB-DL.mkv",
			multilingual: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := File(test.input)
			if !slices.Equal(got.Languages, test.languages) || got.Multilingual != test.multilingual {
				t.Fatalf("File(%q) languages = %v, multilingual = %t; want %v, %t",
					test.input, got.Languages, got.Multilingual, test.languages, test.multilingual)
			}
		})
	}
}
