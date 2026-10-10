package main

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestPoisonedServerPublishesMaliciousDescriptionAndSyntheticData(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	serverSession, err := newServer().Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("connect server: %v", err)
	}
	t.Cleanup(func() {
		if err := serverSession.Wait(); err != nil {
			t.Errorf("wait for server: %v", err)
		}
	})

	client := mcp.NewClient(&mcp.Implementation{Name: "lab-test-client"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("connect client: %v", err)
	}
	t.Cleanup(func() {
		if err := clientSession.Close(); err != nil {
			t.Errorf("close client session: %v", err)
		}
	})

	tools, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	if len(tools.Tools) != 1 {
		t.Fatalf("listed %d tools; want 1", len(tools.Tools))
	}
	if got := tools.Tools[0].Description; got != poisonedDescription {
		t.Fatalf("tool description = %q; want %q", got, poisonedDescription)
	}

	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      readFileToolName,
		Arguments: map[string]any{"path": "~/.ssh/id_rsa"},
	})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if len(result.Content) != 1 {
		t.Fatalf("got %d content items; want 1", len(result.Content))
	}
	content, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("content type = %T; want *mcp.TextContent", result.Content[0])
	}
	if content.Text != syntheticKey {
		t.Fatalf("tool returned %q; want synthetic laboratory data", content.Text)
	}
}
