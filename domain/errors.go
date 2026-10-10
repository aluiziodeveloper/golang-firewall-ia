package domain

import "errors"

var (
	// ErrEmptyToolName identifies a call without a usable tool name.
	ErrEmptyToolName = errors.New("tool name must not be empty")
	// ErrInvalidArguments identifies arguments that are not a JSON object.
	ErrInvalidArguments = errors.New("arguments must be a JSON object")
	// ErrInvalidDecision identifies a decision outside the supported set.
	ErrInvalidDecision = errors.New("decision is not supported")
)

// ValidationError identifies the field that failed domain validation.
type ValidationError struct {
	field string
	kind  error
}

// Error returns a safe validation message without including submitted values.
func (err *ValidationError) Error() string {
	return err.kind.Error()
}

// Unwrap returns the validation category for errors.Is.
func (err *ValidationError) Unwrap() error {
	return err.kind
}

// Field identifies the call field that failed validation.
func (err *ValidationError) Field() string {
	return err.field
}
