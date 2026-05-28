package googleads

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

const (
	defaultAPIVersion = "v18"
	oauthTokenURL     = "https://oauth2.googleapis.com/token"
	adsBaseURL        = "https://googleads.googleapis.com"
)

// Credentials holds the server-side Google Ads secrets. These never leave the
// process and are never serialized to clients.
type Credentials struct {
	DeveloperToken  string
	ClientID        string
	ClientSecret    string
	RefreshToken    string
	LoginCustomerID string
	CustomerID      string
	APIVersion      string
}

func (c Credentials) validate() error {
	missing := c.DeveloperToken == "" || c.ClientID == "" || c.ClientSecret == "" ||
		c.RefreshToken == "" || c.CustomerID == ""
	if missing {
		return utils.NewMissingCredentialsError(nil)
	}
	return nil
}

// httpClient handles OAuth token refresh and authenticated POSTs to the Google
// Ads REST API.
type httpClient struct {
	creds      Credentials
	apiVersion string
	hc         *http.Client

	mu       sync.Mutex
	token    string
	tokenExp time.Time
}

func newHTTPClient(creds Credentials) *httpClient {
	ver := creds.APIVersion
	if ver == "" {
		ver = defaultAPIVersion
	}
	return &httpClient{
		creds:      creds,
		apiVersion: ver,
		hc:         &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *httpClient) accessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token != "" && time.Now().Before(c.tokenExp) {
		return c.token, nil
	}
	form := url.Values{
		"client_id":     {c.creds.ClientID},
		"client_secret": {c.creds.ClientSecret},
		"refresh_token": {c.creds.RefreshToken},
		"grant_type":    {"refresh_token"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, oauthTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", utils.NewInternalError(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.hc.Do(req)
	if err != nil {
		return "", utils.NewGoogleAdsAPIError("Failed to reach the OAuth token endpoint.", true, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		// Token refresh failure almost always means bad/expired credentials.
		return "", utils.NewInvalidCredentialsError(nil)
	}
	var parsed struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil || parsed.AccessToken == "" {
		return "", utils.NewInvalidCredentialsError(err)
	}
	c.token = parsed.AccessToken
	// Refresh a minute early to avoid edge-of-expiry failures.
	c.tokenExp = time.Now().Add(time.Duration(parsed.ExpiresIn-60) * time.Second)
	return c.token, nil
}

// post performs an authenticated POST to customers/{id}:{method} and returns the
// raw response body, or a classified AppError.
func (c *httpClient) post(ctx context.Context, method string, reqBody any) ([]byte, error) {
	token, err := c.accessToken(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, utils.NewInternalError(err)
	}
	customerID := digitsOnly(c.creds.CustomerID)
	endpoint := adsBaseURL + "/" + c.apiVersion + "/customers/" + customerID + ":" + method

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, utils.NewInternalError(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("developer-token", c.creds.DeveloperToken)
	if c.creds.LoginCustomerID != "" {
		req.Header.Set("login-customer-id", digitsOnly(c.creds.LoginCustomerID))
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, utils.NewGoogleAdsAPIError("Failed to reach the Google Ads API.", true, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, classifyAPIError(resp.StatusCode, body)
	}
	return body, nil
}

func digitsOnly(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
