package whatsthis

import (
	"encoding/json"
	"reflect"
	"testing"

	"golang.org/x/text/language"
)

func TestCombinedLanguageMetadata(t *testing.T) {
	for _, test := range []struct {
		metadata     string
		languages    []language.Tag
		multilingual bool
	}{
		{"FRENCH", []language.Tag{language.French}, false},
		{"TRUEFRENCH.FRE.fra.French", []language.Tag{language.French}, false},
		{"GERMAN.DL", []language.Tag{language.German}, true},
		{"FR.EN.eng", []language.Tag{language.French, language.English}, true},
		{"ENG+JPN", []language.Tag{language.English, language.Japanese}, true},
		{"ENG&GER", []language.Tag{language.English, language.German}, true},
		{"Hindi.Japanese.ENG", []language.Tag{language.Hindi, language.Japanese, language.English}, true},
		{"VOSTFR.SweSub", []language.Tag{language.French, language.Swedish}, true},
		{"Eng.Sub", []language.Tag{language.English}, false},
		{"GER.deu.German", []language.Tag{language.German}, false},
		{"CHI.zho.Chinese", []language.Tag{language.Chinese}, false},
		{"CHS.CHSSub", []language.Tag{language.SimplifiedChinese}, false},
		{"CHT.CHTSub", []language.Tag{language.TraditionalChinese}, false},
		{"CHS.CHT", []language.Tag{language.SimplifiedChinese, language.TraditionalChinese}, true},
		{"Dual-Audio", nil, true},
		{"DUAL", nil, true},
		{"Multi-Subs", nil, true},
		{"Multiple.Subtitle", nil, true},
		{"WEB-DL", nil, false},
		{"WEB-DLRip", nil, false},
		{"WEB-DL.DL", nil, true},
	} {
		for _, media := range []struct {
			prefix string
			want   Info
		}{
			{"Example.2024.", Info{Type: Movie, Title: "Example", Year: 2024}},
			{"Example.S02E03.", Info{Type: Episode, Title: "Example", Season: 2, Episode: 3}},
			{"Example.S02.", Info{Type: ShowSeason, Title: "Example", Season: 2}},
			{"Example.S02-S04.", Info{Type: Show, Title: "Example", Seasons: []int{2, 3, 4}}},
		} {
			input := media.prefix + test.metadata + ".1080p.x264.AAC-GROUP"
			want := media.want
			want.Languages, want.Multilingual = test.languages, test.multilingual
			want.ScreenSize, want.VideoCodec, want.AudioCodec, want.ReleaseGroup = "1080p", "H.264", "AAC", "GROUP"
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

func TestLeadingLanguageTags(t *testing.T) {
	for _, test := range []struct {
		input string
		want  Info
	}{
		{
			"[Eng Sub] [GROUP] Example.S02E03.[JPN].1080p.x264-GROUP.mkv",
			Info{Type: Episode, Title: "Example", Season: 2, Episode: 3, Languages: []language.Tag{language.English, language.Japanese}, Multilingual: true,
				ScreenSize: "1080p", VideoCodec: "H.264", ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska"},
		},
		{
			"[GROUP] [FRENCH] Example.2024.1080p.x264-GROUP.mkv",
			Info{Type: Movie, Title: "Example", Year: 2024, Languages: []language.Tag{language.French},
				ScreenSize: "1080p", VideoCodec: "H.264", ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska"},
		},
		{
			"[French] Example.S02E03.mkv",
			Info{Type: Episode, Title: "Example", Season: 2, Episode: 3, Languages: []language.Tag{language.French}, Container: "mkv", MimeType: "video/x-matroska"},
		},
		{
			"Example.[DUAL.AUDIO]2024.1080p.x264-GROUP.mkv",
			Info{Type: Movie, Title: "Example", Year: 2024, Multilingual: true,
				ScreenSize: "1080p", VideoCodec: "H.264", ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska"},
		},
		{
			"[EXAMPLE-TORRENTS.PL] Example.S02E03.1080p.x264-GROUP.mkv",
			Info{Type: Episode, Title: "Example", Season: 2, Episode: 3,
				ScreenSize: "1080p", VideoCodec: "H.264", ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska"},
		},
		{
			"[ENG] [GROUP] Example.S02.[ENG].[GER].1080p.x264-GROUP",
			Info{Type: ShowSeason, Title: "Example", Season: 2, Languages: []language.Tag{language.English, language.German}, Multilingual: true,
				ScreenSize: "1080p", VideoCodec: "H.264", ReleaseGroup: "GROUP"},
		},
	} {
		t.Run(test.input, func(t *testing.T) {
			if got := File(test.input); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("got: %+v\nwant: %+v", got, test.want)
			}
		})
	}
}

func TestLanguageWordsRemainInTitles(t *testing.T) {
	for input, title := range map[string]string{
		"The.English.Patient.1080p.x264-GROUP.mkv": "The English Patient",
		"The.French.Beacon.1080p.x264-GROUP.mkv":   "The French Beacon",
		"This.Is.It.1080p.x264-GROUP.mkv":          "This Is It",
		"French.1080p.x264-GROUP.mkv":              "French",
		"Hindi.Medium.1080p.x264-GROUP.mkv":        "Hindi Medium",
		"The.Audio.Story.1080p.x264-GROUP.mkv":     "The Audio Story",
		"Multiple.Choices.1080p.x264-GROUP.mkv":    "Multiple Choices",
		"Audio.1080p.x264-GROUP.mkv":               "Audio",
	} {
		t.Run(input, func(t *testing.T) {
			want := Info{Type: Movie, Title: title, ScreenSize: "1080p", VideoCodec: "H.264", ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska"}
			if got := File(input); !reflect.DeepEqual(got, want) {
				t.Fatalf("got: %+v\nwant: %+v", got, want)
			}
		})
	}
}

func TestTitleAfterEpisodeMarkerLanguageWords(t *testing.T) {
	for input, title := range map[string]string{
		"S02E03.The.English.Patient.1080p.x264-GROUP.mkv": "The English Patient",
		"S02E03.This.Is.It.1080p.x264-GROUP.mkv":          "This Is It",
		"S02E03.The.Audio.Story.1080p.x264-GROUP.mkv":     "The Audio Story",
	} {
		t.Run(input, func(t *testing.T) {
			want := Info{Type: Episode, Title: title, Season: 2, Episode: 3,
				ScreenSize: "1080p", VideoCodec: "H.264", ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska"}
			if got := File(input); !reflect.DeepEqual(got, want) {
				t.Fatalf("got: %+v\nwant: %+v", got, want)
			}
		})
	}
}

func TestEpisodeTitlesAreNotLanguages(t *testing.T) {
	for _, title := range []string{"It.Was.The.Change", "The.Greek.Life", "Cinco.de.Mayo", "Multiple.Choices"} {
		input := "Example.S02E03." + title + ".1080p.x264-GROUP.mkv"
		t.Run(input, func(t *testing.T) {
			want := Info{Type: Episode, Title: "Example", Season: 2, Episode: 3,
				ScreenSize: "1080p", VideoCodec: "H.264", ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska"}
			if got := File(input); !reflect.DeepEqual(got, want) {
				t.Fatalf("got: %+v\nwant: %+v", got, want)
			}
		})
	}
}

func TestLanguageTokenBoundaries(t *testing.T) {
	languages, multilingual := extractLanguages("Englishman.Multiverse.MultipleChoice.Duality.ADL.Frenchman")
	if languages != nil || multilingual {
		t.Fatalf("matched part of a word: %v, %t", languages, multilingual)
	}
}

func TestLanguageVocabularyIsNotArbitraryISOCodes(t *testing.T) {
	// These can be valid language codes, but are not language evidence in
	// release names. In particular, AAC and AVC are codec tags.
	for _, metadata := range []string{"AAC.AVC.NEW.SUB", "CAM.MAY.MAX", "LAT.SUB.VOF"} {
		input := "Example.2024." + metadata + ".1080p.x264-GROUP.mkv"
		if got := File(input); got.Languages != nil || got.Multilingual {
			t.Fatalf("%q: unexpected language metadata: %+v", input, got)
		}
	}
}

func TestLanguageJSON(t *testing.T) {
	for _, test := range []struct {
		metadata string
		want     string
	}{
		{"ENG.FRE.CHS.CHT.ENG", `["en","fr","zh-Hans","zh-Hant"]`},
		{"PERSIAN", `["fa"]`},
		{"WEB-DL", `null`},
		{"MULTi", `null`},
	} {
		t.Run(test.metadata, func(t *testing.T) {
			info := File("Example.2024." + test.metadata + ".1080p.x264-GROUP.mkv")
			encoded, err := json.Marshal(info.Languages)
			if err != nil {
				t.Fatal(err)
			}
			if string(encoded) != test.want {
				t.Fatalf("got %s, want %s", encoded, test.want)
			}
			var decoded []language.Tag
			if err := json.Unmarshal(encoded, &decoded); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(decoded, info.Languages) {
				t.Fatalf("round trip: got %v, want %v", decoded, info.Languages)
			}
		})
	}
}
