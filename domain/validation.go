package domain

import (
	"encoding/json"
	"strings"
)

// Validate checks the call fields that must be valid before policy evaluation.
func (call Call) Validate() error {
	if strings.TrimSpace(call.ToolName) == "" {
		return &ValidationError{field: "tool_name", kind: ErrEmptyToolName}
	}

	var arguments map[string]json.RawMessage
	if err := json.Unmarshal(call.Arguments, &arguments); err != nil || arguments == nil {
		return &ValidationError{field: "arguments", kind: ErrInvalidArguments}
	}

	return nil
}
