package mcpproxy_test

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aluiziodeveloper/guarita/internal/adapters/mcpproxy"
	"github.com/aluiziodeveloper/guarita/internal/adapters/policystatic"
	"github.com/aluiziodeveloper/guarita/internal/app"
	"github.com/aluiziodeveloper/guarita/internal/domain"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type fileArgs struct {
	Path string `json:"path"`
}

// newServer devolve um servidor MCP com read_file e delete_file que só
// registram o que receberam.
func newServer(executed chan<- string) *mcp.Server {
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
	return server
}

// Cliente e servidor do SDK oficial conversam pela Guarita: a chamada
// permitida chega ao servidor, a negada volta como erro e não chega.
func TestRelay_comSDKOficial(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(),
		10*time.Second)
	t.Cleanup(cancel)
	agentR, clientW := io.Pipe()
	clientR, agentW := io.Pipe()
	serverR, relayW := io.Pipe()
	relayR, serverW := io.Pipe()
	executed := make(chan string, 4)

	// #region mcpproxy-integration-test
	go func() {
		_ = newServer(executed).Run(ctx,
			&mcp.IOTransport{Reader: serverR, Writer: serverW})
	}()
	policy, _ := policystatic.New(map[string]domain.Decision{
		"read_file": domain.DecisionAllow,
	})
	gate := mcpproxy.NewGate(app.NewInterceptCall(policy, nil))
	go func() {
		_, _ = mcpproxy.Relay(ctx,
			mcpproxy.Endpoint{Reader: agentR, Writer: agentW},
			mcpproxy.Endpoint{Reader: relayR, Writer: relayW}, gate)
	}()

	client := mcp.NewClient(&mcp.Implementation{Name: "agent"}, nil)
	session, err := client.Connect(ctx,
		&mcp.IOTransport{Reader: clientR, Writer: clientW}, nil)
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
