package patterns

func registerVideoCodecPatterns(m *Matcher) {
	AddConstPattern(m, "video_codec", `(?i)\b(?:x[ ._-]?264|h[ ._-]?264|avc)\b`, "H.264")
	AddConstPattern(m, "video_codec", `(?i)\b(?:x[ ._-]?265|h[ ._-]?265|hevc)\b`, "H.265")
	AddConstPattern(m, "video_codec", `(?i)\b(?:xvid|x\.vid)\b`, "Xvid")
	AddConstPattern(m, "video_codec", `(?i)\b(?:divx)\b`, "DivX")
	AddConstPattern(m, "video_codec", `(?i)\b(?:vp9)\b`, "VP9")
	AddConstPattern(m, "video_codec", `(?i)\b(?:av1)\b`, "AV1")
	AddConstPattern(m, "video_codec", `(?i)\b(?:mpeg-?2)\b`, "MPEG-2")
}
