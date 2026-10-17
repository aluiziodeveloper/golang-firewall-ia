package mcphttp

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

func request(method, host, origin string) *http.Request {
	r := httptest.NewRequest(method, "http://"+host+"/mcp", nil)
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	return r
}

// #region mcphttp-guard-test
// Os dois pedidos do rebinding que passaram por csrfonly agora param.
func TestGuard_recusaORebinding(t *testing.T) {
	const local, evil = "127.0.0.1:8080", "evil.test:8080"
	const lookalike = "http://localhost.evil.test"
	guard, err := NewGuard([]string{local, "localhost:8080"},
		[]string{"http://localhost:3000"})
	if err != nil {
		t.Fatalf("NewGuard() = %v", err)
	}
	tests := []struct {
		req  *http.Request
		want error
	}{
		{request("POST", local, ""), nil},
		{request("POST", local, "http://localhost:3000"), nil},
		{request("POST", evil, "http://"+evil), ErrHost},
		{request("GET", evil, ""), ErrHost},
		{request("POST", local, lookalike), ErrOrigin},
		{request("POST", local, "null"), ErrOrigin},
	}
	for _, test := range tests {
		err := guard.Check(test.req)
		t.Logf("%-4s Host=%-15s Origin=%-27q -> %v", test.req.Method,
			test.req.Host, test.req.Header.Get("Origin"), err)
		if !errors.Is(err, test.want) {
			t.Errorf("Check() = %v; want %v", err, test.want)
		}
	}
}

// #endregion

func TestGuard_recusaAConfiguracaoInvalida(t *testing.T) {
	if _, err := NewGuard(nil, nil); !errors.Is(err, ErrNoHosts) {
		t.Errorf("sem hosts: %v", err)
	}
	for _, origin := range []string{
		"localhost:3000", "http://localhost:3000/", "ftp://x",
		"http://user@localhost", "http://localhost?x=1",
	} {
		_, err := NewGuard([]string{"127.0.0.1:8080"}, []string{origin})
		if !errors.Is(err, ErrOriginConfig) {
			t.Errorf("origem %q aceita: %v", origin, err)
		}
	}
}

func TestGuard_respondeComErroJSONRPC(t *testing.T) {
	guard, _ := NewGuard([]string{"127.0.0.1:8080"}, nil)
	recorder := httptest.NewRecorder()
	guard.Wrap(http.NotFoundHandler()).ServeHTTP(recorder,
		request("POST", "evil.test:8080", ""))
	want := `{"jsonrpc":"2.0","id":null,"error":{"code":-32600,` +
		`"message":"host not allowed: \"evil.test:8080\""}}` + "\n"
	if recorder.Code != http.StatusForbidden || recorder.Body.String() != want {
		t.Fatalf("%d %s", recorder.Code, recorder.Body)
	}
}

func TestLoopbackHosts(t *testing.T) {
	hosts, err := LoopbackHosts("127.0.0.1:8080")
	if err != nil || !slices.Equal(hosts,
		[]string{"localhost:8080", "127.0.0.1:8080"}) {
		t.Fatalf("LoopbackHosts() = %v, %v", hosts, err)
	}
	for _, addr := range []string{"0.0.0.0:8080", ":8080", "10.0.0.5:80"} {
		if _, err := LoopbackHosts(addr); !errors.Is(err, ErrNotLoopback) {
			t.Errorf("LoopbackHosts(%q) = %v", addr, err)
		}
	}
}
