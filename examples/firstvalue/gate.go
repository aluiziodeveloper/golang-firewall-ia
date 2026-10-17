// Package firstvalue guarda a porta ingênua do capítulo 6: ela decide
// pela primeira mensagem da linha e repassa a linha inteira.
package firstvalue

import (
	"bytes"
	"encoding/json"
)

type message struct {
	Method string `json:"method"`
	Params struct {
		Name string `json:"name"`
	} `json:"params"`
}

// #region examples-firstvalue-admit
// Admit lê um valor e decide; o resto da linha nem é olhado.
func Admit(line []byte, allowed map[string]bool) bool {
	var first message
	decoder := json.NewDecoder(bytes.NewReader(line))
	if err := decoder.Decode(&first); err != nil {
		return false
	}
	if first.Method != "tools/call" {
		return true
	}
	return allowed[first.Params.Name]
}

// #endregion

// #region examples-firstvalue-server
// Executed imita um servidor que lê o fluxo valor a valor, como
// json.Decoder faz, e devolve as ferramentas que ele executaria.
func Executed(stream []byte) []string {
	var tools []string
	decoder := json.NewDecoder(bytes.NewReader(stream))
	for {
		var next message
		if err := decoder.Decode(&next); err != nil {
			return tools // io.EOF ou lixo: o fluxo acabou
		}
		if next.Method == "tools/call" {
			tools = append(tools, next.Params.Name)
		}
	}
}

// #endregion
