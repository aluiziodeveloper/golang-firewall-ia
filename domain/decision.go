package domain

// #region domain-decision-validate
// Validate reports whether the decision is a supported outcome.
func (decision Decision) Validate() error {
	switch decision {
	case DecisionAllow, DecisionDeny, DecisionRequireApproval:
		return nil
	default:
		return &ValidationError{field: "decision", kind: ErrInvalidDecision}
	}
}

// #endregion
