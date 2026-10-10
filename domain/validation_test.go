package domain

import (
	"strings"
	"testing"
)

func TestCallValidate(t *testing.T) {
	tests := []struct {
		name string
		call Call
		want string
	}{
		{
			name: "object arguments",
			call: Call{ToolName: "read_file", Arguments: []byte(`{"path":"notes.txt"}`)},
		},
		{
			name: "empty object arguments",
			call: Call{ToolName: "read_file", Arguments: []byte(`{}`)},
		},
		{
			name: "empty tool name",
			call: Call{Arguments: []byte(`{}`)},
			want: "tool name must not be empty",
		},
		{
			name: "whitespace tool name",
			call: Call{ToolName: " \t\n", Arguments: []byte(`{}`)},
			want: "tool name must not be empty",
		},
		{
			name: "missing arguments",
			call: Call{ToolName: "read_file"},
			want: "arguments must be a JSON object",
		},
		{
			name: "malformed JSON",
			call: Call{ToolName: "read_file", Arguments: []byte(`{"path":`)},
			want: "arguments must be a JSON object",
		},
		{
			name: "null arguments",
			call: Call{ToolName: "read_file", Arguments: []byte(`null`)},
			want: "arguments must be a JSON object",
		},
		{
			name: "array arguments",
			call: Call{ToolName: "read_file", Arguments: []byte(`[]`)},
			want: "arguments must be a JSON object",
		},
		{
			name: "scalar arguments",
			call: Call{ToolName: "read_file", Arguments: []byte(`"secret-value"`)},
			want: "arguments must be a JSON object",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.call.Validate()
			if test.want == "" {
				if err != nil {
					t.Fatalf("Validate() returned an error: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("Validate() returned no error; want %q", test.want)
			}
			if got := err.Error(); got != test.want {
				t.Errorf("Validate() error = %q; want %q", got, test.want)
			}
			if len(test.call.Arguments) > 0 && strings.Contains(err.Error(), string(test.call.Arguments)) {
				t.Errorf("Validate() error contains raw arguments: %q", err)
			}
		})
	}
}
