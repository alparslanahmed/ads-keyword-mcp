package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/alparslanahmed/keyword-market-mcp/internal/app"
	"github.com/alparslanahmed/keyword-market-mcp/internal/validation"
)

func main() {
	var (
		idea      = flag.String("idea", "", "product idea to validate (required)")
		country   = flag.String("country", "Turkey", "target country")
		language  = flag.String("language", "Turkish", "target language")
		seedsFlag = flag.String("seeds", "", "comma-separated seed keywords")
		url       = flag.String("url", "", "optional URL seed")
		maxIdeas  = flag.Int("max", 300, "max keyword ideas")
		force     = flag.Bool("force-refresh", false, "bypass cache")
	)
	flag.Parse()

	if strings.TrimSpace(*idea) == "" {
		fmt.Fprintln(os.Stderr, "error: --idea is required")
		flag.Usage()
		os.Exit(2)
	}

	cfg := app.Load()
	logger := app.NewLogger(cfg.LogLevel)
	application, err := app.New(cfg, logger)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: "+err.Error())
		os.Exit(1)
	}
	defer application.Close()

	result, err := application.Service.ValidateProductIdea(context.Background(), validation.ValidateInput{
		Idea:            *idea,
		Country:         *country,
		Language:        *language,
		SeedKeywords:    splitCSV(*seedsFlag),
		URL:             *url,
		MaxKeywordIdeas: *maxIdeas,
		ForceRefresh:    *force,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "validation failed: "+err.Error())
		os.Exit(1)
	}

	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: "+err.Error())
		os.Exit(1)
	}
	fmt.Println(string(out))
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
