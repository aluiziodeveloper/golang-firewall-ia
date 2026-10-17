package mcpproxy

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

// Uma mensagem acima do limite é recusada sozinha: o agente recebe um
// erro, e a mensagem seguinte ainda chega ao servidor.
func TestRelay_mensagemGrandeNaoDerrubaASessao(t *testing.T) {
	p := newPipes()
	go echoServer(p.serverIn, p.serverOut)
	ctx, cancel := context.WithTimeout(context.Background(),
		5*time.Second)
	t.Cleanup(cancel)
	done := make(chan error, 1)
	go func() {
		_, err := Relay(ctx, p.agent, p.server, nil)
		done <- err
	}()
	// #region mcpproxy-oversize-test
	go func() {
		huge := strings.Repeat("x", maxMessageBytes+1)
		_, _ = fmt.Fprintln(p.agentIn, huge)
		_, _ = fmt.Fprintln(p.agentIn, "depois")
		_ = p.agentIn.Close()
	}()

	replies := bufio.NewReader(p.agentOut)
	first, _ := replies.ReadString('\n')
	if first == "" {
		t.Fatalf("a sessão caiu: Relay() = %v", <-done)
	}
	second, _ := replies.ReadString('\n')
	if !strings.Contains(first, "too large") ||
		second != "eco:depois\n" {
		t.Fatalf("respostas = %.80q, %q", first, second)
	}
	if err := <-done; err != nil {
		t.Fatalf("Relay() = %v", err)
	}
}

// #endregion
