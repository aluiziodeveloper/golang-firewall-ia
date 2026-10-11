package domain

import (
	"encoding/json"
	"strings"
	"unicode/utf8"
)

// #region domain-call-validate
// Validate checks the call fields required before policy evaluation.
func (call Call) Validate() error {
	if strings.TrimSpace(call.ToolName) == "" {
		return &ValidationError{field: "tool_name", kind: ErrEmptyToolName}
	}

	if !utf8.Valid(call.Arguments) {
		return &ValidationError{field: "arguments", kind: ErrInvalidArguments}
	}

	var arguments map[string]json.RawMessage
	err := json.Unmarshal(call.Arguments, &arguments)
	if err != nil || arguments == nil {
		return &ValidationError{field: "arguments", kind: ErrInvalidArguments}
	}

	return nil
}

// #endregion
