package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestOllamaChatSendsToolsAndDecodesToolCalls(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("request method = %q; want POST", r.Method)
		}
		if r.URL.Path != "/api/chat" {
			t.Errorf("request path = %q; want /api/chat", r.URL.Path)
		}

		var request ollamaChatRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode request: %v", err)
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		if request.Model != "qwen3:1.7b" {
			t.Errorf("model = %q; want qwen3:1.7b", request.Model)
		}
		if request.Think {
			t.Error("think = true; want false for the lab's non-thinking tool-call flow")
		}
		if len(request.Tools) != 1 || request.Tools[0].Function.Name != "read_file" {
			t.Errorf("request tools = %#v; want one read_file tool", request.Tools)
		}
		if len(request.Messages) != 1 || request.Messages[0].Role != "user" {
			t.Errorf("request messages = %#v; want one user message", request.Messages)
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"message":{"role":"assistant","tool_calls":[{"function":{"name":"read_file","arguments":{"path":"~/.ssh/id_rsa"}}}]}}`)); err != nil {
			t.Errorf("write Ollama response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := ollamaClient{
		baseURL: server.URL,
		model:   "qwen3:1.7b",
		http:    server.Client(),
	}
	response, err := client.chat(context.Background(), []ollamaMessage{
		{Role: "user", Content: "Read the synthetic lab file."},
	}, []ollamaTool{{
		Type: "function",
		Function: ollamaToolFunction{
			Name:        "read_file",
			Description: "read only synthetic lab data",
			Parameters:  map[string]any{"type": "object"},
		},
	}})
	if err != nil {
		t.Fatalf("chat returned error: %v", err)
	}
	if len(response.Message.ToolCalls) != 1 {
		t.Fatalf("got %d tool calls; want 1", len(response.Message.ToolCalls))
	}
	if got := response.Message.ToolCalls[0].Function.Name; got != "read_file" {
		t.Fatalf("tool call name = %q; want read_file", got)
	}
}

func TestOllamaChatReportsHTTPFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "model is not loaded", http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)

	client := ollamaClient{baseURL: server.URL, model: "qwen3:1.7b", http: server.Client()}
	_, err := client.chat(context.Background(), nil, nil)
	if err == nil || !strings.Contains(err.Error(), "HTTP 503 Service Unavailable") {
		t.Fatalf("chat error = %v; want an explicit HTTP 503 error", err)
	}
}

func TestOllamaChatRejectsOversizedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte(strings.Repeat("x", maxOllamaBodyBytes+1)))
		if err != nil {
			t.Errorf("write oversized response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := ollamaClient{baseURL: server.URL, model: "qwen3:1.7b", http: server.Client()}
	_, err := client.chat(context.Background(), nil, nil)
	if err == nil || !strings.Contains(err.Error(), "exceeds 4194304 bytes") {
		t.Fatalf("chat error = %v; want response-size limit error", err)
	}
}

func TestToolResultTextIncludesTextContentAndErrorState(t *testing.T) {
	result := &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: "synthetic result"},
		},
	}

	if got, want := toolResultText(result), "MCP tool reported an error\nsynthetic result"; got != want {
		t.Fatalf("toolResultText() = %q; want %q", got, want)
	}
}
