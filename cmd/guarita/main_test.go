package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
)

func ping(id int) string {
	return fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"method":"ping"}`+"\n", id)
}

// nopWriteCloser deixa um bytes.Buffer servir de saída padrão.
type nopWriteCloser struct{ *bytes.Buffer }

func (nopWriteCloser) Close() error { return nil }

// #region cmd-guarita-e2e-test
// De ponta a ponta: a Guarita inicia `cat` como servidor, repassa três
// mensagens e devolve o eco ao agente.
func TestRun_repassaParaOServidor(t *testing.T) {
	input := ping(1) + ping(2) + ping(3)
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
	if !strings.Contains(stderr.String(), "3 do agente") {
		t.Fatalf("stderr = %q; want a contagem", stderr.String())
	}
}

// #endregion

func TestRun_exigeOSeparador(t *testing.T) {
	for _, args := range [][]string{
		nil, {"cat"}, {"--"}, {"-nope", "--", "cat"},
	} {
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
	_, _ = io.WriteString(stdinWriter, ping(1))
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

// #region cmd-guarita-allow-test
// Com -allow read_file, a chamada a delete_file volta ao agente como
// erro e não chega ao servidor; a permitida chega e volta pelo cat.
func TestRun_negaOQueNaoFoiPermitido(t *testing.T) {
	call := func(id int, tool string) string {
		return fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,`+
			`"method":"tools/call","params":{"name":%q}}`+"\n",
			id, tool)
	}
	input := call(1, "read_file") + call(2, "delete_file")
	var stdout, stderr bytes.Buffer

	err := run(context.Background(),
		[]string{"-allow", "read_file", "--", "cat"},
		io.NopCloser(strings.NewReader(input)),
		nopWriteCloser{&stdout}, &stderr)
	if err != nil {
		t.Fatalf("run() = %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, call(1, "read_file")) {
		t.Fatalf("a chamada permitida não voltou do servidor: %q", out)
	}
	if strings.Contains(out, `"name":"delete_file"`) ||
		!strings.Contains(out, `"id":2,"error"`) {
		t.Fatalf("a chamada negada chegou ao servidor: %q", out)
	}
}

// #endregion
