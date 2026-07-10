package mcp

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/alparslanahmed/keyword-market-mcp/internal/cache"
	"github.com/alparslanahmed/keyword-market-mcp/internal/db"
	"github.com/alparslanahmed/keyword-market-mcp/internal/googleads"
	"github.com/alparslanahmed/keyword-market-mcp/internal/validation"
)

const protocolVersion = "2024-11-05"

const serverInstructions = "Use validate_product_idea for end-to-end idea research and compare_product_ideas when evaluating alternatives. Default to Turkey and Turkish unless the user specifies another supported market. Treat keyword volume as directional evidence, not proof of sales. Mock mode returns deterministic demo data; do not present it as live Google Ads data."

// toolHandler runs a tool against decoded arguments and returns the data payload
// (which the server wraps in the success envelope) or an error (wrapped in the
// failure envelope).
type toolHandler func(ctx context.Context, args json.RawMessage) (any, error)

type toolDef struct {
	Name        string
	Description string
	InputSchema map[string]any
	Handler     toolHandler
}

type resourceDef struct {
	URI         string
	Name        string
	Description string
	MimeType    string
}

type resourceTemplateDef struct {
	URITemplate string
	Name        string
	Description string
	MimeType    string
}

type promptArg struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

type promptDef struct {
	Name        string
	Description string
	Arguments   []promptArg
	Handler     func(args map[string]string) (string, []promptMessage, error)
}

// Server implements the MCP JSON-RPC interface over any transport.
type Server struct {
	name    string
	version string
	logger  *slog.Logger

	service  *validation.Service
	store    *db.Store
	cache    *cache.Cache
	registry *googleads.TargetingRegistry

	tools     []toolDef
	toolIndex map[string]int
	resources []resourceDef
	templates []resourceTemplateDef
	prompts   []promptDef
}

// Deps bundles the dependencies the MCP server needs.
type Deps struct {
	Name     string
	Version  string
	Logger   *slog.Logger
	Service  *validation.Service
	Store    *db.Store
	Cache    *cache.Cache
	Registry *googleads.TargetingRegistry
}

// NewServer constructs and wires the MCP server.
func NewServer(d Deps) *Server {
	s := &Server{
		name:      d.Name,
		version:   d.Version,
		logger:    d.Logger,
		service:   d.Service,
		store:     d.Store,
		cache:     d.Cache,
		registry:  d.Registry,
		toolIndex: make(map[string]int),
	}
	s.registerTools()
	s.registerResources()
	s.registerPrompts()
	return s
}

// HandleMessage processes one raw JSON-RPC message. It returns the marshaled
// response bytes, or nil for notifications (which expect no reply).
func (s *Server) HandleMessage(ctx context.Context, raw []byte) []byte {
	var req Request
	if err := json.Unmarshal(raw, &req); err != nil {
		return mustMarshal(errorResponse(nil, codeParseError, "parse error"))
	}
	if req.JSONRPC != jsonRPCVersion {
		if req.IsNotification() {
			return nil
		}
		return mustMarshal(errorResponse(req.ID, codeInvalidRequest, "invalid jsonrpc version"))
	}

	resp, isNotification := s.dispatch(ctx, req)
	if isNotification {
		return nil
	}
	return mustMarshal(resp)
}

func (s *Server) dispatch(ctx context.Context, req Request) (Response, bool) {
	switch req.Method {
	case "initialize":
		return resultResponse(req.ID, s.handleInitialize()), false
	case "notifications/initialized", "notifications/cancelled":
		return Response{}, true
	case "ping":
		return resultResponse(req.ID, map[string]any{}), false
	case "tools/list":
		return resultResponse(req.ID, s.handleToolsList()), false
	case "tools/call":
		return s.handleToolsCall(ctx, req), false
	case "resources/list":
		return resultResponse(req.ID, s.handleResourcesList()), false
	case "resources/templates/list":
		return resultResponse(req.ID, s.handleResourceTemplatesList()), false
	case "resources/read":
		return s.handleResourcesRead(ctx, req), false
	case "prompts/list":
		return resultResponse(req.ID, s.handlePromptsList()), false
	case "prompts/get":
		return s.handlePromptsGet(req), false
	default:
		if req.IsNotification() {
			return Response{}, true
		}
		return errorResponse(req.ID, codeMethodNotFound, "method not found: "+req.Method), false
	}
}

func (s *Server) handleInitialize() map[string]any {
	return map[string]any{
		"protocolVersion": protocolVersion,
		"instructions":    serverInstructions,
		"capabilities": map[string]any{
			"tools":     map[string]any{},
			"resources": map[string]any{},
			"prompts":   map[string]any{},
		},
		"serverInfo": map[string]any{
			"name":    s.name,
			"version": s.version,
		},
	}
}

// ---- tool result types ----

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type toolResult struct {
	Content []contentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

func envelopeResult(env Envelope, isError bool) toolResult {
	return toolResult{
		Content: []contentBlock{{Type: "text", Text: string(mustMarshal(env))}},
		IsError: isError,
	}
}

func (s *Server) handleToolsCall(ctx context.Context, req Request) Response {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return errorResponse(req.ID, codeInvalidParams, "invalid params")
	}
	idx, ok := s.toolIndex[params.Name]
	if !ok {
		return errorResponse(req.ID, codeMethodNotFound, "unknown tool: "+params.Name)
	}
	data, err := s.tools[idx].Handler(ctx, params.Arguments)
	if err != nil {
		s.logger.Warn("tool call failed", "tool", params.Name, "code", string(ToMCPError(err).Code))
		return resultResponse(req.ID, envelopeResult(errorEnvelope(err), true))
	}
	return resultResponse(req.ID, envelopeResult(successEnvelope(data), false))
}

func (s *Server) handleToolsList() map[string]any {
	list := make([]map[string]any, 0, len(s.tools))
	for _, t := range s.tools {
		list = append(list, map[string]any{
			"name":        t.Name,
			"description": t.Description,
			"inputSchema": t.InputSchema,
		})
	}
	return map[string]any{"tools": list}
}

func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		// Marshaling our own response types should never fail; fall back to a
		// minimal valid JSON-RPC error so the transport stays well-formed.
		return []byte(`{"jsonrpc":"2.0","id":null,"error":{"code":-32603,"message":"internal error"}}`)
	}
	return b
}

// trimURIScheme returns the part after "scheme://".
func trimURIScheme(uri, scheme string) string {
	return strings.TrimPrefix(uri, scheme)
}
