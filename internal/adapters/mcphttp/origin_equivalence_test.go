package mcphttp

import (
	"errors"
	"testing"
)

// Desafio do capítulo 7: o navegador e o operador podem escrever a mesma
// origem de formas diferentes. Maiúsculas no esquema e no nome e a porta
// padrão explícita não mudam a origem; qualquer outra diferença muda.
func TestGuard_comparaOrigensEquivalentes(t *testing.T) {
	guard, err := NewGuard([]string{"127.0.0.1:8080"}, []string{
		"http://localhost:3000", "https://app.example", "http://ui.test:80",
	})
	if err != nil {
		t.Fatalf("NewGuard() = %v", err)
	}
	for _, origin := range []string{
		"http://localhost:3000", "HTTP://LOCALHOST:3000",
		"http://LocalHost:3000", "https://app.example:443",
		"HTTPS://APP.EXAMPLE", "http://ui.test",
	} {
		if err := guard.Check(request("POST", "127.0.0.1:8080",
			origin)); err != nil {
			t.Errorf("origem equivalente %q recusada: %v", origin, err)
		}
	}
	for _, origin := range []string{
		"null", "http://localhost:3000/", "http://localhost:3001",
		"http://localhost", "https://localhost:3000",
		"http://user@localhost:3000", " http://localhost:3000",
		"http://app.example", "https://app.example:8443",
	} {
		err := guard.Check(request("POST", "127.0.0.1:8080", origin))
		if !errors.Is(err, ErrOrigin) {
			t.Errorf("origem %q: %v; want ErrOrigin", origin, err)
		}
	}
}
