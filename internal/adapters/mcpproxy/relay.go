// Package mcpproxy é o adaptador de entrada da Guarita: fala com o agente
// e com o servidor MCP. No capítulo 5, ele só repassa as mensagens.
package mcpproxy

import (
	"bufio"
	"cmp"
	"context"
	"errors"
	"io"

	"golang.org/x/sync/errgroup"
)

// maxMessageBytes limita uma mensagem; o capítulo 6 trata o excesso.
const maxMessageBytes = 4 << 20

var errServerClosed = errors.New("server closed its output")

// Endpoint é um lado do repasse: de onde lê e para onde escreve.
type Endpoint struct {
	Reader io.ReadCloser
	Writer io.WriteCloser
}

// Stats conta as mensagens de cada sentido.
type Stats struct {
	AgentToServer int
	ServerToAgent int
}

// #region mcpproxy-relay
// Relay repassa mensagens até o servidor fechar a saída ou ctx acabar.
// As três goroutines pertencem ao grupo e terminam antes do retorno.
func Relay(ctx context.Context, agent, server Endpoint) (Stats, error) {
	group, groupCtx := errgroup.WithContext(ctx)
	var stats Stats
	group.Go(func() error {
		n, err := pump(server.Writer, agent.Reader)
		stats.AgentToServer = n
		_ = server.Writer.Close() // o servidor recebe fim de entrada
		return err
	})
	group.Go(func() error {
		n, err := pump(agent.Writer, server.Reader)
		stats.ServerToAgent = n
		return cmp.Or(err, errServerClosed)
	})
	group.Go(func() error {
		<-groupCtx.Done()
		closeAll(agent, server) // desbloqueia Read e Write parados
		return nil
	})
	return stats, outcome(ctx, group.Wait())
}

// #endregion

// outcome traduz o fim do grupo: cancelamento vence, e o servidor que
// fecha a saída é o fim normal do repasse.
func outcome(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if errors.Is(err, errServerClosed) {
		return nil
	}
	return err
}

func closeAll(endpoints ...Endpoint) {
	for _, endpoint := range endpoints {
		_ = endpoint.Reader.Close()
		_ = endpoint.Writer.Close()
	}
}

// #region mcpproxy-pump
// pump copia mensagens delimitadas por linha, uma por vez: se dst não
// lê, pump espera, e nenhuma goroutine nova é criada.
func pump(dst io.Writer, src io.Reader) (int, error) {
	scanner := bufio.NewScanner(src)
	scanner.Buffer(make([]byte, 64<<10), maxMessageBytes)
	var message []byte
	count := 0
	for scanner.Scan() {
		// Bytes aponta para o buffer do Scanner: copie antes de mexer.
		message = append(message[:0], scanner.Bytes()...)
		message = append(message, '\n')
		if _, err := dst.Write(message); err != nil {
			return count, err
		}
		count++
	}
	return count, scanner.Err()
}

// #endregion
