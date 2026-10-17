// Package csrfonly é a defesa ingênua do Capítulo 7: a proteção contra
// CSRF da biblioteca padrão, posta na frente do endpoint MCP.
package csrfonly

import "net/http"

// #region examples-csrfonly-protect
// Protect recusa pedidos de outra origem que mudam estado. Ela compara
// Sec-Fetch-Site ou Origin com o Host do próprio pedido.
func Protect(next http.Handler) http.Handler {
	return http.NewCrossOriginProtection().Handler(next)
}

// #endregion
