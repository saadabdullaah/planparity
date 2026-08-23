package engine

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/saadabdullaah/planparity/internal/model"
	"github.com/saadabdullaah/planparity/internal/runner"
)

// Failure is a typed adapter error with only sanitized detail.
type Failure struct {
	Code      string
	Operation string
	Message   string
	Retryable bool
	Cause     error
}

func (failure *Failure) Error() string {
	return fmt.Sprintf("%s: %s", failure.Code, failure.Message)
}

func (failure *Failure) Unwrap() error { return failure.Cause }

// Public converts a failure to its report-safe representation.
func (failure *Failure) Public() model.PublicFailure {
	return model.PublicFailure{Code: failure.Code, Operation: failure.Operation, Message: failure.Message, Retryable: failure.Retryable}
}

func executionFailure(operation string, err error) error {
	var processError *runner.ExecutionError
	if errors.As(err, &processError) {
		message := "engine command failed"
		if processError.TimedOut {
			message = "engine command timed out"
		}
		return &Failure{Code: "ENGINE_EXECUTION_FAILED", Operation: operation, Message: message, Retryable: processError.TimedOut, Cause: err}
	}
	return &Failure{Code: "ENGINE_EXECUTION_FAILED", Operation: operation, Message: "engine command could not be started", Cause: err}
}

func sanitizeMessage(value string, limit int) string {
	var builder strings.Builder
	for _, r := range value {
		if (r == '\n' || r == '\t' || !unicode.IsControl(r)) && !unicode.Is(unicode.Cf, r) && r != '\x1b' {
			builder.WriteRune(r)
		}
		if builder.Len() >= limit {
			break
		}
	}
	return strings.TrimSpace(builder.String())
}
