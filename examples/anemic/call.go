package anemic

import (
	"errors"
	"strings"
)

// #region examples-anemic-scene
type Call struct {
	ToolName  string
	Arguments string
}

type Tool struct {
	Name        string
	Description string
	InputSchema string
}

type PolicyResult struct {
	Decision string
}

func Validate(call Call, result PolicyResult) error {
	if strings.TrimSpace(call.ToolName) == "" {
		return errors.New("tool name must not be empty")
	}

	return nil
}

// #endregion
