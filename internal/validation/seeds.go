package validation

import (
	"strings"

	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

// maxSeeds bounds the number of seeds sent to idea generation so expansion stays
// reasonable.
const maxSeeds = 8

// connectors split a product phrase from its audience clause (English/Turkish).
var connectors = []string{" for ", " için "}

// GenerateSeeds derives seed keyword candidates from a free-text idea using
// deterministic, structural heuristics (no LLM): the full phrase, the product
// phrase before any audience connector, the head noun phrase, and an
// adjective-dropped variant. This is intentionally simple and documented as
// heuristic; richer expansion can be layered on later.
func GenerateSeeds(idea, language string) []string {
	clean := utils.NormalizeKeyword(idea)
	if clean == "" {
		return nil
	}

	var seeds []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s != "" {
			seeds = append(seeds, s)
		}
	}

	add(clean)

	product, audience := splitOnConnector(clean)
	if product != clean {
		add(product)
	}

	tokens := strings.Fields(product)
	if len(tokens) >= 2 {
		nounCore := strings.Join(tokens[len(tokens)-2:], " ")
		add(nounCore)
		if audience != "" {
			add(nounCore + " " + audience)
		}
	}
	if len(tokens) >= 3 {
		add(strings.Join(tokens[1:], " ")) // drop leading adjective
	}

	deduped := utils.NormalizeKeywords(seeds)
	if len(deduped) > maxSeeds {
		deduped = deduped[:maxSeeds]
	}
	return deduped
}

func splitOnConnector(s string) (product, audience string) {
	for _, c := range connectors {
		if i := strings.Index(s, c); i > 0 {
			return strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+len(c):])
		}
	}
	return s, ""
}
