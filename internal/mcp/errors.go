package mcp

import "github.com/alparslanahmed/keyword-market-mcp/internal/utils"

// MCPError is the safe, client-facing error shape inside a tool envelope.
type MCPError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}

// Envelope is the standard tool response envelope.
type Envelope struct {
	OK    bool      `json:"ok"`
	Data  any       `json:"data,omitempty"`
	Error *MCPError `json:"error,omitempty"`
}

func successEnvelope(data any) Envelope {
	return Envelope{OK: true, Data: data}
}

func errorEnvelope(err error) Envelope {
	return Envelope{OK: false, Error: ToMCPError(err)}
}

// ToMCPError maps any internal error to a safe MCPError, exposing only the
// stable code, the curated message, and the retryable flag. The underlying
// cause (which may reference credentials or internals) is never included.
func ToMCPError(err error) *MCPError {
	ae := utils.AsAppError(err)
	return &MCPError{
		Code:      string(ae.Code),
		Message:   ae.Message,
		Retryable: ae.Retryable,
	}
}
