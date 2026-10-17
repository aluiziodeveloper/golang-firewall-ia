package csrfonly

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// browser monta o pedido como o navegador o envia, segundo o padrão
// Fetch: o POST leva Origin; o GET de mesma origem, não.
func browser(method, host, origin, site string) *http.Request {
	r := httptest.NewRequest(method, "http://"+host+"/mcp",
		strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	r.Header.Set("Sec-Fetch-Site", site)
	return r
}

// #region examples-csrfonly-test
func TestProtect_deixaPassarORebinding(t *testing.T) {
	// O endpoint não escreve nada: 200 quer dizer que o pedido chegou.
	handler := Protect(http.HandlerFunc(
		func(http.ResponseWriter, *http.Request) {}))
	const local, evil = "127.0.0.1:8080", "evil.test:8080"
	tests := []struct {
		name string
		req  *http.Request
		want int
	}{{
		"CSRF de outro site",
		browser("POST", local, "http://evil.test", "cross-site"), 403,
	}, {
		"POST depois do rebinding",
		browser("POST", evil, "http://"+evil, "same-origin"), 200,
	}, {
		"GET depois do rebinding",
		browser("GET", evil, "", "same-origin"), 200,
	}}
	for _, test := range tests {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, test.req)
		t.Logf("%-26s -> %d", test.name, recorder.Code)
		if recorder.Code != test.want {
			t.Errorf("%s = %d", test.name, recorder.Code)
		}
	}
}

// #endregion
