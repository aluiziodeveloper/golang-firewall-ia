package domain_test

import (
	"errors"
	"fmt"

	"github.com/aluiziodeveloper/guarita/domain"
)

// #region domain-example-decision
func ExampleDecision_Validate() {
	known := domain.DecisionDeny
	converted := domain.Decision("execute")

	fmt.Println(known.Validate())
	fmt.Println(converted.Validate())
	// Output:
	// <nil>
	// decision is not supported
}

// #endregion

// #region domain-example-errors
func ExampleValidationError() {
	call := domain.Call{ToolName: "read_file", Arguments: []byte(`[]`)}
	err := fmt.Errorf("intercept call: %w", call.Validate())

	var validation *domain.ValidationError
	if errors.As(err, &validation) {
		fmt.Println("campo:", validation.Field())
	}
	isArguments := errors.Is(err, domain.ErrInvalidArguments)
	fmt.Println("categoria:", isArguments)
	fmt.Println("mensagem:", err)
	// Output:
	// campo: arguments
	// categoria: true
	// mensagem: intercept call: arguments must be a JSON object
}

// #endregion
