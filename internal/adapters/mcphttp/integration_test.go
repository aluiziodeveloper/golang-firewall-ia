package mcphttp_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/aluiziodeveloper/guarita/internal/adapters/mcphttp"
	"github.com/aluiziodeveloper/guarita/internal/adapters/mcpproxy"
	"github.com/aluiziodeveloper/guarita/internal/adapters/policystatic"
	"github.com/aluiziodeveloper/guarita/internal/app"
	"github.com/aluiziodeveloper/guarita/internal/domain"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type fileArgs struct {
	Path string `json:"path"`
}

// newUpstream sobe o servidor Streamable HTTP do SDK com duas
// ferramentas que só registram a chamada.
func newUpstream(t *testing.T, executed chan<- string) *url.URL {
	t.Helper()
	server := mcp.NewServer(&mcp.Implementation{Name: "lab"}, nil)
	for _, name := range []string{"read_file", "delete_file"} {
		mcp.AddTool(server, &mcp.Tool{Name: name}, func(
			context.Context, *mcp.CallToolRequest, fileArgs,
		) (*mcp.CallToolResult, any, error) {
			executed <- name
			return &mcp.CallToolResult{Content: []mcp.Content{
				&mcp.TextContent{Text: "ok"},
			}}, nil, nil
		})
	}
	upstream := httptest.NewServer(mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return server }, nil))
	t.Cleanup(upstream.Close)
	target, _ := url.Parse(upstream.URL + "/mcp")
	return target
}

// #region mcphttp-integration-test
func TestHandler_comSDKOficial(t *testing.T) {
	executed := make(chan string, 2)
	policy, _ := policystatic.New(map[string]domain.Decision{
		"read_file": domain.DecisionAllow,
	})
	gate := mcpproxy.NewGate(app.NewInterceptCall(policy, nil))
	handler, err := mcphttp.NewHandler(newUpstream(t, executed), gate)
	if err != nil {
		t.Fatalf("NewHandler() = %v", err)
	}
	guarita := httptest.NewServer(handler)
	t.Cleanup(guarita.Close)

	ctx, cancel := context.WithTimeout(context.Background(),
		10*time.Second)
	t.Cleanup(cancel)
	client := mcp.NewClient(&mcp.Implementation{Name: "agent"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint: guarita.URL + "/mcp",
	}, nil)
	if err != nil {
		t.Fatalf("Connect() = %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	// #endregion

	args := map[string]any{"path": "notes.txt"}
	if _, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "read_file", Arguments: args,
	}); err != nil {
		t.Fatalf("read_file: %v", err)
	}
	_, err = session.CallTool(ctx, &mcp.CallToolParams{
		Name: "delete_file", Arguments: args,
	})
	if err == nil || !strings.Contains(err.Error(), "denied by policy") {
		t.Fatalf("delete_file: err = %v; want negação da política", err)
	}
	t.Logf("delete_file negada: %v", err)
	if got := <-executed; got != "read_file" {
		t.Fatalf("executada = %s; want read_file", got)
	}
	select {
	case got := <-executed:
		t.Fatalf("o servidor executou %s depois da negação", got)
	case <-time.After(100 * time.Millisecond):
	}
}
