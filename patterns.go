package whatsthis

import "regexp"

// General utility patterns.
var (
	// Matches a double-extension pattern like "file.mkv[tag].mkv".
	reDoubleExt = regexp.MustCompile(`(?i)\.(\w{2,4})\[([^\]]+)\]\.(\w{2,4})$`)

	// Matches one or more whitespace characters.
	reSpaces = regexp.MustCompile(`\s+`)
)

// Season/Episode patterns.
var (
	// SxxEyy-zz pattern: S01E01-08 or S01E01-E08 (case-insensitive).
	reSeasonEpisodeRange = regexp.MustCompile(`(?i)\bS(\d{1,3})[\s.]?E(\d{1,3})\s*-\s*E?(\d{1,3})\b`)

	// Standalone season patterns: S01 and Season 1.
	reSeasonOnly      = regexp.MustCompile(`(?i)\bSn?(\d{1,3})\b`)
	reSeasonWordOnly  = regexp.MustCompile(`(?i)\b(?:Seasons?|Saisons?|Temp)[\s._-]*(\d{1,3})\b`)
	reSeasonOrdinal   = regexp.MustCompile(`(?i)\b(\d{1,3})(?:st|nd|rd|th)\s+Season\b`)
	reSeasonRange     = regexp.MustCompile(`(?i)\bSn?(\d{1,3})\s*-\s*(?:Sn?)?(\d{1,3})\b`)
	reSeasonWordRange = regexp.MustCompile(`(?i)\b(?:Seasons?|Saisons?|Temp)[\s._-]*(\d{1,3})\s*-\s*(?:(?:Seasons?|Saisons?|Temp)[\s._-]*)?(\d{1,3})\b`)

	// Explicit markers may have spaced components, year seasons, long episode
	// numbers, or a split-episode suffix such as S01E02b.
	reSeasonEpisode = regexp.MustCompile(`(?i)(?:^|[^A-Za-z0-9])S(\d{1,4})[\s.x-]*Ep?[\s.#-]*(\d{1,4})(?:$|[^0-9])`)
	reCrossEpisode  = regexp.MustCompile(`(?i)(\d{1,2})x(\d{1,3})`)

	reWordSeasonEpisode  = regexp.MustCompile(`(?i)\bEpisode\s+(\d{1,2})\.(\d{1,3})\b`)
	reLooseSeasonEpisode = regexp.MustCompile(`(?i)\bS(\d{1,3})[\s.-]+(\d{1,4})\b`)
	reEpisodeOnly        = regexp.MustCompile(`(?i)\b(?:Episodes?|Ep|E|Cap(?:itulo)?)[\s.#-]*(\d{1,4})(?:v\d+)?(?:$|[^A-Za-z0-9])`)
	reAbsoluteEpisode    = regexp.MustCompile(`(?i)(?:^|[\s.])-[\s.]+(\d{1,4})(?:[ab])?(?:v\d+)?(?:$|[\s.(\[])`)
	reLeadingEpisode     = regexp.MustCompile(`^(\d{2,3})[\s.-]+`)
	reAnimeEpisode       = regexp.MustCompile(`(?i)(?:^|[\s.])(\d{2,4})(?:v\d+)?(?:$|[\s.(\[])`)
	reEpisodeAfterSeason = regexp.MustCompile(`^[\s.()\[\]-]*(\d{1,4})(?:$|[\s.(\[])`)

	// Compact episode: 3-4 digits where first 1-2 are season, last 2 are episode.
	// E.g., 217 -> S02E17, 1316 -> S13E16, 104 -> S01E04.
	reCompactEpisode = regexp.MustCompile(`(?:^|[.\s])(\d{3,4})(?:[.\s]|$)`)
)

// Year pattern: 4-digit year between 1900 and 2099.
// Boundaries include dots, spaces, underscores, parentheses, and brackets.
var reYear = regexp.MustCompile(`(?:^|[.\s_(\[\]])(((?:19|20)\d{2}))(?:[.\s_)\]]|$)`)

// Known pixel dimensions normalized to the screen-size notation used by Info.
var screenSizeDimensions = map[string]string{
	"1920x1080": "1080p",
	"1920x960":  "1080p",
	"1280x544":  "720p",
	"3840x1632": "2160p",
}

// Screen size token patterns, with optional HD and progressive-scan markers.
var (
	reScreenSizeToken       = regexp.MustCompile(`(?i)^(?:HD)?(2160|1440|1080|720|540|480)([pi])$`)
	reScreenSizeNumberToken = regexp.MustCompile(`(?i)^(2160|1440|1080|720|540|480)$`)
)

// Video codec patterns.
var (
	reVideoCodecH264 = regexp.MustCompile(`(?i)(?:^|[^A-Za-z0-9])(?:x\.?264|h\.?264|avc)(?:$|[^A-Za-z0-9])`)
	reVideoCodecH265 = regexp.MustCompile(`(?i)(?:^|[^A-Za-z0-9])(?:x\.?265|h\.?265|hevc(?:10)?)(?:$|[^A-Za-z0-9])`)
	reVideoCodecXvid = regexp.MustCompile(`(?i)(?:^|[^A-Za-z0-9])xvid(?:$|[^A-Za-z0-9])`)
	reVideoCodecAV1  = regexp.MustCompile(`(?i)(?:^|[^A-Za-z0-9])av1(?:$|[^A-Za-z0-9])`)
	reVideoCodecVP9  = regexp.MustCompile(`(?i)(?:^|[^A-Za-z0-9])vp9(?:$|[^A-Za-z0-9])`)
	reVideoCodecVC1  = regexp.MustCompile(`(?i)(?:^|[^A-Za-z0-9])vc-?1(?:$|[^A-Za-z0-9])`)

	// Spaced video codec: "H 264", "H 265", "X 264", "X 265" (space-separated).
	reVideoCodecH264Spaced = regexp.MustCompile(`(?i)\b[HX]\s+264\b`)
	reVideoCodecH265Spaced = regexp.MustCompile(`(?i)\b[HX]\s+265\b`)
)

// Audio codec patterns.
var (
	// Dolby Digital Plus: DDP or DD+ followed by channel config.
	reAudioDDP = regexp.MustCompile(`(?i)(?:\bDDP[\s.]?\d[\s.]?\d\b|\bDD\+[\s.]?\d[\s.]?\d\b)`)

	// Dolby Digital (non-plus): DD followed by channel config, or AC3.
	reAudioDD = regexp.MustCompile(`(?i)(?:\bDD[\s.]?\d[\s.]?\d\b|\bAC[\s.-]?3\b|\bDolbyD\b)`)

	// EAC3 (Enhanced AC-3, a.k.a. Dolby Digital Plus).
	reAudioEAC3 = regexp.MustCompile(`(?i)\bE[\s.-]?AC[\s.-]?3\b`)

	// AAC with optional channel config.
	reAudioAAC = regexp.MustCompile(`(?i)\bAAC(?:[\s.]?\d[\s.]?\d)?\b`)

	// DTS, FLAC, Opus, and Dolby TrueHD audio codecs.
	reAudioDTS    = regexp.MustCompile(`(?i)\bDTS(?:[\s.-]?(?:HD|MA|ES))?\b`)
	reAudioFLAC   = regexp.MustCompile(`(?i)(?:^|[^A-Za-z0-9])FLAC(?:$|[^A-Za-z0-9])`)
	reAudioOPUS   = regexp.MustCompile(`(?i)(?:^|[^A-Za-z0-9])OPUS(?:$|[^A-Za-z0-9])`)
	reAudioTrueHD = regexp.MustCompile(`(?i)(?:^|[^A-Za-z0-9])TRUEHD(?:$|[^A-Za-z0-9])`)
	reAudioMP3    = regexp.MustCompile(`(?i)\bMP3\b`)
)

// Container / MIME type mappings.
var containerMIME = map[string]string{
	"mp4":  "video/mp4",
	"mkv":  "video/x-matroska",
	"avi":  "video/x-msvideo",
	"webm": "video/webm",
	"mov":  "video/quicktime",
	"m4v":  "video/x-m4v",
	"ts":   "video/mp2t",
	"wmv":  "video/x-ms-wmv",
	"flv":  "video/x-flv",
	"srt":  "application/x-subrip",
}

// knownMetadataTokens are tokens that indicate we've left the title area.
// Matched case-insensitively against individual tokens.
var knownMetadataTokens = map[string]bool{
	// Screen sizes.
	"4k": true, "2160p": true, "1440p": true, "1080p": true, "720p": true, "540p": true, "480p": true,
	// Video codecs.
	"x264": true, "x265": true, "h264": true, "h265": true, "hevc": true, "avc": true, "av1": true, "vp9": true, "vc-1": true,
	"h.264": true, "h.265": true, "x.264": true, "x.265": true,
	// Audio codecs.
	"aac": true, "ac3": true, "eac3": true, "flac": true, "truehd": true, "opus": true,
	"aac2.0": true, "aac5.1": true,
	// Dolby.
	"dd5.1": true, "dd2.0": true, "ddp5.1": true, "ddp2.0": true,
	"dd+5.1": true, "dd+2.0": true,
	// Source / format keywords.
	"hdtv": true, "pdtv": true, "webrip": true, "web-dl": true, "web-dlrip": true, "webdl": true,
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
	"new": true, "dual": true, "vof": true, "multi": true,
	"multiple": true, "dl": true, "audio": true,
	"sub": true, "subs": true, "subtitle": true, "subtitles": true,
	"uncut": true, "unrated": true, "dc": true,
}

// Common release language codes, names, and aliases, normalized to English names.
// Short codes and full names require context when locating the title boundary.
var releaseLanguageNames = map[string]string{
	"eng": "English", "ita": "Italian", "rus": "Russian", "ukr": "Ukrainian",
	"ger": "German", "deu": "German", "fre": "French", "fra": "French",
	"spa": "Spanish", "por": "Portuguese", "pol": "Polish",
	"jpn": "Japanese", "jap": "Japanese", "kor": "Korean",
	"chi": "Chinese", "zho": "Chinese", "chs": "Chinese (Simplified)", "cht": "Chinese (Traditional)",
	"hin": "Hindi", "tur": "Turkish", "dut": "Dutch", "nld": "Dutch",
	"swe": "Swedish", "dan": "Danish", "fin": "Finnish",
	"cze": "Czech", "ces": "Czech", "slo": "Slovak", "slk": "Slovak",
	"hun": "Hungarian", "gre": "Greek", "ell": "Greek", "heb": "Hebrew",
	"ara": "Arabic", "tha": "Thai", "vie": "Vietnamese", "ind": "Indonesian",
	"rum": "Romanian", "ron": "Romanian", "bul": "Bulgarian", "hrv": "Croatian",
	"srp": "Serbian", "slv": "Slovenian", "est": "Estonian", "lav": "Latvian",
	"lit": "Lithuanian", "isl": "Icelandic",
	"en": "English", "fr": "French", "de": "German", "it": "Italian",
	"es": "Spanish", "pt": "Portuguese", "pl": "Polish", "ru": "Russian", "uk": "Ukrainian",
	"ja": "Japanese", "ko": "Korean", "zh": "Chinese", "hi": "Hindi",
	"english": "English", "french": "French", "german": "German", "italian": "Italian",
	"spanish": "Spanish", "portuguese": "Portuguese", "polish": "Polish",
	"russian": "Russian", "ukrainian": "Ukrainian", "japanese": "Japanese",
	"korean": "Korean", "chinese": "Chinese", "hindi": "Hindi", "turkish": "Turkish",
	"dutch": "Dutch", "swedish": "Swedish", "danish": "Danish", "finnish": "Finnish",
	"czech": "Czech", "slovak": "Slovak", "hungarian": "Hungarian", "greek": "Greek",
	"hebrew": "Hebrew", "arabic": "Arabic", "thai": "Thai", "vietnamese": "Vietnamese",
	"indonesian": "Indonesian", "romanian": "Romanian", "bulgarian": "Bulgarian",
	"croatian": "Croatian", "serbian": "Serbian", "slovenian": "Slovenian",
	"estonian": "Estonian", "latvian": "Latvian", "lithuanian": "Lithuanian", "icelandic": "Icelandic",
}

// Episode title separator: " - " after SxxExx indicates an episode title, not a release group.
var reEpisodeTitleSep = regexp.MustCompile(`^\s*-\s+`)

// compoundTokens lists hyphenated tokens that should NOT be split into
// a release-group boundary (e.g., "WEB-DL" is a single metadata token).
var compoundTokens = map[string]bool{
	"web-dl":    true,
	"web-dlrip": true,
	"web-rip":   true,
	"blu-ray":   true,
	"e-subs":    true,
}

// Cleaning / utility patterns used in the parsing pipeline.
var (
	// Matches a parenthesized domain-like suffix, e.g., "(site.com)".
	reParenDomainSuffix = regexp.MustCompile(`\s*\.?\s*\([^)]*\.[^)]*\)\s*$`)

	// Matches a trailing parenthesized group, e.g., "(content)".
	reTrailingParen = regexp.MustCompile(`\s*\(([^)]+)\)\s*$`)

	// Matches numeric tokens with a letter suffix, e.g., "700mb", "6ch".
	reNumericSuffix = regexp.MustCompile(`^\d+[a-z]+$`)

	// Matches a standalone screen-size number in a title, e.g., ".1080.".
	reScreenSizeInTitle = regexp.MustCompile(`(?i)(^|[ ._-])(4k|2160|1440|1080|720|540|480)([ ._-]|$)`)

	// Matches an unclosed trailing parenthesis, e.g., "Title (foo".
	reUnclosedParen = regexp.MustCompile(`\s*\([^)]*$`)

	// Matches a parenthesized year, e.g., "(2019)".
	reParenYear = regexp.MustCompile(`(?i)\(\s*(?:19|20)\d{2}\s*\)`)

	// Matches source tags that sometimes precede an episode marker.
	reEpisodeSourceSuffix = regexp.MustCompile(`(?i)[._\s-]+(?:dvdrip|bdrip|brrip|webrip|web-dl|web-dlrip|hdtv|pdtv|hdrip|blu-ray|bluray)[._\s-]*$`)

	// Matches empty parentheses, e.g., "( )".
	reEmptyParen = regexp.MustCompile(`\(\s*\)`)

	// Matches optional whole-season annotations.
	reComplete = regexp.MustCompile(`(?i)\[?\b(?:The[\s._-]+)?COMPLETE\b\]?`)

	// Matches bracketed whole-season descriptors and non-group file tags.
	reSeasonDescriptorBracket = regexp.MustCompile(`(?i)\[\s*(?:S\d{1,3}|Season[\s._-]*\d{1,3})\b[^\]]*\]`)
	reFileSizeBracket         = regexp.MustCompile(`(?i)\[\s*\d+(?:\.\d+)?\s*(?:KB|MB|GB|TB)\s*\]`)
	reContainerBracket        = regexp.MustCompile(`(?i)\[\s*(?:MP4|MKV|AVI|WEBM|MOV|M4V|TS|WMV|FLV)\s*\]`)

	// Matches a "Part N" suffix (roman or arabic numerals).
	rePartSuffix = regexp.MustCompile(`(?i)\s+part\s+[0-9ivx]+$`)

	// Matches bracket tags in titles, e.g., "[720p]".
	reBracketTag = regexp.MustCompile(`(?:^|[\s._-])\[[^\]]+\]`)

	// Matches a www.* website prefix.
	reWebsitePrefix = regexp.MustCompile(`(?i)^www\.[a-z0-9-]+(?:\.[a-z0-9-]+)+[._\-\s]+`)
)
