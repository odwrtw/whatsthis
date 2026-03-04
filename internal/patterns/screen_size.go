package patterns

func registerScreenSizePatterns(m *Matcher) {
	AddConstPattern(m, "screen_size", `(?i)\b360p\b`, "360p")
	AddConstPattern(m, "screen_size", `(?i)\b480p\b`, "480p")
	AddConstPattern(m, "screen_size", `(?i)\b540p\b`, "540p")
	AddConstPattern(m, "screen_size", `(?i)\b576p\b`, "576p")
	AddConstPattern(m, "screen_size", `(?i)\b720p\b`, "720p")
	AddConstPattern(m, "screen_size", `(?i)\b1080p\b`, "1080p")
	AddConstPattern(m, "screen_size", `(?i)\b1080i\b`, "1080i")
	AddConstPattern(m, "screen_size", `(?i)\b2160p\b`, "2160p")
	AddConstPattern(m, "screen_size", `(?i)\b4k\b`, "2160p")
	AddConstPattern(m, "screen_size", `(?i)\b8k\b`, "4320p")
}
