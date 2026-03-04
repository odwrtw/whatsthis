package patterns

func registerEditionPatterns(m *Matcher) {
	AddConstPattern(m, "edition", `(?i)\bdirector'?s?[ ._-]?cut\b`, "Director's Cut")
	AddConstPattern(m, "edition", `(?i)\bdc\b`, "Director's Cut")
	AddConstPattern(m, "edition", `(?i)\balternat(?:e|ive)[ ._-]?cut\b`, "Alternative Cut")
	AddConstPattern(m, "edition", `(?i)\bextended(?:[ ._-]?cut)?\b`, "Extended")
	AddConstPattern(m, "edition", `(?i)\buncensored\b`, "Uncensored")
	AddConstPattern(m, "edition", `(?i)\bunrated\b`, "Unrated")
	AddConstPattern(m, "edition", `(?i)\btheatrical(?:[ ._-]?cut)?\b`, "Theatrical")
	AddConstPattern(m, "edition", `(?i)\bspecial[ ._-]?edition\b`, "Special")
	AddConstPattern(m, "edition", `(?i)\bedition[ ._-]?collector\b`, "Collector")
	AddConstPattern(m, "edition", `(?i)\bcollector'?s?(?:[ ._-]?edition)?\b`, "Collector")
	AddConstPattern(m, "edition", `(?i)criterion(?:[ ._-]?(?:collection|edition))?`, "Criterion")
	AddConstPattern(m, "edition", `(?i)\blimited\b`, "Limited")
}
