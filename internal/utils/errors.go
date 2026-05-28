package utils

import "fmt"

// ErrorCode is a stable, safe-to-expose error code returned through MCP responses.
type ErrorCode string

const (
	CodeMissingCredentials ErrorCode = "MISSING_CREDENTIALS"
	CodeInvalidCredentials ErrorCode = "INVALID_GOOGLE_ADS_CREDENTIALS"
	CodeInvalidCustomerID  ErrorCode = "INVALID_CUSTOMER_ID"
	CodeRateLimit          ErrorCode = "RATE_LIMIT"
	CodeGoogleAdsAPI       ErrorCode = "GOOGLE_ADS_API_ERROR"
	CodeNoKeywordData      ErrorCode = "NO_KEYWORD_DATA"
	CodeInvalidTargeting   ErrorCode = "INVALID_TARGETING"
	CodeValidationInput    ErrorCode = "VALIDATION_INPUT"
	CodeCache              ErrorCode = "CACHE_ERROR"
	CodeDatabase           ErrorCode = "DATABASE_ERROR"
	CodeInternal           ErrorCode = "INTERNAL_ERROR"
)

// AppError is the single typed error used across the application. It carries a
// stable Code and a Message that is safe to expose to MCP clients, while the
// underlying cause (Err) is kept internal and never serialized to clients.
type AppError struct {
	Code      ErrorCode
	Message   string
	Retryable bool
	Err       error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Err }

// AsAppError extracts an *AppError from err, returning a generic internal
// error wrapper when err is not already typed.
func AsAppError(err error) *AppError {
	if err == nil {
		return nil
	}
	if ae, ok := err.(*AppError); ok {
		return ae
	}
	return NewInternalError(err)
}

// The constructors below correspond to the typed error taxonomy. Each produces a
// safe, client-facing Message and never embeds secret material.

// NewMissingCredentialsError -> MissingCredentialsError
func NewMissingCredentialsError(cause error) *AppError {
	return &AppError{Code: CodeMissingCredentials, Message: "Google Ads credentials are not configured.", Retryable: false, Err: cause}
}

// NewInvalidCredentialsError -> InvalidGoogleAdsCredentialsError
func NewInvalidCredentialsError(cause error) *AppError {
	return &AppError{Code: CodeInvalidCredentials, Message: "Google Ads credentials were rejected.", Retryable: false, Err: cause}
}

// NewInvalidCustomerIDError -> InvalidCustomerIDError
func NewInvalidCustomerIDError(cause error) *AppError {
	return &AppError{Code: CodeInvalidCustomerID, Message: "The configured Google Ads customer ID is invalid.", Retryable: false, Err: cause}
}

// NewRateLimitError -> RateLimitError
func NewRateLimitError(cause error) *AppError {
	return &AppError{Code: CodeRateLimit, Message: "Rate limit reached for keyword planning calls.", Retryable: true, Err: cause}
}

// NewGoogleAdsAPIError -> GoogleAdsAPIError
func NewGoogleAdsAPIError(message string, retryable bool, cause error) *AppError {
	if message == "" {
		message = "The Google Ads API returned an error."
	}
	return &AppError{Code: CodeGoogleAdsAPI, Message: message, Retryable: retryable, Err: cause}
}

// NewNoKeywordDataError -> NoKeywordDataError
func NewNoKeywordDataError(message string) *AppError {
	if message == "" {
		message = "No keyword data was returned for the given input."
	}
	return &AppError{Code: CodeNoKeywordData, Message: message, Retryable: false}
}

// NewInvalidTargetingError -> InvalidTargetingError
func NewInvalidTargetingError(message string) *AppError {
	if message == "" {
		message = "The requested country/language targeting is not supported."
	}
	return &AppError{Code: CodeInvalidTargeting, Message: message, Retryable: false}
}

// NewValidationInputError -> ValidationInputError
func NewValidationInputError(message string) *AppError {
	if message == "" {
		message = "The provided input is invalid."
	}
	return &AppError{Code: CodeValidationInput, Message: message, Retryable: false}
}

// NewCacheError -> CacheError
func NewCacheError(cause error) *AppError {
	return &AppError{Code: CodeCache, Message: "A cache error occurred.", Retryable: false, Err: cause}
}

// NewDatabaseError -> DatabaseError
func NewDatabaseError(cause error) *AppError {
	return &AppError{Code: CodeDatabase, Message: "A database error occurred.", Retryable: false, Err: cause}
}

// NewInternalError is the catch-all for unexpected failures.
func NewInternalError(cause error) *AppError {
	return &AppError{Code: CodeInternal, Message: "An internal error occurred.", Retryable: false, Err: cause}
}
