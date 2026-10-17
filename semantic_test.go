package whatsthis

import (
	"reflect"
	"strings"
	"testing"
)

// These expectations describe filename semantics, not snapshots of parser output.
func assertVideoInfo(t *testing.T, input string, want Info) {
	t.Helper()
	for name, parse := range map[string]func(string) Info{"File": File, "Video": Video} {
		t.Run(name, func(t *testing.T) {
			if got := parse(input); !reflect.DeepEqual(got, want) {
				t.Fatalf("%s(%q) = %+v\nwant %+v", name, input, got, want)
			}
		})
	}
}

func TestMovieTitleIdentity(t *testing.T) {
	for _, test := range []struct {
		name  string
		title string
	}{
		{"Pilot.Worthy.-.The.Wild.Turn", "Pilot Worthy The Wild Turn"},
		{"Example - Hidden Harbor", "Example Hidden Harbor"},
		{"The.Ash.Legend.Rapid.Part.1", "The Ash Legend Rapid Part 1"},
		{"The.Ash.Legend.Rapid.-.Part.2", "The Ash Legend Rapid Part 2"},
		{"Example.Part.IV", "Example Part IV"},
		{"SPRUCE3TIMBER.2.0", "SPRUCE3TIMBER 2.0"},
		{"Copperlight 4.0 You Can (Not) Sail", "Copperlight 4.0 You Can (Not) Sail"},
		{"GARDEN-Falcon", "GARDEN-Falcon"},
		{"MIRROR-SIERRA", "MIRROR-SIERRA"},
		{"DUSK (Tide Storm Breeze)", "DUSK (Tide Storm Breeze)"},
		{"Example.Directors.Cut", "Example"},
		{"Example (1)", "Example"},
	} {
		input := test.name + ".2024.1080p.BluRay.x264.AAC-GROUP.mkv"
		t.Run(test.name, func(t *testing.T) {
			assertVideoInfo(t, input, Info{Type: Movie, Title: test.title, Year: 2024,
				ScreenSize: "1080p", VideoCodec: "H.264", AudioCodec: "AAC", ReleaseGroup: "GROUP",
				Container: "mkv", MimeType: "video/x-matroska"})
		})
	}
}

func TestEpisodeTitleMetadataBoundary(t *testing.T) {
	for _, title := range []string{"Internal Affairs", "Internal.Affairs", "French Connection", "THE.GREEK.LIFE", "The Web Story", "Multiple Choices", "Class of 1999"} {
		for _, technical := range []bool{false, true} {
			input := "Example.S02E03 - " + title
			want := Info{Type: Episode, Title: "Example", Season: 2, Episode: 3, Container: "mkv", MimeType: "video/x-matroska"}
			if strings.Contains(title, "1999") && technical {
				// Keep the existing year-annotation interpretation; no air-date
				// or episode-title field is being introduced in this change.
				want.Year = 1999
			}
			if technical {
				input += " (1080p WEB-DL x264 AAC-GROUP)"
				want.ScreenSize, want.VideoCodec, want.AudioCodec, want.ReleaseGroup = "1080p", "H.264", "AAC", "GROUP"
			}
			assertVideoInfo(t, input+".mkv", want)
		}
	}
	assertVideoInfo(t, "[LEAD] Example.S02E03-TRAIL.mkv", Info{Type: Episode, Title: "Example", Season: 2, Episode: 3, ReleaseGroup: "TRAIL", Container: "mkv", MimeType: "video/x-matroska"})
}

func TestConservativeFilenamePolicies(t *testing.T) {
	for _, test := range []struct {
		input string
		want  Info
	}{
		{"Brillo - Temp.3 [Cap.310][720p][GROUP].mkv", Info{Type: Episode, Title: "Brillo", Season: 3, Episode: 310, ScreenSize: "720p", ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska"}},
		{"Example.2024.720p.BrRip.264.YIFY.mp4", Info{Type: Movie, Title: "Example", Year: 2024, ScreenSize: "720p", ReleaseGroup: "YIFY", Container: "mp4", MimeType: "video/mp4"}},
		{"Oasis.2014.2014.1080p.x264-GROUP.mkv", Info{Type: Movie, Title: "Oasis 2014", Year: 2014, ScreenSize: "1080p", VideoCodec: "H.264", ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska"}},
		{"Example.S02E03.2160p.480p.x264-GROUP.mkv", Info{Type: Episode, Title: "Example", Season: 2, Episode: 3, ScreenSize: "480p", VideoCodec: "H.264", ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska"}},
		{"Example.2024.03.18.1080p.x264-GROUP.mkv", Info{Type: Movie, Title: "Example", Year: 2024, ScreenSize: "1080p", VideoCodec: "H.264", ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska"}},
		{"Example.2024-03-18.1080p.x264-GROUP.mkv", Info{Type: Movie, Title: "Example", Year: 2024, ScreenSize: "1080p", VideoCodec: "H.264", ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska"}},
		{"[GROUP] Scary Movie [1080p][x264][AAC].mkv", Info{Type: Movie, Title: "Scary Movie", ScreenSize: "1080p", VideoCodec: "H.264", AudioCodec: "AAC", ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska"}},
		{"[1080p] [AAC] [GROUP] Example.Movie.x264.mkv", Info{Type: Movie, Title: "Example Movie", ScreenSize: "1080p", AudioCodec: "AAC", VideoCodec: "H.264", ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska"}},
	} {
		t.Run(test.input, func(t *testing.T) { assertVideoInfo(t, test.input, test.want) })
	}
}

func TestBracketedResolutionsAreNotEpisodeNumbers(t *testing.T) {
	for _, size := range []string{"480", "540", "720", "1080", "1440", "2160"} {
		for _, bracket := range []string{"[" + size + "]", "(" + size + ")", "[ " + size + " ]", "( " + size + " )"} {
			for _, suffix := range []string{"", " [x264]"} {
				want := Info{Type: Movie, Title: "Example Movie", ScreenSize: size + "p", ReleaseGroup: "GROUP",
					Container: "mkv", MimeType: "video/x-matroska"}
				if suffix != "" {
					want.VideoCodec = "H.264"
				}
				assertVideoInfo(t, "[GROUP] Example Movie "+bracket+suffix+".mkv", want)
			}
		}
	}
	for _, marker := range []string{"S01E720", "1x720", "Episode 1.720"} {
		assertVideoInfo(t, "[GROUP] Example "+marker+" [1080p].mkv", Info{
			Type: Episode, Title: "Example", Season: 1, Episode: 720, ScreenSize: "1080p", ReleaseGroup: "GROUP",
			Container: "mkv", MimeType: "video/x-matroska",
		})
	}
	for _, marker := range []string{"- 720", "720", "[01]"} {
		episode := 720
		if marker == "[01]" {
			episode = 1
		}
		assertVideoInfo(t, "[GROUP] Example "+marker+" [1080p].mkv", Info{
			Type: Episode, Title: "Example", Episode: episode, ScreenSize: "1080p", ReleaseGroup: "GROUP",
			Container: "mkv", MimeType: "video/x-matroska",
		})
	}
}

func TestMovieTitle3DContext(t *testing.T) {
	for _, test := range []struct {
		input string
		want  Info
	}{
		{"Saw.3D.mkv", Info{Type: Movie, Title: "Saw 3D", Container: "mkv", MimeType: "video/x-matroska"}},
		{"Saw.3D.1080p.x264-GROUP.mkv", Info{Type: Movie, Title: "Saw 3D", ScreenSize: "1080p", VideoCodec: "H.264", ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska"}},
		{"Saw.3D.2010.1080p.x264-GROUP.mkv", Info{Type: Movie, Title: "Saw 3D", Year: 2010, ScreenSize: "1080p", VideoCodec: "H.264", ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska"}},
		{"Saw.[3D].1080p.x264-GROUP.mkv", Info{Type: Movie, Title: "Saw", ScreenSize: "1080p", VideoCodec: "H.264", ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska"}},
		{"Saw.3D.1080p.3D.x264-GROUP.mkv", Info{Type: Movie, Title: "Saw", ScreenSize: "1080p", VideoCodec: "H.264", ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska"}},
		{"Saw.3D.1080p.HSBS.x264-GROUP.mkv", Info{Type: Movie, Title: "Saw", ScreenSize: "1080p", VideoCodec: "H.264", ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska"}},
		{"Example.2024.3D-GROUP.mkv", Info{Type: Movie, Title: "Example", Year: 2024, ReleaseGroup: "GROUP", Container: "mkv", MimeType: "video/x-matroska"}},
		{"Example.2024.1080p.x264.3D.mkv", Info{Type: Movie, Title: "Example", Year: 2024, ScreenSize: "1080p", VideoCodec: "H.264", Container: "mkv", MimeType: "video/x-matroska"}},
	} {
		t.Run(test.input, func(t *testing.T) { assertVideoInfo(t, test.input, test.want) })
	}
}

func TestEpisodeMetadataOnlySuffixes(t *testing.T) {
	for _, size := range []string{"480", "540", "720", "1080", "1440", "2160"} {
		for _, suffix := range []string{"." + size, ".[" + size + "]", " (" + size + ")"} {
			for _, leading := range []string{"", "[720p] "} {
				assertVideoInfo(t, leading+"Example.S01E02"+suffix+".mkv", Info{
					Type: Episode, Title: "Example", Season: 1, Episode: 2, ScreenSize: size + "p",
					Container: "mkv", MimeType: "video/x-matroska",
				})
			}
		}
	}
	for _, suffix := range []string{".2024", ".[2024]", " (2024)"} {
		assertVideoInfo(t, "Example.S01E02"+suffix+".mkv", Info{
			Type: Episode, Title: "Example", Season: 1, Episode: 2, Year: 2024,
			Container: "mkv", MimeType: "video/x-matroska",
		})
		assertVideoInfo(t, "Example.S01E02"+suffix+".[720].mkv", Info{
			Type: Episode, Title: "Example", Season: 1, Episode: 2, Year: 2024, ScreenSize: "720p",
			Container: "mkv", MimeType: "video/x-matroska",
		})
	}
	for _, title := range []string{"Class of 1999", "Class.of.1999", "Room 720", "Room.720", "Real French", "Real.French"} {
		assertVideoInfo(t, "Example.S01E02 - "+title+".mkv", Info{
			Type: Episode, Title: "Example", Season: 1, Episode: 2,
			Container: "mkv", MimeType: "video/x-matroska",
		})
	}
}

func TestTitleIdentityMetamorphisms(t *testing.T) {
	for _, part := range []string{"1", "2", "IV"} {
		for _, technical := range []string{".720p.x264", ".1080p.x265", ".2160p.AV1"} {
			input := "Example.-.Hidden.Harbor.Part." + part + ".2024" + technical + "-GROUP.mkv"
			if got := File(input); got.Title != "Example Hidden Harbor Part "+part || got.ReleaseGroup != "GROUP" {
				t.Fatalf("%q: title identity changed: %+v", input, got)
			}
		}
	}
}

func TestLongEpisodeMetadata(t *testing.T) {
	assertVideoInfo(t, "Example.S02E03."+strings.Repeat("word.", 2000)+"mkv", Info{Type: Episode, Title: "Example", Season: 2, Episode: 3, Container: "mkv", MimeType: "video/x-matroska"})
	assertVideoInfo(t, "Example.S02E03."+strings.Repeat("web.", 10000)+"mkv", Info{Type: Episode, Title: "Example", Season: 2, Episode: 3, Container: "mkv", MimeType: "video/x-matroska"})
	if got := normalizeSubtitleAliases(strings.Repeat("E-Subs.", 2000)); got != strings.Repeat("ESubs.", 2000) {
		t.Fatal("long subtitle-tag sequence was not normalized")
	}
}

func TestUnicodeGroupProcessingSuffix(t *testing.T) {
	for _, letter := range []string{"Ⱥ", "K", "İ"} {
		group := strings.Repeat(letter, 20)
		assertVideoInfo(t, "Example.2024.1080p.x264-"+group+".mkv-note.mkv", Info{
			Type: Movie, Title: "Example", Year: 2024, ScreenSize: "1080p", VideoCodec: "H.264", ReleaseGroup: group, Container: "mkv", MimeType: "video/x-matroska",
		})
	}
}

func TestMovieParentheticalTitleWithoutGroup(t *testing.T) {
	assertVideoInfo(t, "DUSK (Tide Storm Breeze) - 2017.mp4", Info{
		Type: Movie, Title: "DUSK (Tide Storm Breeze)", Year: 2017, Container: "mp4", MimeType: "video/mp4",
	})
}
