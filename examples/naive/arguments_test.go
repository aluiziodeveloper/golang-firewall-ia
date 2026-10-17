package naive

import "testing"

func TestValidateArguments_corrigeACena(t *testing.T) {
	tests := []struct {
		name      string
		arguments string
		wantErr   bool
	}{
		{name: "objeto", arguments: `{"path":"notes.txt"}`},
		{name: "JSON malformado", arguments: `{"path":`, wantErr: true},
		{name: "null", arguments: `null`, wantErr: true},
		{name: "array", arguments: `[]`, wantErr: true},
		{name: "valor residual", arguments: `{} {}`, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateArguments([]byte(test.arguments))
			if (err != nil) != test.wantErr {
				t.Fatalf("ValidateArguments(%q) = %v; wantErr %v",
					test.arguments, err, test.wantErr)
			}
		})
	}
}

// #region examples-naive-bypass-test
// O contorno: json.Valid aceita bytes que não são UTF-8 dentro de uma
// string. O teste fixa a lacuna que o domínio fecha com utf8.Valid.
func TestValidateArguments_aceitaUTF8Invalido(t *testing.T) {
	arguments := []byte("{\"path\":\"notes\xff.txt\"}")

	if err := ValidateArguments(arguments); err != nil {
		t.Fatalf("ValidateArguments() = %v; "+
			"a defesa ingênua não confere UTF-8", err)
	}
}

// #endregion
