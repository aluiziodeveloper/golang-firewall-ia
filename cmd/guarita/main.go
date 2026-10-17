package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/aluiziodeveloper/guarita/internal/adapters/mcpproxy"
)

const usage = "uso: guarita -- <comando do servidor MCP> [argumentos]"

var errUsage = errors.New(usage)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	err := run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	stop()
	if err != nil {
		log.Printf("guarita: %v", err)
		os.Exit(1)
	}
}

// #region cmd-guarita-run
// run é a raiz de composição: monta o servidor e o repasse. As
// entradas e saídas chegam por parâmetro, para o teste trocá-las.
func run(
	ctx context.Context,
	args []string,
	stdin io.ReadCloser,
	stdout io.WriteCloser,
	stderr io.Writer,
) error {
	command, err := serverCommand(args)
	if err != nil {
		return err
	}
	server, serverIO, err := startServer(ctx, command, stderr)
	if err != nil {
		return err
	}
	agent := mcpproxy.Endpoint{Reader: stdin, Writer: stdout}
	stats, relayErr := mcpproxy.Relay(ctx, agent, serverIO, nil)
	waitErr := server.Wait() // só depois que o repasse leu tudo
	_, _ = fmt.Fprintf(stderr, "guarita: %d mensagens do agente, "+
		"%d do servidor\n", stats.AgentToServer, stats.ServerToAgent)
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

// serverCommand exige "--" antes do comando, para que opções futuras da
// Guarita não se confundam com as do servidor.
func serverCommand(args []string) ([]string, error) {
	if len(args) < 2 || args[0] != "--" {
		return nil, errUsage
	}
	return args[1:], nil
}
