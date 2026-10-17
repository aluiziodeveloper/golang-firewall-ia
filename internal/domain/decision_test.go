package domain

import (
	"errors"
	"testing"
)

// #region domain-decision-validate-test
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
		{name: "unknown decision", decision: "execute", wantErr: true},
		{name: "case mismatch", decision: "ALLOW", wantErr: true},
	}

	// #endregion

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
			want := ErrInvalidDecision.Error()
			if got := err.Error(); got != want {
				t.Errorf("Validate() error = %q; want %q", got, want)
			}
		})
	}
}
