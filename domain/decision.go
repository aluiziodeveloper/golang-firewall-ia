package domain

// Validate reports whether the decision is one of the supported outcomes.
func (decision Decision) Validate() error {
	switch decision {
	case DecisionAllow, DecisionDeny, DecisionRequireApproval:
		return nil
	default:
		return &ValidationError{field: "decision", kind: ErrInvalidDecision}
	}
}
