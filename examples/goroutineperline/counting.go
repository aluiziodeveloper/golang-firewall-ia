//go:build racedemo

package goroutineperline

import (
	"bufio"
	"io"
)

// #region examples-goroutineperline-counting
// CountingRelay é o Relay com um contador de mensagens compartilhado
// entre as goroutines, sem sincronização.
type CountingRelay struct {
	Messages int
}

// Run copia as linhas e conta cada uma dentro da goroutine.
func (r *CountingRelay) Run(src io.Reader, dst io.Writer) error {
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		line := append([]byte(nil), scanner.Bytes()...)
		go func() {
			r.Messages++
			_, _ = dst.Write(append(line, '\n'))
		}()
	}
	return scanner.Err()
}

// #endregion
