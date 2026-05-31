package mcp

// Small helpers for building JSON Schema fragments as plain maps.

func object(props map[string]any, required ...string) map[string]any {
	s := map[string]any{"type": "object", "properties": props}
	if len(required) > 0 {
		s["required"] = required
	}
	return s
}

func strProp(desc string) map[string]any {
	return map[string]any{"type": "string", "description": desc}
}
func numProp(desc string) map[string]any {
	return map[string]any{"type": "number", "description": desc}
}
func boolProp(desc string) map[string]any {
	return map[string]any{"type": "boolean", "description": desc}
}

func strArray(desc string) map[string]any {
	return map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": desc}
}

func objArray(item map[string]any, desc string) map[string]any {
	return map[string]any{"type": "array", "items": item, "description": desc}
}

func monthlyVolumeItem() map[string]any {
	return object(map[string]any{
		"year":            numProp("Four-digit year."),
		"month":           numProp("Month number 1-12."),
		"monthlySearches": numProp("Searches in that month."),
	})
}

func schemaValidateProductIdea() map[string]any {
	return object(map[string]any{
		"idea":                 strProp("The product idea to validate."),
		"country":              strProp("Target country (default: Turkey)."),
		"language":             strProp("Target language (default: Turkish)."),
		"seedKeywords":         strArray("Optional seed keywords to merge with generated seeds."),
		"url":                  strProp("Optional URL seed."),
		"audience":             strProp("Optional target audience."),
		"problem":              strProp("Optional problem the product solves."),
		"maxKeywordIdeas":      numProp("Max keyword ideas to fetch (default: 300)."),
		"minSearchVolume":      numProp("Optional minimum average monthly searches filter."),
		"includeAdultKeywords": boolProp("Include adult keywords (default: false)."),
		"forceRefresh":         boolProp("Bypass cache and refresh from the API (default: false)."),
	}, "idea")
}

func schemaDiscoverKeywords() map[string]any {
	return object(map[string]any{
		"seedKeywords": strArray("Seed keywords."),
		"url":          strProp("Optional URL seed."),
		"country":      strProp("Target country (default: Turkey)."),
		"language":     strProp("Target language (default: Turkish)."),
		"limit":        numProp("Max keyword ideas (default: 300)."),
		"forceRefresh": boolProp("Bypass cache (default: false)."),
	})
}

func schemaGetKeywordMetrics() map[string]any {
	return object(map[string]any{
		"keywords":     strArray("Keywords to fetch metrics for."),
		"country":      strProp("Target country (default: Turkey)."),
		"language":     strProp("Target language (default: Turkish)."),
		"forceRefresh": boolProp("Bypass cache (default: false)."),
	}, "keywords")
}

func schemaScoreKeywordOpportunities() map[string]any {
	keyword := object(map[string]any{
		"keyword":              strProp("Keyword text."),
		"avgMonthlySearches":   numProp("Average monthly searches."),
		"competition":          strProp("Competition level (LOW/MEDIUM/HIGH)."),
		"competitionIndex":     numProp("Competition index 0-100."),
		"lowTopOfPageBid":      numProp("Low top-of-page bid in currency units."),
		"highTopOfPageBid":     numProp("High top-of-page bid in currency units."),
		"monthlySearchVolumes": objArray(monthlyVolumeItem(), "Optional monthly volumes."),
	}, "keyword")
	return object(map[string]any{
		"keywords": objArray(keyword, "Keyword metrics to score."),
		"language": strProp("Language (default: Turkish)."),
	}, "keywords")
}

func schemaClusterKeywords() map[string]any {
	keyword := object(map[string]any{
		"keyword":            strProp("Keyword text."),
		"avgMonthlySearches": numProp("Average monthly searches."),
		"competitionIndex":   numProp("Competition index 0-100."),
		"opportunityScore":   numProp("Opportunity score 0-100."),
		"intent":             strProp("Optional precomputed intent."),
	}, "keyword")
	return object(map[string]any{
		"keywords": objArray(keyword, "Keywords to cluster."),
		"language": strProp("Language (default: Turkish)."),
	}, "keywords")
}

func schemaCompareProductIdeas() map[string]any {
	return object(map[string]any{
		"ideas":                  strArray("Product ideas to compare."),
		"country":                strProp("Target country (default: Turkey)."),
		"language":               strProp("Target language (default: Turkish)."),
		"maxKeywordIdeasPerIdea": numProp("Max keyword ideas per idea (default: 150)."),
		"forceRefresh":           boolProp("Bypass cache (default: false)."),
	}, "ideas")
}

func schemaExportResearchRun() map[string]any {
	return object(map[string]any{
		"runId":  strProp("The research run ID to export."),
		"format": map[string]any{"type": "string", "enum": []string{"json", "csv"}, "description": "Export format."},
	}, "runId")
}
