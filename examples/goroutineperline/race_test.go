//go:build racedemo

package goroutineperline

import (
	"io"
	"strings"
	"testing"
	"time"
)

// Rode com: go test -race -tags racedemo ./examples/goroutineperline
// A etiqueta mantém a demonstração fora da CI, que roda com -race.
func TestCountingRelay_contadorCompartilhado(t *testing.T) {
	relay := &CountingRelay{}
	input := strings.Repeat("{}\n", 100)
	if err := relay.Run(strings.NewReader(input), io.Discard); err != nil {
		t.Fatalf("Run() = %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	t.Logf("mensagens contadas: %d de 100", relay.Messages)
}
