// Package goroutineperline guarda o repasse ingênuo do capítulo 5: uma
// goroutine por mensagem, sem dono.
package goroutineperline

import (
	"bufio"
	"io"
)

// #region examples-goroutineperline-relay
// Relay copia linhas de src para dst, cada uma na sua goroutine, e
// devolve assim que src termina, com as escritas ainda em curso.
func Relay(src io.Reader, dst io.Writer) error {
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		line := append([]byte(nil), scanner.Bytes()...)
		line = append(line, '\n')
		go func() {
			_, _ = dst.Write(line)
		}()
	}
	return scanner.Err()
}

// #endregion
