package mcphttp

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
)

// Erros da guarda e da configuração dela.
var (
	ErrHost         = errors.New("host not allowed")
	ErrOrigin       = errors.New("origin not allowed")
	ErrNoHosts      = errors.New("at least one allowed host is required")
	ErrOriginConfig = errors.New("origin must be scheme://host[:port]")
	ErrNotLoopback  = errors.New("listen address is not loopback")
)

// Guard recusa pedidos cujo Host ou Origin não estejam nas listas. É a
// defesa contra DNS rebinding: depois do rebinding, o navegador manda o
// nome do atacante no Host.
type Guard struct {
	hosts   map[string]bool
	origins map[string]bool
}

// #region mcphttp-new-guard
// NewGuard valida a configuração na partida: um Host pelo menos, e cada
// origem no formato que o navegador envia, sem caminho nem usuário.
func NewGuard(hosts, origins []string) (*Guard, error) {
	guard := &Guard{
		hosts: make(map[string]bool), origins: make(map[string]bool),
	}
	for _, host := range hosts {
		guard.hosts[strings.ToLower(host)] = true
	}
	if len(guard.hosts) == 0 {
		return nil, ErrNoHosts
	}
	for _, origin := range origins {
		u, err := url.Parse(origin)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") ||
			u.Host == "" || u.User != nil || u.Path != "" ||
			u.RawQuery != "" || u.Fragment != "" {
			return nil, fmt.Errorf("%w: %q", ErrOriginConfig, origin)
		}
		guard.origins[u.Scheme+"://"+strings.ToLower(u.Host)] = true
	}
	return guard, nil
}

// #endregion

// #region mcphttp-guard-check
// Check confere o Host de todo pedido e o Origin quando ele vem. Sem
// Origin, o pedido não é de navegador ou é um GET de mesma origem, e o
// Host já decide.
func (g *Guard) Check(r *http.Request) error {
	if !g.hosts[strings.ToLower(r.Host)] {
		return fmt.Errorf("%w: %q", ErrHost, r.Host)
	}
	origin := r.Header.Get("Origin")
	if origin != "" && !g.origins[origin] {
		return fmt.Errorf("%w: %q", ErrOrigin, origin)
	}
	return nil
}

// #endregion

// #region mcphttp-guard-wrap
// Wrap decora next: o que a guarda recusa recebe 403 e não segue.
func (g *Guard) Wrap(next http.Handler) http.Handler {
	return guarded{guard: g, next: next}
}

type guarded struct {
	guard *Guard
	next  http.Handler
}

func (h guarded) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.guard.Check(r); err != nil {
		writeError(w, http.StatusForbidden, codeInvalidRequest,
			err.Error())
		return
	}
	h.next.ServeHTTP(w, r)
}

// #endregion

// #region mcphttp-loopback-hosts
// LoopbackHosts devolve os valores de Host com que um cliente local
// alcança addr. Um endereço fora do loopback exige a lista explícita.
func LoopbackHosts(addr string) ([]string, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	if host == "localhost" {
		return []string{
			"localhost:" + port,
			net.JoinHostPort("127.0.0.1", port),
			net.JoinHostPort("::1", port),
		}, nil
	}
	if ip := net.ParseIP(host); ip == nil || !ip.IsLoopback() {
		return nil, fmt.Errorf("%w: %q", ErrNotLoopback, addr)
	}
	return []string{"localhost:" + port, net.JoinHostPort(host, port)},
		nil
}

// #endregion
