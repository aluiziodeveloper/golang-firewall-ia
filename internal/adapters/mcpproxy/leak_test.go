package mcpproxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"runtime"
	"testing"
	"time"
)

// #region mcpproxy-leak-helper
// waitForGoroutines falha se, em um segundo, o número de goroutines não
// voltar ao valor medido antes do teste.
func waitForGoroutines(t *testing.T, baseline int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for runtime.NumGoroutine() > baseline {
		if time.Now().After(deadline) {
			t.Fatalf("goroutines = %d; want no máximo %d",
				runtime.NumGoroutine(), baseline)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// #endregion

// #region mcpproxy-cancel-test
// Um servidor que nunca responde não segura o repasse: cancelar o
// contexto encerra as três goroutines.
func TestRelay_paraQuandoOContextoTermina(t *testing.T) {
	baseline := runtime.NumGoroutine()
	p := newPipes()
	go func() { _, _ = io.Copy(io.Discard, p.serverIn) }()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := Relay(ctx, p.agent, p.server)
		done <- err
	}()

	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Relay() = %v; want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Relay não terminou depois do cancelamento")
	}
	_ = p.serverOut.Close()
	_ = p.agentOut.Close()
	waitForGoroutines(t, baseline)
}

// #endregion

// #region mcpproxy-stalled-test
// Contra um servidor que parou de ler, mil mensagens não criam
// goroutine nenhuma: a escrita espera, e quem envia espera junto.
func TestRelay_servidorParadoNaoAcumulaGoroutines(t *testing.T) {
	p := newPipes()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		_, _ = Relay(ctx, p.agent, p.server)
		close(done)
	}()
	time.Sleep(20 * time.Millisecond) // o Relay inicia o próprio grupo
	before := runtime.NumGoroutine()
	go func() {
		for range 1000 {
			if _, err := fmt.Fprintln(p.agentIn, "{}"); err != nil {
				return
			}
		}
	}()
	time.Sleep(50 * time.Millisecond)
	grew := runtime.NumGoroutine() - before
	t.Logf("goroutines a mais com o servidor parado: %d", grew)

	cancel()
	<-done
	if grew > 1 {
		t.Fatalf("goroutines a mais = %d; want no máximo 1", grew)
	}
}

// #endregion
