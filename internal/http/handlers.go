package http

import (
	"github.com/gofiber/fiber/v2"

	"github.com/alparslanahmed/keyword-market-mcp/internal/mcp"
	"github.com/alparslanahmed/keyword-market-mcp/internal/validation"
)

func (s *Server) handleHealth(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"ok": true, "status": "healthy"})
}

func (s *Server) handleVersion(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"name": s.deps.Name, "version": s.deps.Version})
}

func (s *Server) handleRecentRuns(c *fiber.Ctx) error {
	runs, err := s.deps.Store.ListRecentRuns(20)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(mcp.Envelope{OK: false, Error: mcp.ToMCPError(err)})
	}
	items := make([]fiber.Map, 0, len(runs))
	for _, r := range runs {
		items = append(items, fiber.Map{
			"runId":            r.ID,
			"idea":             r.Idea,
			"country":          r.Country,
			"language":         r.Language,
			"verdict":          r.Verdict,
			"opportunityScore": r.OpportunityScore,
			"createdAt":        r.CreatedAt,
		})
	}
	return c.JSON(mcp.Envelope{OK: true, Data: fiber.Map{"runs": items}})
}

func (s *Server) handleCacheStats(c *fiber.Ctx) error {
	stats, err := s.deps.Cache.Stats()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(mcp.Envelope{OK: false, Error: mcp.ToMCPError(err)})
	}
	return c.JSON(mcp.Envelope{OK: true, Data: stats})
}

func (s *Server) handleValidateProductIdea(c *fiber.Ctx) error {
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
	if err := c.BodyParser(&args); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(mcp.Envelope{
			OK:    false,
			Error: &mcp.MCPError{Code: "VALIDATION_INPUT", Message: "Invalid JSON body.", Retryable: false},
		})
	}
	result, err := s.deps.Service.ValidateProductIdea(c.Context(), validation.ValidateInput{
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
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(mcp.Envelope{OK: false, Error: mcp.ToMCPError(err)})
	}
	return c.JSON(mcp.Envelope{OK: true, Data: result})
}
