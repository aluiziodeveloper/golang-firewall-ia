package domain

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestCallValidate(t *testing.T) {
	// #region domain-call-validate-cases
	tests := []struct {
		name      string
		call      Call
		want      string
		wantField string
		wantKind  error
	}{
		{
			name: "object arguments",
			call: Call{
				ToolName:      "read_file",
				Arguments:     []byte(`{"path":"notes.txt"}`),
				CorrelationID: "req-1",
			},
		},
		{
			name: "malformed JSON",
			call: Call{
				ToolName:  "read_file",
				Arguments: []byte(`{"path":`),
			},
			want:      ErrInvalidArguments.Error(),
			wantField: "arguments",
			wantKind:  ErrInvalidArguments,
		},
		// #endregion
		{
			name: "empty object arguments",
			call: Call{
				ToolName:      "read_file",
				Arguments:     []byte(`{}`),
				CorrelationID: "req-1",
			},
		},
		{
			name: "nested object arguments",
			call: Call{
				ToolName:      "read_file",
				Arguments:     []byte(`{"options":{"recursive":true}}`),
				CorrelationID: "req-1",
			},
		},
		{
			name: "whitespace around object",
			call: Call{
				ToolName:      "read_file",
				Arguments:     []byte(" \n{}\t "),
				CorrelationID: "req-1",
			},
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
			name:      "trailing JSON value",
			call:      Call{ToolName: "read_file", Arguments: []byte(`{} {}`)},
			want:      ErrInvalidArguments.Error(),
			wantField: "arguments",
			wantKind:  ErrInvalidArguments,
		},
		{
			name: "invalid UTF-8 in JSON",
			call: Call{
				ToolName:  "read_file",
				Arguments: []byte("{\"path\":\"\xff\"}"),
			},
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
			name: "scalar arguments",
			call: Call{
				ToolName:  "read_file",
				Arguments: []byte(`"secret-value"`),
			},
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

			// #region domain-call-validate-checks
			var validationError *ValidationError
			if !errors.As(err, &validationError) {
				t.Fatalf("error type = %T; want *ValidationError", err)
			}
			if got := validationError.Field(); got != test.wantField {
				t.Errorf("Field() = %q; want %q", got, test.wantField)
			}

			if !errors.Is(err, test.wantKind) {
				t.Errorf("errors.Is(%v, %v) = false",
					err, test.wantKind)
			}

			raw := string(test.call.Arguments)
			if raw != "" && strings.Contains(err.Error(), raw) {
				t.Errorf("error contains raw arguments: %q", err)
			}
			// #endregion
		})
	}
}

func TestValidationErrorSupportsWrapping(t *testing.T) {
	err := fmt.Errorf("validate incoming call: %w", &ValidationError{
		field: "arguments",
		kind:  ErrInvalidArguments,
	})

	if !errors.Is(err, ErrInvalidArguments) {
		t.Fatalf("errors.Is(%v, ErrInvalidArguments) = false", err)
	}

	var validationError *ValidationError
	if !errors.As(err, &validationError) {
		t.Fatalf("errors.As(%v, *ValidationError) = false", err)
	}
	if got, want := validationError.Field(), "arguments"; got != want {
		t.Errorf("ValidationError.Field() = %q; want %q", got, want)
	}
}
