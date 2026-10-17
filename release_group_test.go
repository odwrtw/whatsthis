package whatsthis

import (
	"reflect"
	"strings"
	"testing"

	"golang.org/x/text/language"
)

func TestPrimaryReleaseGroup(t *testing.T) {
	for _, test := range []struct {
		suffix   string
		explicit string
		fallback string
	}{
		{"-TRAIL", "TRAIL", ""},
		{"-TRAIL[Uploader]", "TRAIL", ""},
		{"-TRAIL.[720p]", "TRAIL", ""},
		{"-TRAIL.AAC", "TRAIL", ""},
		{".my_group", "", "my_group"},
		{".HDR10", "", ""},
		{".Atmos", "", ""},
		{"-A1B2C3D4_[site.com]", "", ""},
		{"-[example.com/path]", "", ""},
		{"-Example-Raws (CR) mkv", "Example-Raws", ""},
		{"-(AtlaN64)", "AtlaN64", ""},
		{"-E03", "E03", ""},
		{"-S56", "S56", ""},
		{"-GCJM-0", "GCJM-0", ""},
		{"-my_group", "my_group", ""},
		{"-GRP.EX", "GRP.EX", ""},
		{"-1XBET", "1XBET", ""},
		{"-2HD", "2HD", ""},
		{"-0SEC", "0SEC", ""},
		{"-[PWET.LUMA]", "PWET.LUMA", ""},
		{".SNAKE[example.com]", "", "SNAKE"},
		{".YIFY.[Uploader]", "", "YIFY"},
		{".[TRAIL]", "", "TRAIL"},
		{".[FJRD].[GLIM2NEST]", "", "FJRD"},
		{".[A1B2C3D4]", "", ""},
		{".[MP4 Movie]", "", ""},
		{".[COMMENTARY]", "", ""},
		{".[3D.BLU-RAY].[HSBS]", "", ""},
		// An unknown tag without a domain or another credit is ambiguous;
		// it can only be a fallback, never part of another group's name.
		{"[Uploader]", "", "Uploader"},
		{".example.com", "", ""},
		{"-example.com", "", ""},
		{".HDLIGHT", "", ""},
	} {
		for _, leading := range []string{"", "LEAD"} {
			for _, media := range []struct {
				prefix string
				want   Info
			}{
				{"Example.2024", Info{Type: Movie, Title: "Example", Year: 2024}},
				{"Example.S02E03", Info{Type: Episode, Title: "Example", Season: 2, Episode: 3}},
				{"Example.S02", Info{Type: ShowSeason, Title: "Example", Season: 2}},
				{"Example.S02-S04", Info{Type: Show, Title: "Example", Seasons: []int{2, 3, 4}}},
			} {
				input := media.prefix + ".1080p.x264.AAC" + test.suffix
				if leading != "" {
					input = "[" + leading + "] " + input
				}
				want := media.want
				want.ScreenSize, want.VideoCodec, want.AudioCodec = "1080p", "H.264", "AAC"
				if strings.Contains(test.suffix, "[720p]") {
					want.ScreenSize = "720p"
				}
				want.ReleaseGroup = test.explicit
				if want.ReleaseGroup == "" {
					want.ReleaseGroup = leading
					if leading == "" {
						want.ReleaseGroup = test.fallback
					}
				}
				parsers := map[string]func(string) Info{"File": File}
				if (want.Type == Show || want.Type == ShowSeason) && test.explicit == "GRP.EX" {
					// A bare .EX is indistinguishable from an unknown filename
					// extension; a final metadata tag establishes a release name.
					input += ".[1080p]"
				}
				if want.Type == Movie || want.Type == Episode {
					input += ".mkv"
					want.Container, want.MimeType = "mkv", "video/x-matroska"
					parsers["Video"] = Video
				}
				for parser, parse := range parsers {
					t.Run(parser+"/"+input, func(t *testing.T) {
						if got := parse(input); !reflect.DeepEqual(got, want) {
							t.Fatalf("got %+v\nwant %+v", got, want)
						}
					})
				}
			}
		}
	}
}

func TestReleaseAnnotations(t *testing.T) {
	for _, input := range []string{
		"Example.2024.1080p.x264.AAC-NTb.mkv[example].mkv",
		"Example.2024.1080p.x264.AAC-NTb.mkv-muxed_old.mkv",
		"Example.2024.1080p.x264.AAC-NTb.mkv-ENG.mkv",
		"Example.2024.1080p.x264.AAC-NTb[EXAMPLE.PL].mkv",
		"Example.2024.1080p.WEB-DLs.x264.AAC-NTb.mkv",
		"Example.2024.1080p.x264.AAC-NTb (The Hidden Lantern).mkv",
		"Example.2024.1080p.x264.AAC-NTb .Le phare De Cuivre..mkv",
	} {
		t.Run(input, func(t *testing.T) {
			assertVideoInfo(t, input, Info{Type: Movie, Title: "Example", Year: 2024,
				ScreenSize: "1080p", VideoCodec: "H.264", AudioCodec: "AAC", ReleaseGroup: "NTb", Container: "mkv", MimeType: "video/x-matroska"})
		})
	}
}

func TestGroupTagsDoNotBecomeLanguages(t *testing.T) {
	assertVideoInfo(t, "[example.pl] [1080p] [AAC] [A1B2C3D4] [S01] [05x02] [Hi10] Example.S02E03.1080p.x264.AAC-GROUP.mkv", Info{
		Type: Episode, Title: "Example", Season: 2, Episode: 3, ScreenSize: "1080p", AudioCodec: "AAC", VideoCodec: "H.264", ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska",
	})
}

func TestReleaseGroupMetamorphisms(t *testing.T) {
	for _, prefix := range []string{"Example.2024", "Example.S02E03", "Example.S02", "Example.S02-S04"} {
		base := File(prefix + ".1080p.x264.AAC-GROUP")
		for _, suffix := range []string{"[example.com]", "[Uploader]", ".[A1B2C3D4]", ".[9e8d7c6b]", ".[ENG]"} {
			got := File(prefix + ".1080p.x264.AAC-GROUP" + suffix)
			want := base
			if suffix == ".[ENG]" {
				want.Languages = []language.Tag{language.English}
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("%s%s: got %+v\nwant %+v", prefix, suffix, got, want)
			}
		}
	}
}
