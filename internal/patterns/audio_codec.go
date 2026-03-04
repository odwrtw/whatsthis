package patterns

func registerAudioCodecPatterns(m *Matcher) {
	AddConstPattern(m, "audio_codec", `(?i)\b(?:dts[ .-]?hd)\b`, "DTS-HD")
	AddConstPattern(m, "audio_codec", `(?i)\b(?:dts)\b`, "DTS")
	AddConstPattern(m, "audio_codec", `(?i)\b(?:true[ .-]?hd)\b`, "TrueHD")
	AddConstPattern(m, "audio_codec", `(?i)\b(?:e-?ac-?3|ac-?3|dd(?:p)?(?:[ ._-]?\d(?:[ ._-]?\d)?)?)\b`, "Dolby Digital")
	AddConstPattern(m, "audio_codec", `(?i)\b(?:aac(?:[ ._-]?\d\.\d)?)\b`, "AAC")
	AddConstPattern(m, "audio_codec", `(?i)\b(?:flac)\b`, "FLAC")
	AddConstPattern(m, "audio_codec", `(?i)\b(?:mp3)\b`, "MP3")
	AddConstPattern(m, "audio_codec", `(?i)\b(?:opus)\b`, "Opus")

	AddPatternSpec(m, "audio_channels", `\b(?:7\.1|5\.1|2\.1|2\.0|1\.0)\b`, PatternSpec{})
}
