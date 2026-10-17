package mcpproxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// #region mcpproxy-fuzz-test
// FuzzParseMessage parte das linhas de testdata e muta os bytes. Para
// qualquer entrada, o parser não pode entrar em pânico, e uma mensagem
// aceita, codificada de novo, precisa ser aceita com o mesmo tipo.
func FuzzParseMessage(f *testing.F) {
	addSeeds(f)
	f.Fuzz(func(t *testing.T, data []byte) {
		message, err := ParseMessage(data)
		if err != nil {
			return
		}
		if message.Kind == KindInvalid {
			t.Fatalf("mensagem aceita com Kind inválido: %q", data)
		}
		encoded, err := json.Marshal(message)
		if err != nil {
			t.Fatalf("codificar %q: %v", data, err)
		}
		again, err := ParseMessage(encoded)
		if err != nil || again.Kind != message.Kind {
			t.Fatalf("ida e volta mudou %q: %v", data, err)
		}
	})
}

// #endregion

func addSeeds(f *testing.F) {
	files, _ := filepath.Glob(filepath.Join("testdata", "*.jsonl"))
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			f.Fatalf("ler %s: %v", file, err)
		}
		scanner := bufio.NewScanner(bytes.NewReader(data))
		for scanner.Scan() {
			f.Add(append([]byte(nil), scanner.Bytes()...))
		}
	}
}
