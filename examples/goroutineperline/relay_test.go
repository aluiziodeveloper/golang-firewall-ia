package goroutineperline

import (
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// stalledWriter imita um servidor que parou de ler a entrada.
type stalledWriter struct {
	release chan struct{}
	done    sync.WaitGroup
}

func (w *stalledWriter) Write(p []byte) (int, error) {
	defer w.done.Done()
	<-w.release
	return len(p), nil
}

// #region examples-goroutineperline-leak-test
// Mil mensagens contra um servidor parado deixam mil goroutines presas.
func TestRelay_acumulaGoroutinesComServidorParado(t *testing.T) {
	const messages = 1000
	writer := &stalledWriter{release: make(chan struct{})}
	writer.done.Add(messages)
	before := runtime.NumGoroutine()

	input := strings.Repeat("{}\n", messages)
	if err := Relay(strings.NewReader(input), writer); err != nil {
		t.Fatalf("Run() = %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	grew := runtime.NumGoroutine() - before
	t.Logf("goroutines a mais depois de Run: %d", grew)

	close(writer.release)
	writer.done.Wait()
	if grew < messages {
		t.Fatalf("goroutines a mais = %d; want pelo menos %d",
			grew, messages)
	}
}

// #endregion
