package intent

import (
	"strings"

	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

// Intent labels.
const (
	Commercial    = "commercial"
	Informational = "informational"
	Navigational  = "navigational"
	ProblemAware  = "problem-aware"
	Comparison    = "comparison"
	Unknown       = "unknown"
)

// ClassifyIntent assigns a single intent label to a keyword. Navigational
// (domain/URL-like) is detected on the raw text; the remaining labels are
// matched against the normalized keyword in priority order so that the strongest
// signal wins when several apply.
func ClassifyIntent(keyword, language string) string {
	if looksNavigational(keyword) {
		return Navigational
	}
	normalized := utils.NormalizeKeyword(keyword)
	if normalized == "" {
		return Unknown
	}
	tokens := strings.Fields(normalized)
	tokenSet := make(map[string]struct{}, len(tokens))
	for _, t := range tokens {
		tokenSet[t] = struct{}{}
	}

	switch {
	case matchesAny(normalized, tokenSet, commercialTerms):
		return Commercial
	case matchesAny(normalized, tokenSet, comparisonTerms):
		return Comparison
	case matchesAny(normalized, tokenSet, problemAwareTerms):
		return ProblemAware
	case matchesAny(normalized, tokenSet, informationalTerms):
		return Informational
	default:
		return Unknown
	}
}

// HasBuyingModifier reports whether the keyword contains any English or Turkish
// buying modifier. Used by scoring's commercial-intent component.
func HasBuyingModifier(keyword string) bool {
	normalized := utils.NormalizeKeyword(keyword)
	if normalized == "" {
		return false
	}
	tokens := strings.Fields(normalized)
	tokenSet := make(map[string]struct{}, len(tokens))
	for _, t := range tokens {
		tokenSet[t] = struct{}{}
	}
	return matchesAny(normalized, tokenSet, BuyingModifiersEN) ||
		matchesAny(normalized, tokenSet, BuyingModifiersTR)
}

func matchesAny(normalized string, tokenSet map[string]struct{}, terms []string) bool {
	for _, term := range terms {
		if strings.ContainsRune(term, ' ') {
			if strings.Contains(normalized, term) {
				return true
			}
			continue
		}
		if _, ok := tokenSet[term]; ok {
			return true
		}
	}
	return false
}

func looksNavigational(keyword string) bool {
	lower := strings.ToLower(strings.TrimSpace(keyword))
	if lower == "" {
		return false
	}
	if strings.Contains(lower, "://") || strings.HasPrefix(lower, "www.") || strings.Contains(lower, " www.") {
		return true
	}
	for _, tld := range []string{".com", ".net", ".org", ".io", ".co", ".gov", ".edu", ".com.tr", ".tr"} {
		if strings.Contains(lower, tld+"/") || strings.HasSuffix(lower, tld) {
			return true
		}
	}
	return false
}
