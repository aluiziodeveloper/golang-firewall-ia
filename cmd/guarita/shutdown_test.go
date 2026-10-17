package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

// Desafio do capítulo 5: ao cancelar, a Guarita fecha a entrada do
// servidor e espera até shutdownGrace antes de matar o processo.

func runUntilCanceled(
	t *testing.T, command ...string,
) (time.Duration, error) {
	t.Helper()
	stdinReader, stdinWriter := io.Pipe()
	t.Cleanup(func() { _ = stdinWriter.Close() })
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		args := append([]string{"--"}, command...)
		done <- run(ctx, args, stdinReader,
			nopWriteCloser{&bytes.Buffer{}}, io.Discard)
	}()
	time.Sleep(200 * time.Millisecond)
	start := time.Now()
	cancel()
	select {
	case err := <-done:
		return time.Since(start), err
	case <-time.After(5 * time.Second):
		t.Fatal("run não terminou depois do cancelamento")
		return 0, nil
	}
}

func TestRun_encerraOServidorSemMatar(t *testing.T) {
	_, err := runUntilCanceled(t, "sh", "-c", "cat >/dev/null")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("run() = %v; want context.Canceled", err)
	}
	if strings.Contains(err.Error(), "killed") {
		t.Fatalf("run() = %v; o servidor foi morto em vez de encerrar", err)
	}
}

func TestRun_mataQuemIgnoraOFimDepoisDoPrazo(t *testing.T) {
	took, err := runUntilCanceled(t, "sh", "-c",
		"trap '' TERM INT; while :; do sleep 0.1; done")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("run() = %v; want context.Canceled", err)
	}
	if took < shutdownGrace {
		t.Fatalf("encerrou em %v; want esperar pelo menos %v",
			took, shutdownGrace)
	}
}
