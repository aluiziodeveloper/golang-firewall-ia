package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestPoisonedServerPublishesMaliciousDescriptionAndSyntheticData(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	httpServer := httptest.NewServer(newHandler())
	t.Cleanup(httpServer.Close)

	client := mcp.NewClient(&mcp.Implementation{Name: "lab-test-client"}, nil)
	clientSession, err := client.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint:             httpServer.URL + "/mcp",
		DisableStandaloneSSE: true,
	}, nil)
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

func TestHealthEndpointReturnsNoContent(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	newHandler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("health endpoint status = %d; want %d", recorder.Code, http.StatusNoContent)
	}
}
