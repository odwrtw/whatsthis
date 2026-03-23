package guessit

import "regexp"

// Season/Episode patterns.
var (
	// SxxExx pattern: S01E02, s01e02, S003E03 (case-insensitive).
	reSeasonEpisode = regexp.MustCompile(`(?i)\bS(\d{1,3})[\s.]?E(\d{1,3})\b`)

	// NxNN pattern: 1x02, 10x20.
	reCrossEpisode = regexp.MustCompile(`(?i)\b(\d{1,2})x(\d{2})\b`)

	// Compact episode: 3-4 digits where first 1-2 are season, last 2 are episode.
	// e.g., 217 -> S02E17, 1316 -> S13E16, 104 -> S01E04
	reCompactEpisode = regexp.MustCompile(`(?:^|[.\s])(\d{3,4})(?:[.\s]|$)`)
)

// Year pattern: 4-digit year between 1900 and 2099.
// Boundaries include dots, spaces, underscores, parentheses, and brackets.
var reYear = regexp.MustCompile(`(?:^|[.\s_(\[\]])(((?:19|20)\d{2}))(?:[.\s_)\]]|$)`)

// Screen size patterns.
var reScreenSize = regexp.MustCompile(`(?i)(?:^|[.\s_\[(])(2160|1080|720|540|480)p(?:$|[.\s_\])])`)

// Screen size token pattern.
var reScreenSizeToken = regexp.MustCompile(`(?i)^(2160|1080|720|540|480)p$`)

// Video codec patterns.
var (
	reVideoCodecH264 = regexp.MustCompile(`(?i)\b(?:x\.?264|h\.?264)\b`)
	reVideoCodecH265 = regexp.MustCompile(`(?i)\b(?:x\.?265|h\.?265|hevc)\b`)

	// Spaced video codec: "H 264", "H 265", "X 264", "X 265" (space-separated).
	reVideoCodecH264Spaced = regexp.MustCompile(`(?i)\b[HX]\s+264\b`)
	reVideoCodecH265Spaced = regexp.MustCompile(`(?i)\b[HX]\s+265\b`)
)

// Audio codec patterns.
var (
	// Dolby Digital Plus: DDP or DD+ followed by channel config.
	reAudioDDP = regexp.MustCompile(`(?i)(?:\bDDP[\s.]?\d[\s.]?\d\b|\bDD\+[\s.]?\d[\s.]?\d\b)`)

	// Dolby Digital (non-plus): DD followed by channel config, or AC3.
	reAudioDD = regexp.MustCompile(`(?i)(?:\bDD[\s.]?\d[\s.]?\d\b|\bAC[\s.]?3\b)`)

	// EAC3 (Enhanced AC-3, a.k.a. Dolby Digital Plus).
	reAudioEAC3 = regexp.MustCompile(`(?i)\bE[\s.]?AC[\s.]?3\b`)

	// AAC with optional channel config.
	reAudioAAC = regexp.MustCompile(`(?i)\bAAC(?:[\s.]?\d[\s.]?\d)?\b`)

	// DTS audio codec.
	reAudioDTS = regexp.MustCompile(`(?i)\bDTS(?:[\s.-]?(?:HD|MA|ES))?\b`)
)

// Container / MIME type mappings.
var containerMIME = map[string]string{
	"mp4": "video/mp4",
	"mkv": "",
}

// knownMetadataTokens are tokens that indicate we've left the title area.
// Matched case-insensitively against individual tokens.
var knownMetadataTokens = map[string]bool{
	// Screen sizes.
	"2160p": true, "1080p": true, "720p": true, "540p": true, "480p": true,
	// Video codecs.
	"x264": true, "x265": true, "h264": true, "h265": true, "hevc": true,
	"h.264": true, "h.265": true, "x.264": true, "x.265": true,
	// Audio codecs.
	"aac": true, "ac3": true, "eac3": true,
	"aac2.0": true, "aac5.1": true,
	// Dolby.
	"dd5.1": true, "dd2.0": true, "ddp5.1": true, "ddp2.0": true,
	"dd+5.1": true, "dd+2.0": true,
	// Source / format keywords.
	"hdtv": true, "pdtv": true, "webrip": true, "web-dl": true, "webdl": true,
	"web": true, "web-rip": true,
	"brrip": true, "bdrip": true, "bluray": true, "blu-ray": true, "dvdrip": true, "hdrip": true,
	"e-subs": true, "esubs": true,
	// Streaming services.
	"amzn": true, "nf": true, "dsnp": true, "hmax": true, "ip": true, "cbs": true, "hulu": true, "amazon": true,
	// Other metadata keywords.
	"proper": true, "repack": true, "internal": true, "real": true,
	"hdr": true, "10bit": true, "6ch": true, "5.1": true, "2.0": true,
	"dd5": true, "dd2": true, "ddp5": true, "ddp2": true, "dd+5": true, "dd+2": true,
	"hd":     true,
	"dvdscr": true, "camrip": true, "hdts": true, "hdcam": true,
	"new": true, "dual": true, "vof": true,
}

// Episode title separator: " - " after SxxExx indicates an episode title, not a release group.
var reEpisodeTitleSep = regexp.MustCompile(`^\s*-\s+`)
