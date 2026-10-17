package mcphttp

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// stuckServer lê o pedido, como um servidor MCP real, e nunca responde.
// O canal fecha quando o pedido é cancelado do lado da Guarita.
func stuckServer(t *testing.T) (*url.URL, <-chan struct{}) {
	t.Helper()
	canceled := make(chan struct{})
	stuck := httptest.NewServer(http.HandlerFunc(
		func(_ http.ResponseWriter, r *http.Request) {
			_, _ = io.ReadAll(r.Body)
			<-r.Context().Done()
			close(canceled)
		}))
	t.Cleanup(stuck.Close)
	target, _ := url.Parse(stuck.URL + "/mcp")
	return target, canceled
}

// #region mcphttp-call-timeout-test
// O servidor nunca responde. A Guarita desiste no prazo, responde 504 e
// cancela o pedido que fez ao servidor.
func TestHandler_prazoDaChamada(t *testing.T) {
	target, canceled := stuckServer(t)
	handler, _ := NewHandler(target, allowOnly(t, "read_file"),
		WithCallTimeout(200*time.Millisecond))

	start := time.Now()
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("POST", "/mcp",
		strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`)))
	t.Logf("%d em %v", recorder.Code,
		time.Since(start).Round(10*time.Millisecond))
	if recorder.Code != http.StatusGatewayTimeout {
		t.Fatalf("status = %d; want 504", recorder.Code)
	}
	select {
	case <-canceled:
	case <-time.After(time.Second):
		t.Fatal("o pedido ao servidor não foi cancelado")
	}
}

// #endregion

// #region mcphttp-slow-body-test
// O cliente promete 100 bytes, manda 10 e para. A Guarita responde 408
// no prazo de leitura do corpo, sem esperar para sempre.
func TestHandler_prazoDoCorpo(t *testing.T) {
	target, _ := url.Parse("http://127.0.0.1:1/mcp") // nunca alcançado
	handler, _ := NewHandler(target, allowOnly(t, "read_file"),
		WithBodyTimeout(200*time.Millisecond))
	guarita := httptest.NewServer(handler)
	t.Cleanup(guarita.Close)

	conn, err := net.Dial("tcp", guarita.Listener.Addr().String())
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	start := time.Now()
	_, _ = fmt.Fprintf(conn, "POST /mcp HTTP/1.1\r\nHost: %s\r\n"+
		"Content-Length: 100\r\n\r\n{\"jsonrpc\"",
		guarita.Listener.Addr())
	status, _ := bufio.NewReader(conn).ReadString('\n')
	t.Logf("%q em %v", strings.TrimSpace(status),
		time.Since(start).Round(10*time.Millisecond))
	if !strings.HasPrefix(status, "HTTP/1.1 408") {
		t.Fatalf("status = %q; want 408", status)
	}
}

// #endregion
