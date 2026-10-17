package whatsthis

import (
	"strings"

	"golang.org/x/text/language"
)

// releaseLanguageToken keeps recognition context separate from language identity.
// Full names can be title words; codes and affixed release markers are stronger.
type releaseLanguageToken struct {
	tag      language.Tag
	fullName bool
}

func (l releaseLanguageToken) valid() bool {
	return l.tag != language.Und
}

// This is a release vocabulary, not a general language-code parser. Accepting
// arbitrary ISO codes would also interpret codecs (AAC, AVC), source labels,
// title words and group names as languages.
var releaseLanguageTokens = func() map[string]releaseLanguageToken {
	tokens := make(map[string]releaseLanguageToken)
	for _, entry := range []struct {
		tag     language.Tag
		name    string
		aliases string
	}{
		{language.English, "english", "eng en esubs"},
		{language.Italian, "italian", "ita it"},
		{language.Russian, "russian", "rus ru"},
		{language.Ukrainian, "ukrainian", "ukr uk"},
		{language.German, "german", "ger deu de"},
		{language.French, "french", "fre fra fr"},
		{language.Spanish, "spanish", "spa es"},
		{language.Portuguese, "portuguese", "por pt"},
		{language.Polish, "polish", "pol pl"},
		{language.Japanese, "japanese", "jpn jap ja"},
		{language.Korean, "korean", "kor ko"},
		{language.Chinese, "chinese", "chi zho zh"},
		{language.SimplifiedChinese, "", "chs"},
		{language.TraditionalChinese, "", "cht"},
		{language.Hindi, "hindi", "hin hi"},
		{language.Turkish, "turkish", "tur"},
		{language.Dutch, "dutch", "dut nld"},
		{language.Swedish, "swedish", "swe"},
		{language.Danish, "danish", "dan"},
		{language.Finnish, "finnish", "fin"},
		{language.Czech, "czech", "cze ces"},
		{language.Slovak, "slovak", "slo slk"},
		{language.Hungarian, "hungarian", "hun"},
		{language.Greek, "greek", "gre ell"},
		{language.Hebrew, "hebrew", "heb"},
		{language.Arabic, "arabic", "ara"},
		{language.Thai, "thai", "tha"},
		{language.Vietnamese, "vietnamese", "vie"},
		{language.Indonesian, "indonesian", "ind"},
		{language.Romanian, "romanian", "rum ron"},
		{language.Bulgarian, "bulgarian", "bul"},
		{language.Croatian, "croatian", "hrv"},
		{language.Serbian, "serbian", "srp"},
		{language.Slovenian, "slovenian", "slv"},
		{language.Estonian, "estonian", "est"},
		{language.Latvian, "latvian", "lav"},
		{language.Lithuanian, "lithuanian", "lit"},
		{language.Icelandic, "icelandic", "isl"},
		{language.Persian, "persian", ""},
	} {
		if entry.name != "" {
			tokens[entry.name] = releaseLanguageToken{tag: entry.tag, fullName: true}
		}
		for _, alias := range strings.Fields(entry.aliases) {
			tokens[alias] = releaseLanguageToken{tag: entry.tag}
		}
	}
	return tokens
}()

// lookupReleaseLanguage handles explicit language tags, including subtitle
// affixes (VOSTFR, SweSub) and the TRUEFRENCH audio-language marker.
// Callers supply lowercase tokens. An unrecognized token has the zero value.
func lookupReleaseLanguage(token string) releaseLanguageToken {
	if lang, ok := releaseLanguageTokens[token]; ok {
		return lang
	}
	for _, prefix := range []string{"vost", "true"} {
		if rest, ok := strings.CutPrefix(token, prefix); ok {
			if lang, ok := releaseLanguageTokens[rest]; ok {
				return releaseLanguageToken{tag: lang.tag}
			}
		}
	}
	for _, suffix := range []string{"subtitles", "subtitle", "subs", "sub", "audio"} {
		if rest, ok := strings.CutSuffix(token, suffix); ok {
			if lang, ok := releaseLanguageTokens[rest]; ok {
				return releaseLanguageToken{tag: lang.tag}
			}
		}
	}
	return releaseLanguageToken{}
}

// extractLanguages identifies generic multilingual markers and explicit
// audio or subtitle language tags from the metadata region.
func extractLanguages(meta string) ([]language.Tag, bool) {
	meta = normalizeSubtitleAliases(meta)
	var languages []language.Tag
	seen := make(map[language.Tag]bool)
	multilingual := false

	tokens := strings.FieldsFunc(meta, isMetaSep)
	for i, token := range tokens {
		lower := strings.ToLower(token)
		upper := token == strings.ToUpper(token)
		metadataNeighbor := i > 0 && isKnownToken(tokens[i-1]) || i+1 < len(tokens) && isKnownToken(tokens[i+1])
		switch lower {
		case "multi", "multiple", "dual", "dl":
			// DL means dual language, except in the source tag WEB-DL.
			if lower == "dl" && i > 0 && strings.EqualFold(tokens[i-1], "web") {
				continue
			}
			multilingual = multilingual || upper || metadataNeighbor
			continue
		}
		lang := lookupReleaseLanguage(lower)
		// Weak words and short codes can also be episode-title text (It Was
		// the Change, Cinco de Mayo, Greek Life). Require release-tag context.
		if len(lower) == 2 && !upper || lang.fullName && !upper && !metadataNeighbor {
			continue
		}
		if lang.valid() && !seen[lang.tag] {
			languages = append(languages, lang.tag)
			seen[lang.tag] = true
		}
	}

	return languages, multilingual || len(languages) > 1
}
