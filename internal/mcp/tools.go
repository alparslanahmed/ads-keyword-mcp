package mcp

import (
	"context"
	"encoding/json"

	"github.com/alparslanahmed/keyword-market-mcp/internal/export"
	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
	"github.com/alparslanahmed/keyword-market-mcp/internal/validation"
)

const defaultLanguage = "Turkish"

func (s *Server) registerTools() {
	s.tools = []toolDef{
		{
			Name:        "validate_product_idea",
			Description: "Validate a product idea end-to-end using keyword demand, competition, commercial intent, clusters, and opportunity scoring.",
			InputSchema: schemaValidateProductIdea(),
			Handler:     s.toolValidateProductIdea,
		},
		{
			Name:        "discover_keywords",
			Description: "Generate keyword ideas from seed keywords and/or a URL.",
			InputSchema: schemaDiscoverKeywords(),
			Handler:     s.toolDiscoverKeywords,
		},
		{
			Name:        "get_keyword_metrics",
			Description: "Fetch historical search metrics for a list of keywords.",
			InputSchema: schemaGetKeywordMetrics(),
			Handler:     s.toolGetKeywordMetrics,
		},
		{
			Name:        "score_keyword_opportunities",
			Description: "Score already-known keyword metrics without calling Google Ads.",
			InputSchema: schemaScoreKeywordOpportunities(),
			Handler:     s.toolScoreKeywordOpportunities,
		},
		{
			Name:        "cluster_keywords",
			Description: "Cluster keywords into themed groups with aggregate metrics.",
			InputSchema: schemaClusterKeywords(),
			Handler:     s.toolClusterKeywords,
		},
		{
			Name:        "compare_product_ideas",
			Description: "Validate and compare multiple product ideas, returning a ranked result.",
			InputSchema: schemaCompareProductIdeas(),
			Handler:     s.toolCompareProductIdeas,
		},
		{
			Name:        "export_research_run",
			Description: "Export a previous research run as JSON or CSV.",
			InputSchema: schemaExportResearchRun(),
			Handler:     s.toolExportResearchRun,
		},
	}
	for i, t := range s.tools {
		s.toolIndex[t.Name] = i
	}
}

func decodeArgs(raw json.RawMessage, v any) error {
	if len(raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, v); err != nil {
		return utils.NewValidationInputError("Invalid tool arguments.")
	}
	return nil
}

func (s *Server) toolValidateProductIdea(ctx context.Context, raw json.RawMessage) (any, error) {
	var args struct {
		Idea                 string   `json:"idea"`
		Country              string   `json:"country"`
		Language             string   `json:"language"`
		SeedKeywords         []string `json:"seedKeywords"`
		URL                  string   `json:"url"`
		Audience             string   `json:"audience"`
		Problem              string   `json:"problem"`
		MaxKeywordIdeas      int      `json:"maxKeywordIdeas"`
		MinSearchVolume      *int64   `json:"minSearchVolume"`
		IncludeAdultKeywords bool     `json:"includeAdultKeywords"`
		ForceRefresh         bool     `json:"forceRefresh"`
	}
	if err := decodeArgs(raw, &args); err != nil {
		return nil, err
	}
	return s.service.ValidateProductIdea(ctx, validation.ValidateInput{
		Idea:                 args.Idea,
		Country:              args.Country,
		Language:             args.Language,
		SeedKeywords:         args.SeedKeywords,
		URL:                  args.URL,
		Audience:             args.Audience,
		Problem:              args.Problem,
		MaxKeywordIdeas:      args.MaxKeywordIdeas,
		MinSearchVolume:      args.MinSearchVolume,
		IncludeAdultKeywords: args.IncludeAdultKeywords,
		ForceRefresh:         args.ForceRefresh,
	})
}

func (s *Server) toolDiscoverKeywords(ctx context.Context, raw json.RawMessage) (any, error) {
	var args struct {
		SeedKeywords []string `json:"seedKeywords"`
		URL          string   `json:"url"`
		Country      string   `json:"country"`
		Language     string   `json:"language"`
		Limit        int      `json:"limit"`
		ForceRefresh bool     `json:"forceRefresh"`
	}
	if err := decodeArgs(raw, &args); err != nil {
		return nil, err
	}
	return s.service.DiscoverKeywords(ctx, validation.DiscoverInput{
		SeedKeywords: args.SeedKeywords,
		URL:          args.URL,
		Country:      args.Country,
		Language:     args.Language,
		Limit:        args.Limit,
		ForceRefresh: args.ForceRefresh,
	})
}

func (s *Server) toolGetKeywordMetrics(ctx context.Context, raw json.RawMessage) (any, error) {
	var args struct {
		Keywords     []string `json:"keywords"`
		Country      string   `json:"country"`
		Language     string   `json:"language"`
		ForceRefresh bool     `json:"forceRefresh"`
	}
	if err := decodeArgs(raw, &args); err != nil {
		return nil, err
	}
	return s.service.GetKeywordMetrics(ctx, validation.MetricsInput{
		Keywords:     args.Keywords,
		Country:      args.Country,
		Language:     args.Language,
		ForceRefresh: args.ForceRefresh,
	})
}

func (s *Server) toolScoreKeywordOpportunities(_ context.Context, raw json.RawMessage) (any, error) {
	var args struct {
		Keywords []validation.ScoreKeywordInput `json:"keywords"`
		Language string                         `json:"language"`
	}
	if err := decodeArgs(raw, &args); err != nil {
		return nil, err
	}
	if len(args.Keywords) == 0 {
		return nil, utils.NewValidationInputError("Provide at least one keyword to score.")
	}
	return map[string]any{"keywords": validation.ScoreKeywords(args.Keywords, languageOrDefault(args.Language))}, nil
}

func (s *Server) toolClusterKeywords(_ context.Context, raw json.RawMessage) (any, error) {
	var args struct {
		Keywords []validation.ClusterKeywordInput `json:"keywords"`
		Language string                           `json:"language"`
	}
	if err := decodeArgs(raw, &args); err != nil {
		return nil, err
	}
	if len(args.Keywords) == 0 {
		return nil, utils.NewValidationInputError("Provide at least one keyword to cluster.")
	}
	return map[string]any{"clusters": validation.ClusterKeywordsTool(args.Keywords, languageOrDefault(args.Language))}, nil
}

func (s *Server) toolCompareProductIdeas(ctx context.Context, raw json.RawMessage) (any, error) {
	var args struct {
		Ideas                  []string `json:"ideas"`
		Country                string   `json:"country"`
		Language               string   `json:"language"`
		MaxKeywordIdeasPerIdea int      `json:"maxKeywordIdeasPerIdea"`
		ForceRefresh           bool     `json:"forceRefresh"`
	}
	if err := decodeArgs(raw, &args); err != nil {
		return nil, err
	}
	return s.service.CompareProductIdeas(ctx, validation.CompareInput{
		Ideas:                  args.Ideas,
		Country:                args.Country,
		Language:               args.Language,
		MaxKeywordIdeasPerIdea: args.MaxKeywordIdeasPerIdea,
		ForceRefresh:           args.ForceRefresh,
	})
}

func (s *Server) toolExportResearchRun(_ context.Context, raw json.RawMessage) (any, error) {
	var args struct {
		RunID  string `json:"runId"`
		Format string `json:"format"`
	}
	if err := decodeArgs(raw, &args); err != nil {
		return nil, err
	}
	if args.RunID == "" {
		return nil, utils.NewValidationInputError("runId is required.")
	}
	exp, err := export.BuildRunExport(s.store, args.RunID)
	if err != nil {
		return nil, err
	}
	if args.Format == export.FormatCSV {
		csvText, err := export.ToCSV(exp)
		if err != nil {
			return nil, err
		}
		return map[string]any{"runId": args.RunID, "format": export.FormatCSV, "csv": csvText}, nil
	}
	return map[string]any{"runId": args.RunID, "format": export.FormatJSON, "run": exp}, nil
}

func languageOrDefault(lang string) string {
	if lang == "" {
		return defaultLanguage
	}
	return lang
}
