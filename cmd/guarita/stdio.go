package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"

	"github.com/aluiziodeveloper/guarita/internal/adapters/mcpproxy"
)

// #region cmd-guarita-serve-stdio
// serveStdio inicia o servidor como subprocesso e repassa as mensagens
// até ele fechar a saída ou ctx acabar.
func serveStdio(
	ctx context.Context,
	command []string,
	gate mcpproxy.Admitter,
	stdin io.ReadCloser,
	stdout io.WriteCloser,
	stderr io.Writer,
) error {
	server, serverIO, err := startServer(ctx, command, stderr)
	if err != nil {
		return err
	}
	agent := mcpproxy.Endpoint{Reader: stdin, Writer: stdout}
	stats, relayErr := mcpproxy.Relay(ctx, agent, serverIO, gate)
	waitErr := server.Wait() // só depois que o repasse leu tudo
	report(stderr, stats)
	return errors.Join(relayErr, waitErr)
}

// #endregion

// #region cmd-guarita-start-server
// startServer inicia o servidor MCP como subprocesso e devolve as
// pontas da entrada e da saída padrão dele.
func startServer(
	ctx context.Context, command []string, stderr io.Writer,
) (*exec.Cmd, mcpproxy.Endpoint, error) {
	// G204 avisa sobre comando vindo de fora. Aqui ele vem do operador,
	// na linha de comando, nunca de uma mensagem do agente.
	//nolint:gosec // o operador escolhe o servidor
	server := exec.CommandContext(ctx, command[0], command[1:]...)
	server.Stderr = stderr
	var none mcpproxy.Endpoint
	in, err := server.StdinPipe()
	if err != nil {
		return nil, none, fmt.Errorf("server stdin: %w", err)
	}
	out, err := server.StdoutPipe()
	if err != nil {
		return nil, none, fmt.Errorf("server stdout: %w", err)
	}
	if err := server.Start(); err != nil {
		return nil, none, fmt.Errorf("start server: %w", err)
	}
	return server, mcpproxy.Endpoint{Reader: out, Writer: in}, nil
}

// #endregion

func report(stderr io.Writer, stats mcpproxy.Stats) {
	_, _ = fmt.Fprintf(stderr, "guarita: %d do agente, %d do servidor, "+
		"%d recusadas\n", stats.AgentToServer, stats.ServerToAgent,
		stats.Rejected)
}
