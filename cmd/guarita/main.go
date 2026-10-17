package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"slices"
	"strings"
	"syscall"

	"github.com/aluiziodeveloper/guarita/internal/adapters/mcpproxy"
	"github.com/aluiziodeveloper/guarita/internal/adapters/policystatic"
	"github.com/aluiziodeveloper/guarita/internal/app"
	"github.com/aluiziodeveloper/guarita/internal/domain"
)

const usage = "uso: guarita [-allow ferramenta,…] -- " +
	"<comando do servidor MCP> [argumentos]"

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
// run é a raiz de composição: lê a configuração, monta a política, o
// caso de uso, a porta, o servidor e o repasse.
func run(
	ctx context.Context,
	args []string,
	stdin io.ReadCloser,
	stdout io.WriteCloser,
	stderr io.Writer,
) error {
	config, err := parseConfig(args, stderr)
	if err != nil {
		return err
	}
	policy, err := policystatic.New(config.decisions)
	if err != nil {
		return err
	}
	// O stdio só usa Authorize: o repasse fica com o Relay.
	gate := mcpproxy.NewGate(app.NewInterceptCall(policy, nil))
	server, serverIO, err := startServer(ctx, config.command, stderr)
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

type config struct {
	decisions map[string]domain.Decision
	command   []string
}

// #region cmd-guarita-parse-config
// parseConfig separa as opções da Guarita, antes de "--", do comando do
// servidor, depois dele. Sem -allow, nenhuma ferramenta é permitida.
func parseConfig(args []string, stderr io.Writer) (config, error) {
	separator := slices.Index(args, "--")
	if separator < 0 || separator == len(args)-1 {
		return config{}, errUsage
	}
	flags := flag.NewFlagSet("guarita", flag.ContinueOnError)
	flags.SetOutput(stderr)
	allow := flags.String("allow", "",
		"ferramentas permitidas, separadas por vírgula")
	if err := flags.Parse(args[:separator]); err != nil {
		return config{}, errors.Join(errUsage, err)
	}
	decisions := make(map[string]domain.Decision)
	for tool := range strings.SplitSeq(*allow, ",") {
		if tool = strings.TrimSpace(tool); tool != "" {
			decisions[tool] = domain.DecisionAllow
		}
	}
	return config{decisions, args[separator+1:]}, nil
}

// #endregion
