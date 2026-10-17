package whatsthis

import (
	"reflect"
	"testing"

	"golang.org/x/text/language"
)

func TestMetadataAliasAudioCodecs(t *testing.T) {
	for _, test := range []struct {
		metadata string
		want     string
	}{
		{"DDP", "Dolby Digital Plus"},
		{"DD+", "Dolby Digital Plus"},
		{"DDP5", "Dolby Digital Plus"},
		{"DD+5", "Dolby Digital Plus"},
		{"DDP2", "Dolby Digital Plus"},
		{"DD+2", "Dolby Digital Plus"},
		{"DDP5.1", "Dolby Digital Plus"},
		{"DD+5.1", "Dolby Digital Plus"},
		{"DDP 5 1", "Dolby Digital Plus"},
		{"DD+ 5.1", "Dolby Digital Plus"},
		{"[ddp5]", "Dolby Digital Plus"},
		{"WEB_DDP5_x264", "Dolby Digital Plus"},
		{"QAACx2", "AAC"},
		{"qaacx2", "AAC"},
		{"[QAACx2]", "AAC"},
		{"QAAC", "AAC"},
		{"AAC2.0", "AAC"},
		{"AAC5.1", "AAC"},
		{"DDPlex", ""},
		{"DDPlus", ""},
		{"ADDP5", ""},
		{"DDP5Team", ""},
		{"ADD+5", ""},
		{"DD+5Team", ""},
		{"DD+Team", ""},
		{"TeamQAACx2", ""},
		{"QAACx2Team", ""},
		{"QAACtion", ""},
		{"AACademy", ""},
		{"MPA2.0", ""},
	} {
		t.Run(test.metadata, func(t *testing.T) {
			if got := extractAudioCodec(test.metadata); got != test.want {
				t.Fatalf("got %q, want %q", got, test.want)
			}
		})
	}
}

func TestMetadataAliasScreenSizes(t *testing.T) {
	for _, test := range []struct {
		metadata string
		want     string
	}{
		{"m720p", "720p"},
		{"[M720P]", "720p"},
		{"720pBrRip", "720p"},
		{"720PBRRIP", "720p"},
		{"m720p.1080p", "1080p"},
		{"1080p.720pBrRip", "720p"},
		{"am720p", ""},
		{"m720plus", ""},
		{"m720pTeam", ""},
		{"720pBrRipper", ""},
		{"1720pBrRip", ""},
	} {
		t.Run(test.metadata, func(t *testing.T) {
			if got := extractScreenSize(test.metadata); got != test.want {
				t.Fatalf("got %q, want %q", got, test.want)
			}
		})
	}
}

func TestMetadataAliasesFullInfo(t *testing.T) {
	for _, alias := range []struct {
		metadata string
		audio    string
		lang     language.Tag
	}{
		{"Persian.m720p.WEB-DL.x264.DDP5", "Dolby Digital Plus", language.Persian},
		{"Persian.720pBrRip.x264.DD+", "Dolby Digital Plus", language.Persian},
		{"ESubs.m720p.x264.QAACx2", "AAC", language.English},
		{"E-Subs.m720p.x264.QAACx2", "AAC", language.English},
		{"ESubs.720pBrRip.x264.DDP", "Dolby Digital Plus", language.English},
		{"720pBrRip.x264.DD+5.ESubs", "Dolby Digital Plus", language.English},
	} {
		for _, media := range []struct {
			prefix string
			want   Info
		}{
			{"Example.2024.", Info{Type: Movie, Title: "Example", Year: 2024}},
			{"Example.S04E12.", Info{Type: Episode, Title: "Example", Season: 4, Episode: 12}},
			{"Example.S04.", Info{Type: ShowSeason, Title: "Example", Season: 4}},
			{"Example.S04-S06.", Info{Type: Show, Title: "Example", Seasons: []int{4, 5, 6}}},
		} {
			input := media.prefix + alias.metadata + "-GROUP"
			want := media.want
			want.ScreenSize, want.VideoCodec, want.AudioCodec = "720p", "H.264", alias.audio
			want.ReleaseGroup, want.Languages = "GROUP", []language.Tag{alias.lang}
			parsers := map[string]func(string) Info{"File": File}
			if want.Type == Movie || want.Type == Episode {
				input += ".mkv"
				want.Container, want.MimeType = "mkv", "video/x-matroska"
				parsers["Video"] = Video
			}
			for parser, parse := range parsers {
				t.Run(parser+"/"+input, func(t *testing.T) {
					if got := parse(input); !reflect.DeepEqual(got, want) {
						t.Fatalf("got: %+v\nwant: %+v", got, want)
					}
				})
			}
		}
	}
}

func TestMetadataAliasAttachedEpisode(t *testing.T) {
	input := "Example.4x12m720p.x264.DDP5-GROUP.mkv"
	want := Info{Type: Episode, Title: "Example", Season: 4, Episode: 12,
		ScreenSize: "720p", VideoCodec: "H.264", AudioCodec: "Dolby Digital Plus",
		ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska"}
	for parser, parse := range map[string]func(string) Info{"File": File, "Video": Video} {
		t.Run(parser, func(t *testing.T) {
			if got := parse(input); !reflect.DeepEqual(got, want) {
				t.Fatalf("got: %+v\nwant: %+v", got, want)
			}
		})
	}
}

func TestYearlessSubtitleAlias(t *testing.T) {
	for _, alias := range []string{"E-Subs", "ESubs"} {
		assertVideoInfo(t, "Example."+alias+".720p.x264-GROUP.mkv", Info{Type: Movie, Title: "Example", ScreenSize: "720p", VideoCodec: "H.264", ReleaseGroup: "GROUP", Languages: []language.Tag{language.English}, Container: "mkv", MimeType: "video/x-matroska"})
	}
}

func TestMetadataAliasNegativeInfo(t *testing.T) {
	for _, test := range []struct {
		input string
		want  Info
	}{
		{
			"The.Persian.Story.720p.x264-GROUP.mkv",
			Info{Type: Movie, Title: "The Persian Story", ScreenSize: "720p", VideoCodec: "H.264", ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska"},
		},
		{
			"Persian.720p.x264-GROUP.mkv",
			Info{Type: Movie, Title: "Persian", ScreenSize: "720p", VideoCodec: "H.264", ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska"},
		},
		{
			"Example.2024.720p.264.MPA2.0.LAT.VOF.es-GROUP.mkv",
			Info{Type: Movie, Title: "Example", Year: 2024, ScreenSize: "720p", ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska"},
		},
		{
			"Example.2024.720p.x264-QAACx2Team.mkv",
			Info{Type: Movie, Title: "Example", Year: 2024, ScreenSize: "720p", VideoCodec: "H.264", ReleaseGroup: "QAACx2Team", Container: "mkv", MimeType: "video/x-matroska"},
		},
		{
			"Example.2024.720p.x264-DDP5Team.mkv",
			Info{Type: Movie, Title: "Example", Year: 2024, ScreenSize: "720p", VideoCodec: "H.264", ReleaseGroup: "DDP5Team", Container: "mkv", MimeType: "video/x-matroska"},
		},
		{
			"Example.2024.720p.x264-ESubscription.mkv",
			Info{Type: Movie, Title: "Example", Year: 2024, ScreenSize: "720p", VideoCodec: "H.264", ReleaseGroup: "ESubscription", Container: "mkv", MimeType: "video/x-matroska"},
		},
	} {
		t.Run(test.input, func(t *testing.T) {
			if got := File(test.input); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("got: %+v\nwant: %+v", got, test.want)
			}
		})
	}
}

func TestMetadataAliasLanguageBoundaries(t *testing.T) {
	for _, metadata := range []string{"Persian", "Persianate.ESubscription.LAT.VOF.es", "The.Persian.Story"} {
		languages, multilingual := extractLanguages(metadata)
		if languages != nil || multilingual {
			t.Errorf("%q: got %v, %t; want no language metadata", metadata, languages, multilingual)
		}
	}
	for _, metadata := range []string{"PERSIAN", "Persian.720p", "Persian.QAACx2"} {
		languages, multilingual := extractLanguages(metadata)
		if !reflect.DeepEqual(languages, []language.Tag{language.Persian}) || multilingual {
			t.Errorf("%q: got %v, %t; want Persian only", metadata, languages, multilingual)
		}
	}
}

func TestEnglishSubtitleAliasNormalization(t *testing.T) {
	for input, want := range map[string]string{
		"E-Subs":                  "ESubs",
		"e-subs":                  "ESubs",
		"[E-Subs]":                "[ESubs]",
		"E-Subs_E-Subs":           "ESubs_ESubs",
		"E-Subs.E-Subs":           "ESubs.ESubs",
		"ENG.E-Subs.JPN":          "ENG.ESubs.JPN",
		"ESubs":                   "ESubs",
		"SomeE-Subs.E-Subscriber": "SomeE-Subs.E-Subscriber",
		"1E-Subs.E-Subs2":         "1E-Subs.E-Subs2",
	} {
		t.Run(input, func(t *testing.T) {
			if got := normalizeSubtitleAliases(input); got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}
	languages, multilingual := extractLanguages(normalizeSubtitleAliases("GER.E-Subs.ENG.JPN"))
	want := []language.Tag{language.German, language.English, language.Japanese}
	if !reflect.DeepEqual(languages, want) || !multilingual {
		t.Fatalf("got %v, %t; want %v, true", languages, multilingual, want)
	}
}
