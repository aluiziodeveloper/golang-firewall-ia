package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/aluiziodeveloper/guarita/internal/adapters/mcphttp"
)

// httpShutdownGrace é quanto o encerramento espera as chamadas em curso.
const httpShutdownGrace = 5 * time.Second

// #region cmd-guarita-serve-http
// serveHTTP escuta, monta o endpoint e atende até ctx acabar. Aí para
// de aceitar conexões e espera as chamadas em curso, com prazo.
func serveHTTP(
	ctx context.Context, c config, gate mcphttp.Admitter,
	stderr io.Writer,
) error {
	listener, err := net.Listen("tcp", c.listen)
	if err != nil {
		return err
	}
	endpoint, err := newEndpoint(c, gate, listener.Addr().String())
	if err != nil {
		_ = listener.Close()
		return err
	}
	server := &http.Server{
		Handler:           endpoint,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       time.Minute,
		MaxHeaderBytes:    64 << 10,
		// Sem WriteTimeout: ele cortaria o fluxo longo do GET. O prazo
		// de cada POST fica no handler.
	}
	_, _ = fmt.Fprintf(stderr, "guarita: escutando em http://%s/mcp\n",
		listener.Addr())
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	return shutdown(ctx, server, served)
}

// #endregion

// #region cmd-guarita-shutdown
// shutdown espera o fim de ctx ou do servidor. Shutdown não interrompe
// fluxos abertos; esgotado o prazo, Close os corta.
func shutdown(
	ctx context.Context, server *http.Server, served <-chan error,
) error {
	select {
	case err := <-served:
		return err
	case <-ctx.Done():
	}
	grace, cancel := context.WithTimeout(context.WithoutCancel(ctx),
		httpShutdownGrace)
	defer cancel()
	err := server.Shutdown(grace)
	if err != nil {
		_ = server.Close()
	}
	<-served // a goroutine de Serve termina antes do retorno
	return errors.Join(ctx.Err(), err)
}

// #endregion

// #region cmd-guarita-new-endpoint
// newEndpoint liga a guarda ao handler em /mcp. Sem -host, os hosts
// aceitos saem do endereço em que a Guarita de fato escuta.
func newEndpoint(
	c config, gate mcphttp.Admitter, addr string,
) (http.Handler, error) {
	upstream, err := url.Parse(c.upstream)
	if err != nil {
		return nil, err
	}
	handler, err := mcphttp.NewHandler(upstream, gate,
		mcphttp.WithCallTimeout(c.callTimeout))
	if err != nil {
		return nil, err
	}
	hosts := c.hosts
	if len(hosts) == 0 {
		if hosts, err = mcphttp.LoopbackHosts(addr); err != nil {
			return nil, err // fora do loopback, -host é obrigatório
		}
	}
	guard, err := mcphttp.NewGuard(hosts, c.origins)
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.Handle("/mcp", guard.Wrap(handler))
	return mux, nil
}

// #endregion
