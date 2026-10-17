package mcpproxy

import (
	"bufio"
	"context"
	"fmt"
	"testing"
	"time"
)

// Desafio do capítulo 6: com uma porta configurada, as mensagens do
// servidor também passam pelo parser. As inválidas são descartadas e
// contadas em Stats.Dropped, e a sessão continua.
func TestRelay_descartaMensagemInvalidaDoServidor(t *testing.T) {
	p := newPipes()
	ctx, cancel := context.WithTimeout(context.Background(),
		5*time.Second)
	t.Cleanup(cancel)
	done := make(chan Stats, 1)
	go func() {
		stats, _ := Relay(ctx, p.agent, p.server, NewGate(allowOnly("")))
		done <- stats
	}()
	go func() {
		_, _ = fmt.Fprintln(p.serverOut, `{"jsonrpc":"2.0","id":1,`+
			`"result":{}}{"jsonrpc":"2.0","method":"x"}`)
		_, _ = fmt.Fprintln(p.serverOut, `não é JSON`)
		_, _ = fmt.Fprintln(p.serverOut, `{"jsonrpc":"2.0","id":2,`+
			`"result":{}}`)
		_ = p.serverOut.Close()
	}()

	got, _ := bufio.NewReader(p.agentOut).ReadString('\n')
	if got != `{"jsonrpc":"2.0","id":2,"result":{}}`+"\n" {
		t.Fatalf("primeira mensagem ao agente = %q", got)
	}
	_ = p.agentIn.Close()
	stats := <-done
	if stats.Dropped != 2 || stats.ServerToAgent != 1 {
		t.Fatalf("Stats = %+v; want Dropped 2 e ServerToAgent 1", stats)
	}
}
