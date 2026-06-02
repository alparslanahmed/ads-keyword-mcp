package googleads

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

type fakeRoundTripper struct {
	status int
	body   string
}

func (f fakeRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: f.status,
		Body:       io.NopCloser(strings.NewReader(f.body)),
		Header:     make(http.Header),
	}, nil
}

// On a non-2xx Google Ads response, the raw body must be written to the debug
// log so operators can read the underlying sub-code. The safe AppError returned
// to callers still hides the cause.
func TestPostLogsRawBodyAtDebugLevel(t *testing.T) {
	const body = `{"error":{"status":"INVALID_ARGUMENT","message":"Unsupported API version"}}`

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	c := newHTTPClient(Credentials{
		DeveloperToken: "x", ClientID: "x", ClientSecret: "x", RefreshToken: "x", CustomerID: "1234567890",
	}, logger)
	c.hc.Transport = fakeRoundTripper{status: http.StatusBadRequest, body: body}
	// Skip OAuth refresh.
	c.token = "test-token"
	c.tokenExp = time.Now().Add(time.Hour)

	_, err := c.post(context.Background(), "generateKeywordIdeas", map[string]any{"keywordSeed": map[string]any{"keywords": []string{"x"}}})

	require.Error(t, err)
	ae := utils.AsAppError(err)
	// The MCP-facing error stays safe; no body leakage there.
	assert.Equal(t, utils.CodeGoogleAdsAPI, ae.Code)
	assert.NotContains(t, ae.Message, "INVALID_ARGUMENT")

	logged := buf.String()
	assert.Contains(t, logged, "google ads api error response")
	assert.Contains(t, logged, "INVALID_ARGUMENT")
	assert.Contains(t, logged, "Unsupported API version")
}

func TestTruncateForLog(t *testing.T) {
	short := []byte("hello")
	assert.Equal(t, "hello", truncateForLog(short))

	big := bytes.Repeat([]byte("a"), maxLoggedBodyBytes+100)
	out := truncateForLog(big)
	assert.Equal(t, maxLoggedBodyBytes+len("...[truncated]"), len(out))
	assert.True(t, strings.HasSuffix(out, "...[truncated]"))
}
