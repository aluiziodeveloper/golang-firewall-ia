package anemic

import "testing"

// #region examples-anemic-scene-test
func TestValidateReproducesScene(t *testing.T) {
	tests := []struct {
		name   string
		call   Call
		result PolicyResult
	}{
		{
			name:   "malformed JSON arguments",
			call:   Call{ToolName: "read_file", Arguments: `{"path":`},
			result: PolicyResult{Decision: "allow"},
		},
		{
			name:   "unknown decision",
			call:   Call{ToolName: "read_file", Arguments: `{}`},
			result: PolicyResult{Decision: "execute"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := Validate(test.call, test.result); err != nil {
				t.Fatalf("Validate() = %v; the scene accepts this input",
					err)
			}
		})
	}
}

// #endregion
