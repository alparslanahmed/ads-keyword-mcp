package http

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"github.com/alparslanahmed/keyword-market-mcp/internal/cache"
	"github.com/alparslanahmed/keyword-market-mcp/internal/db"
	"github.com/alparslanahmed/keyword-market-mcp/internal/validation"
)

// Deps are the dependencies for the HTTP debug server.
type Deps struct {
	Service     *validation.Service
	Store       *db.Store
	Cache       *cache.Cache
	Logger      *slog.Logger
	Name        string
	Version     string
	AppEnv      string
	DebugAPIKey string
}

// Server is the optional Fiber HTTP server for health checks and local
// debugging. It is intentionally secondary to the MCP server.
type Server struct {
	app  *fiber.App
	deps Deps
}

// New builds the HTTP server and registers routes.
func New(deps Deps) *Server {
	app := fiber.New(fiber.Config{
		AppName:               deps.Name,
		DisableStartupMessage: true,
	})
	s := &Server{app: app, deps: deps}
	s.registerRoutes()
	return s
}

// App exposes the underlying Fiber app (used in tests).
func (s *Server) App() *fiber.App { return s.app }

// Listen starts serving on addr (e.g., ":8080").
func (s *Server) Listen(addr string) error {
	return s.app.Listen(addr)
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}
