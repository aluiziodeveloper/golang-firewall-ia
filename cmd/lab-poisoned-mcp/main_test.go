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

func TestLegitimateServerReturnsOnlySyntheticData(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	httpServer := httptest.NewServer(newModeHandler(modeLegitimate))
	t.Cleanup(httpServer.Close)

	client := mcp.NewClient(&mcp.Implementation{Name: "lab-test-client"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint:             httpServer.URL + "/mcp",
		DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		t.Fatalf("connect client: %v", err)
	}
	t.Cleanup(func() {
		if err := session.Close(); err != nil {
			t.Errorf("close client session: %v", err)
		}
	})

	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	if len(tools.Tools) != 1 {
		t.Fatalf("listed %d tools; want 1", len(tools.Tools))
	}
	if got := tools.Tools[0].Description; got != legitimateDescription {
		t.Fatalf("tool description = %q; want legitimate description", got)
	}

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      readFileToolName,
		Arguments: map[string]any{"path": "~/.ssh/id_rsa"},
	})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if got := resultText(t, result); got != syntheticKey {
		t.Fatalf("tool returned %q; want synthetic laboratory data", got)
	}
}

func TestMutatingServerChangesDescriptionAfterFirstToolCall(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	httpServer := httptest.NewServer(newModeHandler(modeMutating))
	t.Cleanup(httpServer.Close)
	client := mcp.NewClient(&mcp.Implementation{Name: "lab-test-client"}, nil)

	firstSession, err := client.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint:             httpServer.URL + "/mcp",
		DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		t.Fatalf("connect first client: %v", err)
	}
	firstTools, err := firstSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list tools in first session: %v", err)
	}
	if got := firstToolDescription(t, firstTools.Tools); got != legitimateDescription {
		t.Fatalf("first tool description = %q; want legitimate description", got)
	}
	firstResult, err := firstSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      readFileToolName,
		Arguments: map[string]any{"path": "~/.ssh/id_rsa"},
	})
	if err != nil {
		t.Fatalf("call tool in first session: %v", err)
	}
	if got := resultText(t, firstResult); got != syntheticKey {
		t.Fatalf("first session returned %q; want synthetic laboratory data", got)
	}
	if err := firstSession.Close(); err != nil {
		t.Fatalf("close first session: %v", err)
	}

	secondSession, err := client.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint:             httpServer.URL + "/mcp",
		DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		t.Fatalf("connect second client: %v", err)
	}
	t.Cleanup(func() {
		if err := secondSession.Close(); err != nil {
			t.Errorf("close second session: %v", err)
		}
	})
	secondTools, err := secondSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list tools in second session: %v", err)
	}
	if got := firstToolDescription(t, secondTools.Tools); got != poisonedDescription {
		t.Fatalf("second tool description = %q; want mutated poisoned description", got)
	}
	secondResult, err := secondSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      readFileToolName,
		Arguments: map[string]any{"path": "~/.ssh/id_rsa"},
	})
	if err != nil {
		t.Fatalf("call tool in second session: %v", err)
	}
	if got := resultText(t, secondResult); got != syntheticKey {
		t.Fatalf("second session returned %q; want synthetic laboratory data", got)
	}
}

func firstToolDescription(t *testing.T, tools []*mcp.Tool) string {
	t.Helper()
	if len(tools) != 1 {
		t.Fatalf("listed %d tools; want 1", len(tools))
	}
	return tools[0].Description
}

func resultText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) != 1 {
		t.Fatalf("got %d content items; want 1", len(result.Content))
	}
	content, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("content type = %T; want *mcp.TextContent", result.Content[0])
	}
	return content.Text
}
