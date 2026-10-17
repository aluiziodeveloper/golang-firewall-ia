package bufferedproxy

import (
	"bufio"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// O servidor manda um evento, espera um segundo e manda o último. Pelo
// proxy, o primeiro só chega quando o servidor já terminou.
//
// #region examples-bufferedproxy-test
func TestProxy_seguraOPrimeiroEvento(t *testing.T) {
	const pause = time.Second
	upstream := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = io.WriteString(w, "data: progresso 1/2\n\n")
			_ = http.NewResponseController(w).Flush()
			time.Sleep(pause)
			_, _ = io.WriteString(w, "data: resultado\n\n")
		}))
	t.Cleanup(upstream.Close)
	proxy := httptest.NewServer(Proxy{Target: upstream.URL})
	t.Cleanup(proxy.Close)

	start := time.Now()
	response, err := http.Get(proxy.URL)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = response.Body.Close() }()
	line, _ := bufio.NewReader(response.Body).ReadString('\n')
	elapsed := time.Since(start)
	t.Logf("%q chegou depois de %v", line,
		elapsed.Round(time.Millisecond))
	if elapsed < pause {
		t.Fatalf("o evento chegou antes do fim do fluxo: %v", elapsed)
	}
}

// #endregion
