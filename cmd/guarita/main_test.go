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

// nopWriteCloser deixa um bytes.Buffer servir de saída padrão.
type nopWriteCloser struct{ *bytes.Buffer }

func (nopWriteCloser) Close() error { return nil }

// #region cmd-guarita-e2e-test
// De ponta a ponta: a Guarita inicia `cat` como servidor, repassa três
// mensagens e devolve o eco ao agente.
func TestRun_repassaParaOServidor(t *testing.T) {
	input := "{\"id\":1}\n{\"id\":2}\n{\"id\":3}\n"
	var stdout, stderr bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(),
		5*time.Second)
	t.Cleanup(cancel)

	err := run(ctx, []string{"--", "cat"},
		io.NopCloser(strings.NewReader(input)),
		nopWriteCloser{&stdout}, &stderr)
	if err != nil {
		t.Fatalf("run() = %v; stderr=%s", err, stderr.String())
	}
	if stdout.String() != input {
		t.Fatalf("stdout = %q; want %q", stdout.String(), input)
	}
	if !strings.Contains(stderr.String(), "3 mensagens do agente") {
		t.Fatalf("stderr = %q; want a contagem", stderr.String())
	}
}

// #endregion

func TestRun_exigeOSeparador(t *testing.T) {
	for _, args := range [][]string{nil, {"cat"}, {"--"}} {
		err := run(context.Background(), args,
			io.NopCloser(strings.NewReader("")),
			nopWriteCloser{&bytes.Buffer{}}, io.Discard)
		if !errors.Is(err, errUsage) {
			t.Errorf("run(%q) = %v; want errUsage", args, err)
		}
	}
}

func TestRun_encerraComOContexto(t *testing.T) {
	stdinReader, stdinWriter := io.Pipe()
	t.Cleanup(func() { _ = stdinWriter.Close() })
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- run(ctx, []string{"--", "cat"}, stdinReader,
			nopWriteCloser{&bytes.Buffer{}}, io.Discard)
	}()
	_, _ = io.WriteString(stdinWriter, "{}\n")
	time.Sleep(100 * time.Millisecond)

	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("run() = %v; want context.Canceled", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("run não terminou depois do cancelamento")
	}
}
