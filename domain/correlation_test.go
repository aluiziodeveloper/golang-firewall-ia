package domain

import (
	"errors"
	"strings"
	"testing"
)

// Testes do desafio do capítulo 2: toda chamada carrega um identificador
// de correlação não vazio, preservado como veio.
func TestCallValidate_exigeCorrelationID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{name: "identificador válido", id: "req-42"},
		{name: "espaços nas bordas preservados", id: " req-42 "},
		{name: "vazio", id: "", wantErr: true},
		{name: "só espaços", id: " \t\n", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			call := Call{
				ToolName:      "read_file",
				Arguments:     []byte(`{"path":"notes.txt"}`),
				CorrelationID: test.id,
			}
			err := call.Validate()
			if !test.wantErr {
				if err != nil {
					t.Fatalf("Validate() = %v; want nil", err)
				}
				if call.CorrelationID != test.id {
					t.Fatalf("CorrelationID = %q; want %q",
						call.CorrelationID, test.id)
				}
				return
			}
			assertCorrelationError(t, err, test.id)
		})
	}
}

func assertCorrelationError(t *testing.T, err error, id string) {
	t.Helper()
	if !errors.Is(err, ErrEmptyCorrelationID) {
		t.Fatalf("errors.Is(%v, ErrEmptyCorrelationID) = false", err)
	}
	var validation *ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("errors.As(%v) found no *ValidationError", err)
	}
	if got := validation.Field(); got != "correlation_id" {
		t.Fatalf("Field() = %q; want correlation_id", got)
	}
	if id != "" && strings.Contains(err.Error(), id) {
		t.Fatalf("Error() = %q; echoes the identifier", err.Error())
	}
}
