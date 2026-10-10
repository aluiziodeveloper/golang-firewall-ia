package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestCallValidate(t *testing.T) {
	tests := []struct {
		name      string
		call      Call
		want      string
		wantField string
		wantKind  error
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
			name:      "empty tool name",
			call:      Call{Arguments: []byte(`{}`)},
			want:      ErrEmptyToolName.Error(),
			wantField: "tool_name",
			wantKind:  ErrEmptyToolName,
		},
		{
			name:      "whitespace tool name",
			call:      Call{ToolName: " \t\n", Arguments: []byte(`{}`)},
			want:      ErrEmptyToolName.Error(),
			wantField: "tool_name",
			wantKind:  ErrEmptyToolName,
		},
		{
			name:      "missing arguments",
			call:      Call{ToolName: "read_file"},
			want:      ErrInvalidArguments.Error(),
			wantField: "arguments",
			wantKind:  ErrInvalidArguments,
		},
		{
			name:      "malformed JSON",
			call:      Call{ToolName: "read_file", Arguments: []byte(`{"path":`)},
			want:      ErrInvalidArguments.Error(),
			wantField: "arguments",
			wantKind:  ErrInvalidArguments,
		},
		{
			name:      "null arguments",
			call:      Call{ToolName: "read_file", Arguments: []byte(`null`)},
			want:      ErrInvalidArguments.Error(),
			wantField: "arguments",
			wantKind:  ErrInvalidArguments,
		},
		{
			name:      "array arguments",
			call:      Call{ToolName: "read_file", Arguments: []byte(`[]`)},
			want:      ErrInvalidArguments.Error(),
			wantField: "arguments",
			wantKind:  ErrInvalidArguments,
		},
		{
			name:      "scalar arguments",
			call:      Call{ToolName: "read_file", Arguments: []byte(`"secret-value"`)},
			want:      ErrInvalidArguments.Error(),
			wantField: "arguments",
			wantKind:  ErrInvalidArguments,
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

			var validationError *ValidationError
			if !errors.As(err, &validationError) {
				t.Fatalf("Validate() error type = %T; want *ValidationError", err)
			}
			if got, want := validationError.Field(), test.wantField; got != want {
				t.Errorf("ValidationError.Field() = %q; want %q", got, want)
			}

			if !errors.Is(err, test.wantKind) {
				t.Errorf("errors.Is(%v, %v) = false", err, test.wantKind)
			}

			if len(test.call.Arguments) > 0 && strings.Contains(err.Error(), string(test.call.Arguments)) {
				t.Errorf("Validate() error contains raw arguments: %q", err)
			}
		})
	}
}
