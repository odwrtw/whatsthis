package patterns

func registerCountryPatterns(m *Matcher) {
	AddConstPattern(m, "country", `(?i)\bUS\b|\bUSA\b`, "US")
	AddConstPattern(m, "country", `(?i)\bUK\b`, "UK")
	AddConstPattern(m, "country", `(?i)\bFR\b|\bFrance\b`, "FR")
	AddConstPattern(m, "country", `(?i)\bCA\b|\bCanada\b|\bQuébec\b`, "CA")
}
