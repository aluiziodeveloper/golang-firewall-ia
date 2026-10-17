package mcphttp

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// fakeMCP responde como o transporte Streamable HTTP: pedido recebe JSON;
// notificação e resposta recebem 202 sem corpo.
func fakeMCP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	if !bytes.Contains(body, []byte(`"id"`)) {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, `{"jsonrpc":"2.0","id":1,"result":{}}`)
}

// newGuarded liga a guarda e o handler a um servidor de teste cujo Host é o
// único aceito, e devolve o endpoint.
func newGuarded(t *testing.T, upstream http.Handler) string {
	t.Helper()
	server := httptest.NewServer(upstream)
	t.Cleanup(server.Close)
	target, _ := url.Parse(server.URL + "/mcp")
	handler, _ := NewHandler(target, allowOnly(t, "read_file"))
	guarita := httptest.NewUnstartedServer(nil)
	guard, err := NewGuard([]string{guarita.Listener.Addr().String()},
		[]string{"http://localhost:3000"})
	if err != nil {
		t.Fatalf("NewGuard() = %v", err)
	}
	guarita.Config.Handler = guard.Wrap(handler)
	guarita.Start()
	t.Cleanup(guarita.Close)
	return guarita.URL + "/mcp"
}

type contract struct {
	name, method, body string
	header             http.Header
	status             int
	contains           string
}

// #region mcphttp-contract-test
func TestHandler_contrato(t *testing.T) {
	endpoint := newGuarded(t, http.HandlerFunc(fakeMCP))
	const list = `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`
	const note = `{"jsonrpc":"2.0",` +
		`"method":"notifications/initialized"}`
	huge := strings.Repeat(" ", maxBodyBytes+1) + list
	evilOrigin := http.Header{"Origin": {"http://evil.test"}}
	evilHost := http.Header{"Host": {"evil.test"}}
	denied := toolCall(2, "delete_file")
	for _, c := range []contract{
		{"pedido", "POST", list, nil, 200, `"result"`},
		{"notificação", "POST", note, nil, 202, ""},
		{"negada", "POST", denied, nil, 200, `"id":2,"error"`},
		{"contrabando", "POST", list + list, nil, 400, `-32700`},
		{"grande demais", "POST", huge, nil, 413, `too large`},
		{"método", "PUT", list, nil, 405, `method not allowed`},
		{"origem", "POST", list, evilOrigin, 403, `origin not allowed`},
		{"host", "POST", list, evilHost, 403, `host not allowed`},
	} {
		t.Run(c.name, func(t *testing.T) { c.check(t, endpoint) })
	}
}

// #endregion

func (c contract) check(t *testing.T, endpoint string) {
	r, _ := http.NewRequest(c.method, endpoint, strings.NewReader(c.body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json, text/event-stream")
	for name, values := range c.header {
		r.Header[name] = values
	}
	if host := c.header.Get("Host"); host != "" {
		r.Host = host
	}
	response, err := http.DefaultClient.Do(r)
	if err != nil {
		t.Fatalf("%s: %v", c.method, err)
	}
	defer func() { _ = response.Body.Close() }()
	body, _ := io.ReadAll(response.Body)
	if response.StatusCode != c.status ||
		!strings.Contains(string(body), c.contains) {
		t.Fatalf("%d %s; want %d com %q", response.StatusCode, body,
			c.status, c.contains)
	}
	if c.status == http.StatusMethodNotAllowed &&
		response.Header.Get("Allow") != "GET, POST, DELETE" {
		t.Fatalf("Allow = %q", response.Header.Get("Allow"))
	}
}

func TestHandler_servidorForaDoAr(t *testing.T) {
	down := httptest.NewServer(http.NotFoundHandler())
	down.Close() // o endereço existe, mas ninguém escuta
	target, _ := url.Parse(down.URL + "/mcp")
	handler, _ := NewHandler(target, allowOnly(t, "read_file"))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("POST", "/mcp",
		strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`)))
	if recorder.Code != http.StatusBadGateway ||
		!strings.Contains(recorder.Body.String(), "upstream unavailable") {
		t.Fatalf("%d %s", recorder.Code, recorder.Body)
	}
}
