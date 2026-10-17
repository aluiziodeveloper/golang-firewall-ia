package mcpproxy

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"testing"
)

// pipes monta os dois lados do repasse com io.Pipe: o teste escreve em
// agentIn, lê agentOut e faz o papel do servidor em serverIn e serverOut.
type pipes struct {
	agent, server Endpoint
	agentIn       *io.PipeWriter
	agentOut      *io.PipeReader
	serverIn      *io.PipeReader
	serverOut     *io.PipeWriter
}

func newPipes() pipes {
	agentReader, agentIn := io.Pipe()
	agentOut, agentWriter := io.Pipe()
	serverIn, serverWriter := io.Pipe()
	serverReader, serverOut := io.Pipe()
	return pipes{
		agent:     Endpoint{Reader: agentReader, Writer: agentWriter},
		server:    Endpoint{Reader: serverReader, Writer: serverWriter},
		agentIn:   agentIn,
		agentOut:  agentOut,
		serverIn:  serverIn,
		serverOut: serverOut,
	}
}

// echoServer responde cada linha com um prefixo e fecha a saída quando
// a entrada termina, como um servidor stdio bem comportado.
func echoServer(in io.Reader, out io.WriteCloser) {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		_, _ = fmt.Fprintf(out, "eco:%s\n", scanner.Text())
	}
	_ = out.Close()
}

func TestRelay_repassaNosDoisSentidos(t *testing.T) {
	p := newPipes()
	go echoServer(p.serverIn, p.serverOut)
	go func() {
		for _, message := range []string{"a", "b", "c"} {
			_, _ = fmt.Fprintln(p.agentIn, message)
		}
		_ = p.agentIn.Close()
	}()
	replies := make(chan []string, 1)
	go func() {
		var got []string
		scanner := bufio.NewScanner(p.agentOut)
		for scanner.Scan() {
			got = append(got, scanner.Text())
		}
		replies <- got
	}()

	stats, err := Relay(context.Background(), p.agent, p.server, nil)
	_ = p.agent.Writer.Close()

	if err != nil {
		t.Fatalf("Relay() = %v; want nil", err)
	}
	if stats != (Stats{AgentToServer: 3, ServerToAgent: 3}) {
		t.Fatalf("Stats = %+v; want 3 e 3", stats)
	}
	if got := <-replies; fmt.Sprint(got) != "[eco:a eco:b eco:c]" {
		t.Fatalf("respostas = %v", got)
	}
}
