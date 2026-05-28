package mcp

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/alparslanahmed/keyword-market-mcp/internal/export"
	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

const (
	uriRecentRuns  = "research-runs://recent"
	uriRunPrefix   = "research-run://"
	uriCacheStats  = "keyword-cache://stats"
	uriTargeting   = "config://targeting"
	mimeJSON       = "application/json"
	recentRunLimit = 20
)

func (s *Server) registerResources() {
	s.resources = []resourceDef{
		{URI: uriRecentRuns, Name: "Recent research runs", Description: "The most recent product-idea validation runs.", MimeType: mimeJSON},
		{URI: uriCacheStats, Name: "Keyword cache stats", Description: "Cache hit/miss counts and metric freshness.", MimeType: mimeJSON},
		{URI: uriTargeting, Name: "Targeting config", Description: "Supported countries/languages and their Google Ads constants.", MimeType: mimeJSON},
	}
	s.templates = []resourceTemplateDef{
		{URITemplate: "research-run://{runId}", Name: "Research run", Description: "Full stored result for one research run.", MimeType: mimeJSON},
	}
}

func (s *Server) handleResourcesList() map[string]any {
	list := make([]map[string]any, 0, len(s.resources))
	for _, r := range s.resources {
		list = append(list, map[string]any{
			"uri":         r.URI,
			"name":        r.Name,
			"description": r.Description,
			"mimeType":    r.MimeType,
		})
	}
	return map[string]any{"resources": list}
}

func (s *Server) handleResourceTemplatesList() map[string]any {
	list := make([]map[string]any, 0, len(s.templates))
	for _, t := range s.templates {
		list = append(list, map[string]any{
			"uriTemplate": t.URITemplate,
			"name":        t.Name,
			"description": t.Description,
			"mimeType":    t.MimeType,
		})
	}
	return map[string]any{"resourceTemplates": list}
}

type resourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text"`
}

type readResourceResult struct {
	Contents []resourceContent `json:"contents"`
}

func (s *Server) handleResourcesRead(_ context.Context, req Request) Response {
	var params struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return errorResponse(req.ID, codeInvalidParams, "invalid params")
	}

	data, err := s.readResource(params.URI)
	if err != nil {
		return errorResponse(req.ID, codeInvalidParams, ToMCPError(err).Message)
	}
	return resultResponse(req.ID, readResourceResult{
		Contents: []resourceContent{{
			URI:      params.URI,
			MimeType: mimeJSON,
			Text:     string(mustMarshal(data)),
		}},
	})
}

func (s *Server) readResource(uri string) (any, error) {
	switch {
	case uri == uriRecentRuns:
		return s.recentRunsResource()
	case uri == uriCacheStats:
		return s.cache.Stats()
	case uri == uriTargeting:
		return s.targetingResource(), nil
	case strings.HasPrefix(uri, uriRunPrefix):
		runID := trimURIScheme(uri, uriRunPrefix)
		return export.BuildRunExport(s.store, runID)
	default:
		return nil, errUnknownResource(uri)
	}
}

type recentRunItem struct {
	RunID            string  `json:"runId"`
	Idea             string  `json:"idea"`
	Country          string  `json:"country"`
	Language         string  `json:"language"`
	Verdict          string  `json:"verdict"`
	OpportunityScore float64 `json:"opportunityScore"`
	CreatedAt        string  `json:"createdAt"`
}

func (s *Server) recentRunsResource() (any, error) {
	runs, err := s.store.ListRecentRuns(recentRunLimit)
	if err != nil {
		return nil, err
	}
	items := make([]recentRunItem, 0, len(runs))
	for _, r := range runs {
		items = append(items, recentRunItem{
			RunID:            r.ID,
			Idea:             r.Idea,
			Country:          r.Country,
			Language:         r.Language,
			Verdict:          r.Verdict,
			OpportunityScore: r.OpportunityScore,
			CreatedAt:        r.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return map[string]any{"runs": items}, nil
}

func (s *Server) targetingResource() any {
	return map[string]any{
		"countries": s.registry.Countries(),
		"languages": s.registry.Languages(),
	}
}

func errUnknownResource(uri string) error {
	return utils.NewValidationInputError("Unknown resource: " + uri)
}
