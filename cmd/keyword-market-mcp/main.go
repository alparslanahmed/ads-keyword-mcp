package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/alparslanahmed/keyword-market-mcp/internal/app"
	httpserver "github.com/alparslanahmed/keyword-market-mcp/internal/http"
	"github.com/alparslanahmed/keyword-market-mcp/internal/mcp"
)

func main() {
	var (
		stdioFlag   = flag.Bool("stdio", false, "run the MCP server over stdio (primary transport)")
		httpFlag    = flag.Bool("http", false, "run the Fiber HTTP debug server")
		httpAddr    = flag.String("http-addr", "", "HTTP listen address (default :HTTP_PORT)")
		migrateFlag = flag.Bool("migrate", false, "run database migrations and exit")
	)
	flag.Parse()

	// Default to stdio when no mode is selected.
	runStdio := *stdioFlag
	runHTTP := *httpFlag
	if !runStdio && !runHTTP && !*migrateFlag {
		runStdio = true
	}

	cfg := app.Load()
	logger := app.NewLogger(cfg.LogLevel)

	application, err := app.New(cfg, logger)
	if err != nil {
		logger.Error("failed to start application", "error", err.Error())
		os.Exit(1)
	}
	defer application.Close()

	// app.New runs AutoMigrate; --migrate just confirms and exits.
	if *migrateFlag {
		logger.Info("database migrations applied", "dsn", cfg.DatabaseDSN)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	addr := *httpAddr
	if addr == "" {
		addr = ":" + cfg.HTTPPort
	}

	var httpSrv *httpserver.Server
	if runHTTP {
		httpSrv = httpserver.New(httpserver.Deps{
			Service:     application.Service,
			Store:       application.Store,
			Cache:       application.Cache,
			Logger:      logger,
			Name:        cfg.MCPServerName,
			Version:     cfg.MCPServerVersion,
			AppEnv:      cfg.AppEnv,
			DebugAPIKey: cfg.DebugAPIKey,
		})
		go func() {
			logger.Info("starting HTTP debug server", "addr", addr)
			if err := httpSrv.Listen(addr); err != nil {
				logger.Error("HTTP server stopped", "error", err.Error())
			}
		}()
	}

	if runStdio {
		server := mcp.NewServer(mcp.Deps{
			Name:     cfg.MCPServerName,
			Version:  cfg.MCPServerVersion,
			Logger:   logger,
			Service:  application.Service,
			Store:    application.Store,
			Cache:    application.Cache,
			Registry: application.Registry,
		})
		logger.Info("starting MCP server over stdio", "name", cfg.MCPServerName, "version", cfg.MCPServerVersion, "mock", cfg.UseMock)
		transport := mcp.NewStdioTransport(server, os.Stdin, os.Stdout)
		if err := transport.Serve(ctx); err != nil && ctx.Err() == nil {
			logger.Error("stdio transport error", "error", err.Error())
		}
	} else {
		// HTTP-only mode: block until signal.
		<-ctx.Done()
	}

	if httpSrv != nil {
		_ = httpSrv.Shutdown()
	}
	logger.Info("shutdown complete")
}
