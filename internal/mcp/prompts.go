package mcp

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

type promptContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type promptMessage struct {
	Role    string        `json:"role"`
	Content promptContent `json:"content"`
}

type getPromptResult struct {
	Description string          `json:"description"`
	Messages    []promptMessage `json:"messages"`
}

func userMessage(text string) []promptMessage {
	return []promptMessage{{Role: "user", Content: promptContent{Type: "text", Text: text}}}
}

func (s *Server) registerPrompts() {
	s.prompts = []promptDef{
		{
			Name:        "product_validation_brief",
			Description: "Guide the assistant to write a concise product validation brief from a research run.",
			Arguments:   []promptArg{{Name: "runId", Description: "The research run ID to summarize.", Required: true}},
			Handler:     promptValidationBrief,
		},
		{
			Name:        "product_idea_comparison",
			Description: "Guide the assistant to compare validated ideas and recommend which to pursue first.",
			Arguments:   []promptArg{{Name: "ideas", Description: "Optional comma-separated list of ideas or run IDs.", Required: false}},
			Handler:     promptIdeaComparison,
		},
		{
			Name:        "keyword_opportunity_analysis",
			Description: "Guide the assistant to find low-competition, high-intent keyword opportunities.",
			Arguments:   []promptArg{{Name: "runId", Description: "Optional research run ID to analyze.", Required: false}},
			Handler:     promptKeywordOpportunity,
		},
	}
}

func (s *Server) handlePromptsList() map[string]any {
	list := make([]map[string]any, 0, len(s.prompts))
	for _, p := range s.prompts {
		list = append(list, map[string]any{
			"name":        p.Name,
			"description": p.Description,
			"arguments":   p.Arguments,
		})
	}
	return map[string]any{"prompts": list}
}

func (s *Server) handlePromptsGet(req Request) Response {
	var params struct {
		Name      string            `json:"name"`
		Arguments map[string]string `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return errorResponse(req.ID, codeInvalidParams, "invalid params")
	}
	for _, p := range s.prompts {
		if p.Name != params.Name {
			continue
		}
		desc, messages, err := p.Handler(params.Arguments)
		if err != nil {
			return errorResponse(req.ID, codeInvalidParams, ToMCPError(err).Message)
		}
		return resultResponse(req.ID, getPromptResult{Description: desc, Messages: messages})
	}
	return errorResponse(req.ID, codeMethodNotFound, "unknown prompt: "+params.Name)
}

func promptValidationBrief(args map[string]string) (string, []promptMessage, error) {
	runID := strings.TrimSpace(args["runId"])
	if runID == "" {
		return "", nil, utils.NewValidationInputError("runId is required for product_validation_brief.")
	}
	text := fmt.Sprintf(`Read the research run resource "research-run://%s" (or call the export_research_run tool with runId "%s").

Using only that data, write a concise product validation brief with these sections:
- Final verdict (strong / promising / weak / inconclusive) and confidence
- Demand evidence (search volumes, totals)
- Competition evidence (competition levels and indices)
- Commercial intent evidence (commercial-intent keywords, top-of-page bids)
- Best keyword clusters
- Product positioning suggestions
- Landing page / content suggestions
- Risks
- Next validation steps

Be factual and careful. Treat keyword volume as a directional signal, not proof of sales.`, runID, runID)
	return "Product validation brief", userMessage(text), nil
}

func promptIdeaComparison(args map[string]string) (string, []promptMessage, error) {
	ideas := strings.TrimSpace(args["ideas"])
	var subject string
	if ideas != "" {
		subject = fmt.Sprintf("the following ideas or run IDs: %s", ideas)
	} else {
		subject = "the most recent validated ideas (see the research-runs://recent resource)"
	}
	text := fmt.Sprintf(`Compare %s.

Use the compare_product_ideas tool or the stored research runs. Then:
- Rank the ideas by opportunity, demand, competition, and commercial intent
- Note the best keyword clusters for each
- Call out the main risks for each
- Recommend which idea to pursue first and explain why

Keep claims grounded in keyword data; it is a directional signal, not confirmed sales.`, subject)
	return "Product idea comparison", userMessage(text), nil
}

func promptKeywordOpportunity(args map[string]string) (string, []promptMessage, error) {
	runID := strings.TrimSpace(args["runId"])
	source := "the keyword metrics provided"
	if runID != "" {
		source = fmt.Sprintf("the keywords from research run \"research-run://%s\"", runID)
	}
	text := fmt.Sprintf(`Analyze %s to identify the best opportunities.

Focus on:
- Low-competition keywords (low competition index) with meaningful search volume
- High commercial-intent keywords (buying modifiers, higher top-of-page bids)
- Long-tail opportunities (3+ word keywords)
- Themed clusters worth targeting first

You may use the score_keyword_opportunities and cluster_keywords tools. Present a short, prioritized list with rationale, and note that keyword volume is a directional signal.`, source)
	return "Keyword opportunity analysis", userMessage(text), nil
}
