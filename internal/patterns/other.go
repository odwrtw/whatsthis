package patterns

func registerOtherPatterns(m *Matcher) {
	AddConstPattern(m, "other", `(?i)\bproper\b`, "Proper")
	AddConstPattern(m, "other", `(?i)\brepack\b`, "Proper")
	AddConstPattern(m, "other", `(?i)\brerip\b`, "Proper")
	AddConstPattern(m, "other", `(?i)\brip\b`, "Rip")
	AddConstPattern(m, "other", `(?i)\bre-?encoded\b|\breenc\b`, "Reencoded")
	AddConstPattern(m, "other", `(?i)\bcomplete\b`, "Complete")
	AddConstPattern(m, "other", `(?i)\bdual[ ._-]?audio\b`, "Dual Audio")
}
