package domain

import (
	"encoding/json"
	"strings"
	"unicode/utf8"
)

// Validate checks the call fields that must be valid before policy evaluation.
func (call Call) Validate() error {
	if strings.TrimSpace(call.ToolName) == "" {
		return &ValidationError{field: "tool_name", kind: ErrEmptyToolName}
	}

	if !utf8.Valid(call.Arguments) {
		return &ValidationError{field: "arguments", kind: ErrInvalidArguments}
	}

	var arguments map[string]json.RawMessage
	if err := json.Unmarshal(call.Arguments, &arguments); err != nil || arguments == nil {
		return &ValidationError{field: "arguments", kind: ErrInvalidArguments}
	}

	return nil
}
