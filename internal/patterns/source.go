package patterns

func registerSourcePatterns(m *Matcher) {
	AddConstPattern(m, "source", `(?i)\b(?:blu[ .-]?ray|bdrip|bdremux)\b`, "Blu-ray")
	AddConstPattern(m, "source", `(?i)\b(?:hddvd|hd[ .-]?dvd)\b`, "HD-DVD")
	AddConstPattern(m, "source", `(?i)\b(?:web[ ._-]?dl|web[ ._-]?rip|webrip)\b`, "Web")
	AddConstPattern(m, "source", `(?i)\b(?:hdtv)\b`, "HDTV")
	AddConstPattern(m, "source", `(?i)\b(?:dvdrip|dvd)\b`, "DVD")
	AddConstPattern(m, "source", `(?i)\b(?:dvb|pdtv|sdtv)\b`, "Digital TV")
	AddConstPattern(m, "source", `(?i)\b(?:vhs)\b`, "VHS")
	AddConstPattern(m, "source", `(?i)\b(?:cam)\b`, "CAM")

	// Normalize accidental whitespace variants if a pattern is reused later.
	AddConstPattern(m, "source", `(?i)\b(?:hd[ .-]?tv)\b`, "HDTV")
}
