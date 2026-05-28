package tests

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alparslanahmed/keyword-market-mcp/internal/mcp"
	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *mcp.RPCError   `json:"error"`
}

type toolCallResult struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	IsError bool `json:"isError"`
}

type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error *struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		Retryable bool   `json:"retryable"`
	} `json:"error"`
}

func call(t *testing.T, server *mcp.Server, method string, params any) rpcResponse {
	t.Helper()
	req := map[string]any{"jsonrpc": "2.0", "id": 1, "method": method}
	if params != nil {
		req["params"] = params
	}
	raw, err := json.Marshal(req)
	require.NoError(t, err)
	out := server.HandleMessage(context.Background(), raw)
	require.NotNil(t, out)
	var resp rpcResponse
	require.NoError(t, json.Unmarshal(out, &resp))
	return resp
}

func toolEnvelope(t *testing.T, resp rpcResponse) (toolCallResult, envelope) {
	t.Helper()
	require.Nil(t, resp.Error, "unexpected JSON-RPC error")
	var tr toolCallResult
	require.NoError(t, json.Unmarshal(resp.Result, &tr))
	require.Len(t, tr.Content, 1)
	var env envelope
	require.NoError(t, json.Unmarshal([]byte(tr.Content[0].Text), &env))
	return tr, env
}

func TestMCPInitialize(t *testing.T) {
	server, _ := newTestMCPServer(t)
	resp := call(t, server, "initialize", map[string]any{})
	require.Nil(t, resp.Error)
	var result struct {
		ProtocolVersion string `json:"protocolVersion"`
		ServerInfo      struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"serverInfo"`
	}
	require.NoError(t, json.Unmarshal(resp.Result, &result))
	assert.Equal(t, "keyword-market", result.ServerInfo.Name)
	assert.NotEmpty(t, result.ProtocolVersion)
}

func TestMCPToolsList(t *testing.T) {
	server, _ := newTestMCPServer(t)
	resp := call(t, server, "tools/list", nil)
	var result struct {
		Tools []struct {
			Name        string         `json:"name"`
			InputSchema map[string]any `json:"inputSchema"`
		} `json:"tools"`
	}
	require.NoError(t, json.Unmarshal(resp.Result, &result))
	assert.Len(t, result.Tools, 7)
	for _, tool := range result.Tools {
		assert.NotEmpty(t, tool.Name)
		assert.NotNil(t, tool.InputSchema)
	}
}

func TestMCPValidateProductIdeaTool(t *testing.T) {
	server, _ := newTestMCPServer(t)
	resp := call(t, server, "tools/call", map[string]any{
		"name":      "validate_product_idea",
		"arguments": map[string]any{"idea": "portable espresso maker", "country": "Turkey", "language": "Turkish", "maxKeywordIdeas": 60},
	})
	tr, env := toolEnvelope(t, resp)
	assert.False(t, tr.IsError)
	assert.True(t, env.OK)
	var data map[string]any
	require.NoError(t, json.Unmarshal(env.Data, &data))
	assert.NotEmpty(t, data["runId"])
}

func TestMCPToolValidationError(t *testing.T) {
	server, _ := newTestMCPServer(t)
	resp := call(t, server, "tools/call", map[string]any{
		"name":      "validate_product_idea",
		"arguments": map[string]any{"idea": ""},
	})
	tr, env := toolEnvelope(t, resp)
	assert.True(t, tr.IsError)
	assert.False(t, env.OK)
	require.NotNil(t, env.Error)
	assert.Equal(t, "VALIDATION_INPUT", env.Error.Code)
}

func TestMCPUnknownTool(t *testing.T) {
	server, _ := newTestMCPServer(t)
	resp := call(t, server, "tools/call", map[string]any{"name": "does_not_exist", "arguments": map[string]any{}})
	require.NotNil(t, resp.Error)
	assert.Equal(t, -32601, resp.Error.Code)
}

func TestMCPResources(t *testing.T) {
	server, _ := newTestMCPServer(t)

	listResp := call(t, server, "resources/list", nil)
	var list struct {
		Resources []struct {
			URI string `json:"uri"`
		} `json:"resources"`
	}
	require.NoError(t, json.Unmarshal(listResp.Result, &list))
	assert.Len(t, list.Resources, 3)

	tmplResp := call(t, server, "resources/templates/list", nil)
	var tmpl struct {
		ResourceTemplates []map[string]any `json:"resourceTemplates"`
	}
	require.NoError(t, json.Unmarshal(tmplResp.Result, &tmpl))
	assert.Len(t, tmpl.ResourceTemplates, 1)

	readResp := call(t, server, "resources/read", map[string]any{"uri": "config://targeting"})
	var read struct {
		Contents []struct {
			Text string `json:"text"`
		} `json:"contents"`
	}
	require.NoError(t, json.Unmarshal(readResp.Result, &read))
	require.Len(t, read.Contents, 1)
	assert.Contains(t, read.Contents[0].Text, "Turkey")
}

func TestMCPPrompts(t *testing.T) {
	server, _ := newTestMCPServer(t)

	listResp := call(t, server, "prompts/list", nil)
	var list struct {
		Prompts []struct {
			Name string `json:"name"`
		} `json:"prompts"`
	}
	require.NoError(t, json.Unmarshal(listResp.Result, &list))
	assert.Len(t, list.Prompts, 3)

	getResp := call(t, server, "prompts/get", map[string]any{
		"name":      "product_validation_brief",
		"arguments": map[string]string{"runId": "abc123"},
	})
	var get struct {
		Messages []struct {
			Role    string `json:"role"`
			Content struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"messages"`
	}
	require.NoError(t, json.Unmarshal(getResp.Result, &get))
	require.Len(t, get.Messages, 1)
	assert.Contains(t, get.Messages[0].Content.Text, "abc123")
}

func TestMCPParseError(t *testing.T) {
	server, _ := newTestMCPServer(t)
	out := server.HandleMessage(context.Background(), []byte("this is not json"))
	require.NotNil(t, out)
	var resp rpcResponse
	require.NoError(t, json.Unmarshal(out, &resp))
	require.NotNil(t, resp.Error)
	assert.Equal(t, -32700, resp.Error.Code)
}

func TestMCPNotificationNoResponse(t *testing.T) {
	server, _ := newTestMCPServer(t)
	out := server.HandleMessage(context.Background(), []byte(`{"jsonrpc":"2.0","method":"notifications/initialized"}`))
	assert.Nil(t, out)
}

func TestToMCPErrorMapping(t *testing.T) {
	rl := mcp.ToMCPError(utils.NewRateLimitError(nil))
	assert.Equal(t, "RATE_LIMIT", rl.Code)
	assert.True(t, rl.Retryable)

	mc := mcp.ToMCPError(utils.NewMissingCredentialsError(nil))
	assert.Equal(t, "MISSING_CREDENTIALS", mc.Code)
	assert.False(t, mc.Retryable)

	// A raw error maps to a generic internal error and never leaks its cause.
	generic := mcp.ToMCPError(errors.New("secret token leak boom"))
	assert.Equal(t, "INTERNAL_ERROR", generic.Code)
	assert.NotContains(t, generic.Message, "boom")
}
