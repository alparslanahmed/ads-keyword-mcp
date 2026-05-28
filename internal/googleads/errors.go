package googleads

import (
	"encoding/json"
	"strings"

	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

// classifyAPIError converts an HTTP status + response body from the Google Ads
// REST API into a safe AppError. Raw bodies (which may reference credentials)
// are never embedded in the returned message.
func classifyAPIError(status int, body []byte) *utils.AppError {
	statusName := extractStatus(body)
	switch {
	case status == 401 || status == 403:
		return utils.NewInvalidCredentialsError(nil)
	case status == 429 || statusName == "RESOURCE_EXHAUSTED":
		return utils.NewRateLimitError(nil)
	case status == 400:
		if mentionsCustomerID(body) {
			return utils.NewInvalidCustomerIDError(nil)
		}
		return utils.NewGoogleAdsAPIError("The Google Ads API rejected the request.", false, nil)
	case status >= 500:
		return utils.NewGoogleAdsAPIError("The Google Ads API is temporarily unavailable.", true, nil)
	default:
		return utils.NewGoogleAdsAPIError("The Google Ads API returned an unexpected response.", false, nil)
	}
}

func extractStatus(body []byte) string {
	var parsed struct {
		Error struct {
			Status string `json:"status"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return ""
	}
	return parsed.Error.Status
}

func mentionsCustomerID(body []byte) bool {
	lower := strings.ToLower(string(body))
	return strings.Contains(lower, "customer") &&
		(strings.Contains(lower, "not found") || strings.Contains(lower, "invalid") || strings.Contains(lower, "id"))
}
