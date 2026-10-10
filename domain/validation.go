package domain

import (
	"encoding/json"
	"errors"
	"strings"
)

// Validate checks the call fields that must be valid before policy evaluation.
func (call Call) Validate() error {
	if strings.TrimSpace(call.ToolName) == "" {
		return errors.New("tool name must not be empty")
	}

	var arguments map[string]json.RawMessage
	if err := json.Unmarshal(call.Arguments, &arguments); err != nil || arguments == nil {
		return errors.New("arguments must be a JSON object")
	}

	return nil
}
