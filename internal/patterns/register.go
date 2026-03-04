package patterns

func RegisterDefault(m *Matcher) {
	registerContainerPatterns(m)
	registerSourcePatterns(m)
	registerVideoCodecPatterns(m)
	registerAudioCodecPatterns(m)
	registerScreenSizePatterns(m)
	registerEditionPatterns(m)
	registerStreamingServicePatterns(m)
	registerCRC32Patterns(m)
	registerSizePatterns(m)
	registerBitRatePatterns(m)
	registerCDPatterns(m)
	registerPartPatterns(m)
	registerDatePatterns(m)
	registerWebsitePatterns(m)
	registerLanguagePatterns(m)
	registerCountryPatterns(m)
	registerOtherPatterns(m)
	registerBonusPatterns(m)
	registerFilmPatterns(m)
	registerEpisodePatterns(m)
}
