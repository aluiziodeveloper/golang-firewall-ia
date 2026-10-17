// Package mcpproxy é o adaptador de entrada da Guarita: fala com o agente
// e com o servidor MCP. No capítulo 5, ele só repassa as mensagens.
package mcpproxy

import (
	"bufio"
	"bytes"
	"cmp"
	"context"
	"errors"
	"io"
	"sync"

	"golang.org/x/sync/errgroup"
)

// maxMessageBytes limita uma mensagem; acima dele, ela é recusada.
const maxMessageBytes = 4 << 20

var errServerClosed = errors.New("server closed its output")

// ErrMessageTooLarge recusa uma mensagem acima de maxMessageBytes.
var ErrMessageTooLarge = errors.New("message too large")

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
	Dropped       int
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
		n, err := pump(server.Writer, agent.Reader, keep,
			replyTooLarge(toAgent, &stats.Rejected))
		stats.AgentToServer = n
		_ = server.Writer.Close() // o servidor recebe fim de entrada
		return err
	})
	group.Go(func() error {
		keep := keepAll
		if gate != nil { // com porta, o servidor também passa pelo parser
			keep = keepValid(&stats.Dropped)
		}
		n, err := pump(toAgent, server.Reader, keep, dropTooLarge)
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
) keepFunc {
	return func(message []byte) (bool, error) {
		if gate == nil {
			return true, nil // repasse transparente
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

func keepAll([]byte) (bool, error) { return true, nil }

// keepValid descarta a mensagem do servidor que o parser recusa.
func keepValid(dropped *int) keepFunc {
	return func(message []byte) (bool, error) {
		if _, err := ParseMessage(message); err != nil {
			*dropped++
			return false, nil
		}
		return true, nil
	}
}

// dropTooLarge descarta uma resposta grande demais do servidor; o
// agente espera e recebe o próprio erro de prazo.
func dropTooLarge() error { return nil }

// replyTooLarge conta a mensagem recusada e avisa o agente.
func replyTooLarge(toAgent io.Writer, rejected *int) func() error {
	return func() error {
		*rejected++
		_, err := toAgent.Write(errorReply(nil, codeInvalidRequest,
			ErrMessageTooLarge.Error()))
		return err
	}
}

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

// keepFunc decide se uma mensagem segue para o destino.
type keepFunc func(message []byte) (bool, error)

// pump copia mensagens uma por vez: se dst não lê, pump espera. A
// mensagem acima do limite é descartada, e tooLarge é avisada.
//
// #region mcpproxy-pump
func pump(
	dst io.Writer, src io.Reader, keep keepFunc, tooLarge func() error,
) (int, error) {
	reader := bufio.NewReaderSize(src, 64<<10)
	count, message := 0, []byte(nil)
	for {
		line, err := readMessage(reader, message)
		if errors.Is(err, io.EOF) {
			return count, nil
		}
		forward := false
		switch {
		case errors.Is(err, ErrMessageTooLarge):
			err = tooLarge()
		case err == nil:
			message = line
			forward, err = keep(message)
		}
		if err != nil {
			return count, err
		}
		if forward {
			if _, err := dst.Write(append(message, '\n')); err != nil {
				return count, err
			}
			count++
		}
	}
}

// #endregion

// #region mcpproxy-read-message
// readMessage lê uma linha para buf. Uma linha maior que
// maxMessageBytes é consumida até o fim e devolvida como
// ErrMessageTooLarge, sem crescer a memória além do limite.
func readMessage(reader *bufio.Reader, buf []byte) ([]byte, error) {
	buf = buf[:0]
	for {
		chunk, err := reader.ReadSlice('\n')
		if len(buf)+len(chunk) > maxMessageBytes+1 {
			for errors.Is(err, bufio.ErrBufferFull) {
				_, err = reader.ReadSlice('\n')
			}
			return nil, ErrMessageTooLarge
		}
		buf = append(buf, chunk...)
		if errors.Is(err, bufio.ErrBufferFull) {
			continue // a linha continua no próximo pedaço
		}
		if err != nil && (len(buf) == 0 || !errors.Is(err, io.EOF)) {
			return nil, err
		}
		return bytes.TrimRight(buf, "\r\n"), nil
	}
}

// #endregion
