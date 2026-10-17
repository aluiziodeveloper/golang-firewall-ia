// Package mcphttp serve o MCP sobre HTTP (Streamable HTTP) na frente de
// um servidor remoto, com a mesma porta do transporte stdio.
package mcphttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/aluiziodeveloper/guarita/internal/adapters/mcpproxy"
)

// maxBodyBytes limita o corpo de um POST, como o limite de mensagem do
// transporte stdio.
const maxBodyBytes = 4 << 20

// Códigos JSON-RPC das respostas que a própria Guarita produz.
const (
	codeInvalidRequest = -32600
	codeInternalError  = -32603
)

// ErrUpstream recusa um endereço de servidor que não seja http ou https.
var ErrUpstream = errors.New("upstream must be an http or https URL")

// #region mcphttp-admitter
// Admitter é a porta do Capítulo 6 vista daqui: decide sobre uma
// mensagem e, quando nega, devolve a resposta de erro.
type Admitter interface {
	Admit(ctx context.Context, message []byte) (bool, []byte)
}

// #endregion

var _ Admitter = (*mcpproxy.Gate)(nil)

// Handler é o endpoint MCP da Guarita: inspeciona o que chega por POST
// e repassa ao servidor pelo proxy reverso.
type Handler struct {
	gate  Admitter
	proxy *httputil.ReverseProxy
}

// #region mcphttp-new-handler
// NewHandler monta o proxy reverso para upstream. Rewrite troca o
// destino e mantém o caminho do servidor, qualquer que seja o recebido.
func NewHandler(upstream *url.URL, gate Admitter) (*Handler, error) {
	if upstream == nil || upstream.Host == "" ||
		(upstream.Scheme != "http" && upstream.Scheme != "https") {
		return nil, ErrUpstream
	}
	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(upstream)
			r.Out.URL.Path = upstream.Path
			r.Out.URL.RawPath = upstream.RawPath
		},
		ErrorHandler: proxyError,
	}
	return &Handler{gate: gate, proxy: proxy}, nil
}

// #endregion

// #region mcphttp-serve
// ServeHTTP atende os três métodos do endpoint MCP. Só o POST leva
// mensagens do agente; GET e DELETE abrem e encerram o fluxo da sessão.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.post(w, r)
	case http.MethodGet, http.MethodDelete:
		h.proxy.ServeHTTP(w, r)
	default:
		w.Header().Set("Allow", "GET, POST, DELETE")
		writeError(w, http.StatusMethodNotAllowed, codeInvalidRequest,
			"method not allowed")
	}
}

// #endregion

// #region mcphttp-post
// post lê o corpo inteiro, até o limite, e só o repassa se a porta
// deixar. O servidor recebe os mesmos bytes que a porta leu.
func (h *Handler) post(w http.ResponseWriter, r *http.Request) {
	limited := http.MaxBytesReader(w, r.Body, maxBodyBytes)
	body, err := io.ReadAll(limited)
	var tooLarge *http.MaxBytesError
	switch {
	case errors.As(err, &tooLarge):
		writeError(w, http.StatusRequestEntityTooLarge,
			codeInvalidRequest, mcpproxy.ErrMessageTooLarge.Error())
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, codeInvalidRequest,
			"cannot read request body")
		return
	}
	if forward, reply := h.gate.Admit(r.Context(), body); !forward {
		writeJSON(w, replyStatus(reply), reply)
		return
	}
	out := r.Clone(r.Context()) // o pedido recebido não é alterado
	out.Body = io.NopCloser(bytes.NewReader(body))
	out.ContentLength = int64(len(body))
	h.proxy.ServeHTTP(w, out)
}

// #endregion

// replyStatus escolhe o código HTTP da recusa: sem id, a mensagem nem
// foi lida (400); com id, é uma resposta JSON-RPC de erro válida (200).
func replyStatus(reply []byte) int {
	var head struct {
		ID json.RawMessage `json:"id"`
	}
	if json.Unmarshal(reply, &head) != nil || len(head.ID) == 0 ||
		string(head.ID) == "null" {
		return http.StatusBadRequest
	}
	return http.StatusOK
}

// #region mcphttp-proxy-error
// proxyError traduz a falha do repasse. Prazo esgotado é 504; servidor
// fora do ar é 502; o agente que desistiu não recebe nada.
func proxyError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(r.Context().Err(), context.Canceled):
		return
	case errors.Is(err, context.DeadlineExceeded):
		writeError(w, http.StatusGatewayTimeout, codeInternalError,
			"upstream timeout")
	default:
		writeError(w, http.StatusBadGateway, codeInternalError,
			"upstream unavailable")
	}
}

// #endregion

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type errorResponse struct {
	JSONRPC string   `json:"jsonrpc"`
	ID      any      `json:"id"`
	Error   rpcError `json:"error"`
}

// writeError responde com um erro JSON-RPC sem id, como a especificação
// prevê para o que a Guarita recusa antes de ler a mensagem.
func writeError(w http.ResponseWriter, status, code int, text string) {
	body, _ := json.Marshal(errorResponse{
		JSONRPC: "2.0", Error: rpcError{code, text},
	})
	writeJSON(w, status, append(body, '\n'))
}

func writeJSON(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
