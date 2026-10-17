// Package mcpproxy é o adaptador de entrada da Guarita: fala com o agente
// e com o servidor MCP. No capítulo 5, ele só repassa as mensagens.
package mcpproxy

import (
	"bufio"
	"cmp"
	"context"
	"errors"
	"io"
	"sync"

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
	Rejected      int
}

// Admitter decide, mensagem a mensagem, o que segue do agente para o
// servidor. Uma mensagem barrada pode ter uma resposta para o agente.
type Admitter interface {
	Admit(ctx context.Context, line []byte) (forward bool, reply []byte)
}

// #region mcpproxy-relay
// Relay repassa mensagens até o servidor fechar a saída ou ctx acabar.
// As três goroutines pertencem ao grupo e terminam antes do retorno.
func Relay(
	ctx context.Context, agent, server Endpoint, gate Admitter,
) (Stats, error) {
	group, groupCtx := errgroup.WithContext(ctx)
	// Duas goroutines escrevem para o agente: a porta e o servidor.
	toAgent := &lockedWriter{w: agent.Writer}
	var stats Stats
	group.Go(func() error {
		keep := admit(groupCtx, gate, toAgent, &stats.Rejected)
		n, err := pump(server.Writer, agent.Reader, keep)
		stats.AgentToServer = n
		_ = server.Writer.Close() // o servidor recebe fim de entrada
		return err
	})
	group.Go(func() error {
		n, err := pump(toAgent, server.Reader, nil)
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

// #region mcpproxy-admit
// admit liga a porta à cópia do agente: o que a porta barra não segue,
// e a resposta dela vai direto para o agente.
func admit(
	ctx context.Context,
	gate Admitter,
	toAgent io.Writer,
	rejected *int,
) func([]byte) (bool, error) {
	return func(message []byte) (bool, error) {
		if gate == nil {
			return true, nil
		}
		forward, reply := gate.Admit(ctx, message)
		if forward {
			return true, nil
		}
		*rejected++
		if len(reply) == 0 {
			return false, nil
		}
		_, err := toAgent.Write(reply)
		return false, err
	}
}

// #endregion

// #region mcpproxy-locked-writer
// lockedWriter serializa as escritas: sem ele, a resposta da porta e a
// do servidor poderiam se misturar na mesma linha da saída do agente.
type lockedWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (l *lockedWriter) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w.Write(p)
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
// lê, pump espera, e nenhuma goroutine nova é criada. keep, se houver,
// decide se cada mensagem segue.
func pump(
	dst io.Writer, src io.Reader, keep func([]byte) (bool, error),
) (int, error) {
	scanner := bufio.NewScanner(src)
	scanner.Buffer(make([]byte, 64<<10), maxMessageBytes)
	var message []byte
	count := 0
	for scanner.Scan() {
		// Bytes aponta para o buffer do Scanner: copie antes de mexer.
		message = append(message[:0], scanner.Bytes()...)
		if keep != nil {
			if ok, err := keep(message); !ok || err != nil {
				if err != nil {
					return count, err
				}
				continue
			}
		}
		if _, err := dst.Write(append(message, '\n')); err != nil {
			return count, err
		}
		count++
	}
	return count, scanner.Err()
}

// #endregion
