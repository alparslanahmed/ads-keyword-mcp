package http

import "github.com/gofiber/fiber/v2"

func (s *Server) registerRoutes() {
	s.app.Get("/health", s.handleHealth)
	s.app.Get("/version", s.handleVersion)

	debug := s.app.Group("/debug", s.debugAuth)
	debug.Get("/recent-runs", s.handleRecentRuns)
	debug.Get("/cache-stats", s.handleCacheStats)
	debug.Post("/validate-product-idea", s.handleValidateProductIdea)
}

// debugAuth protects the /debug routes. In development they are open; otherwise a
// matching X-API-Key header is required. If no key is configured outside
// development, debug routes are disabled.
func (s *Server) debugAuth(c *fiber.Ctx) error {
	if s.deps.AppEnv == "development" {
		return c.Next()
	}
	if s.deps.DebugAPIKey == "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"ok":    false,
			"error": "debug endpoints are disabled (set DEBUG_API_KEY to enable outside development)",
		})
	}
	if c.Get("X-API-Key") != s.deps.DebugAPIKey {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"ok":    false,
			"error": "unauthorized",
		})
	}
	return c.Next()
}
