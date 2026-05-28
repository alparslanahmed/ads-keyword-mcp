package utils

import (
	"net"
	"net/url"
	"strings"
)

// ValidateURL parses raw, enforces http(s), and rejects localhost / private /
// link-local / metadata hosts. It returns a cleaned URL string. This guards the
// case where a URL might ever be fetched directly; for Google Ads we pass URL
// seeds through rather than fetching them.
func ValidateURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", NewValidationInputError("URL is empty.")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", NewValidationInputError("URL is not parseable.")
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
	default:
		return "", NewValidationInputError("URL must use http or https.")
	}
	host := u.Hostname()
	if host == "" {
		return "", NewValidationInputError("URL is missing a host.")
	}
	if isBlockedHost(host) {
		return "", NewValidationInputError("URL host is not allowed.")
	}
	return u.String(), nil
}

func isBlockedHost(host string) bool {
	lower := strings.ToLower(host)
	if lower == "localhost" || strings.HasSuffix(lower, ".localhost") {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return isBlockedIP(ip)
	}
	return false
}

func isBlockedIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsMulticast() || ip.IsPrivate() {
		return true
	}
	// Cloud metadata endpoint.
	if ip.Equal(net.ParseIP("169.254.169.254")) {
		return true
	}
	return false
}

// MustHost extracts a lowercase host for fingerprinting; empty on parse failure.
func MustHost(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Hostname())
}
