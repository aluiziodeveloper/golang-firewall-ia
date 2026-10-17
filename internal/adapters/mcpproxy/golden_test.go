package mcpproxy

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// #region mcpproxy-golden-test
var update = flag.Bool("update", false, "regrava os arquivos .golden")

// Cada linha de testdata/*.jsonl vira uma linha de resumo no .golden:
// a mudança de comportamento do parser aparece como diff na revisão.
func TestParseMessage_golden(t *testing.T) {
	inputs, err := filepath.Glob(filepath.Join("testdata", "*.jsonl"))
	if err != nil || len(inputs) == 0 {
		t.Fatalf("testdata sem entradas: %v", err)
	}
	for _, input := range inputs {
		t.Run(filepath.Base(input), func(t *testing.T) {
			got := summarize(t, input)
			golden := strings.TrimSuffix(input, ".jsonl") + ".golden"
			if *update {
				if err := os.WriteFile(golden, got, 0o600); err != nil {
					t.Fatalf("gravar %s: %v", golden, err)
				}
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("ler %s: %v (rode com -update)", golden, err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("%s mudou:\n%s", golden, got)
			}
		})
	}
}

// #endregion

func summarize(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ler %s: %v", path, err)
	}
	var out bytes.Buffer
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for line := 1; scanner.Scan(); line++ {
		message, err := ParseMessage(scanner.Bytes())
		if err != nil {
			fmt.Fprintf(&out, "%d: erro: %v\n", line, err)
			continue
		}
		fmt.Fprintf(&out, "%d: %s id=%s method=%q\n",
			line, message.Kind, message.ID, message.Method)
	}
	return out.Bytes()
}
