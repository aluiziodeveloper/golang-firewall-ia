// Package bufferedproxy é o proxy HTTP escrito à mão do Capítulo 7: copia
// a resposta com io.Copy e, sem perceber, segura os eventos do fluxo.
package bufferedproxy

import (
	"io"
	"net/http"
)

// Proxy repassa cada pedido a Target e copia a resposta.
type Proxy struct {
	Target string
}

// #region examples-bufferedproxy
// ServeHTTP copia a resposta com io.Copy. Os bytes ficam no buffer do
// ResponseWriter até ele encher ou o handler terminar.
func (p Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	out, err := http.NewRequestWithContext(r.Context(), r.Method,
		p.Target, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	out.Header = r.Header.Clone()
	//nolint:gosec // G704: o destino vem da configuração, não do pedido
	response, err := http.DefaultClient.Do(out)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer func() { _ = response.Body.Close() }()
	contentType := response.Header.Get("Content-Type")
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(response.StatusCode)
	_, _ = io.Copy(w, response.Body)
}

// #endregion
