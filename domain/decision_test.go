package domain

import (
	"errors"
	"testing"
)

func TestDecisionValidate(t *testing.T) {
	tests := []struct {
		name     string
		decision Decision
		wantErr  bool
	}{
		{name: "allow", decision: DecisionAllow},
		{name: "deny", decision: DecisionDeny},
		{name: "require approval", decision: DecisionRequireApproval},
		{name: "empty decision", wantErr: true},
		{name: "unknown decision", decision: Decision("execute"), wantErr: true},
		{name: "case mismatch", decision: Decision("ALLOW"), wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.decision.Validate()
			if !test.wantErr {
				if err != nil {
					t.Fatalf("Validate() returned an error: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("Validate() returned no error")
			}
			if !errors.Is(err, ErrInvalidDecision) {
				t.Errorf("errors.Is(%v, ErrInvalidDecision) = false", err)
			}

			var validationError *ValidationError
			if !errors.As(err, &validationError) {
				t.Fatalf("Validate() error type = %T; want *ValidationError", err)
			}
			if got, want := validationError.Field(), "decision"; got != want {
				t.Errorf("ValidationError.Field() = %q; want %q", got, want)
			}
			if got, want := err.Error(), ErrInvalidDecision.Error(); got != want {
				t.Errorf("Validate() error = %q; want %q", got, want)
			}
		})
	}
}
