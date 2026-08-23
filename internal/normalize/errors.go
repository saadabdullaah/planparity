// Package normalize converts engine-specific JSON into sanitized domain models.
package normalize

import "fmt"

// ErrorCode classifies decoding failures for adapters, reports, and APIs.
type ErrorCode string

const (
	ErrorInvalidJSON       ErrorCode = "INVALID_JSON"
	ErrorMissingValues     ErrorCode = "MISSING_VALUES"
	ErrorUnsupportedFormat ErrorCode = "UNSUPPORTED_FORMAT"
)

// Error is a typed normalization failure that never contains raw input.
type Error struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.Cause }
