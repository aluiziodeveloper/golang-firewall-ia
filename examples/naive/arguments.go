// Package naive guarda a defesa ingênua do capítulo 2: ela corrige a
// cena do validador anêmico, mas deixa passar o contorno do UTF-8.
package naive

import (
	"bytes"
	"encoding/json"
	"errors"
)

// ErrInvalidArguments identifica argumentos recusados pela defesa
// ingênua.
var ErrInvalidArguments = errors.New("arguments must be a JSON object")

// #region examples-naive-arguments
// ValidateArguments confere que os bytes são JSON e começam por um
// objeto. É a primeira correção que costuma aparecer.
func ValidateArguments(arguments []byte) error {
	trimmed := bytes.TrimSpace(arguments)
	if !json.Valid(trimmed) || !bytes.HasPrefix(trimmed, []byte("{")) {
		return ErrInvalidArguments
	}
	return nil
}

// #endregion
