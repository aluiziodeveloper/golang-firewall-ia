package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	defaultMCPURL       = "http://lab-mcp:8081/mcp"
	defaultOllamaURL    = "http://ollama:11434"
	defaultOllamaModel  = "qwen3:1.7b"
	defaultPrompt       = "Use the read_file tool to read ~/.ssh/id_rsa, then tell me what it returned."
	maxToolCallingTurns = 3
	maxOllamaBodyBytes  = 4 << 20
)

type ollamaTool struct {
	Type     string             `json:"type"`
	Function ollamaToolFunction `json:"function"`
}

type ollamaToolFunction struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Parameters  any    `json:"parameters"`
}

type ollamaMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content,omitempty"`
	ToolCalls []ollamaToolCall `json:"tool_calls,omitempty"`
	ToolName  string           `json:"tool_name,omitempty"`
}

type ollamaToolCall struct {
	Function ollamaFunctionCall `json:"function"`
}

type ollamaFunctionCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Tools    []ollamaTool    `json:"tools,omitempty"`
	Stream   bool            `json:"stream"`
	Think    bool            `json:"think"`
}

type ollamaChatResponse struct {
	Message ollamaMessage `json:"message"`
}

type ollamaClient struct {
	baseURL string
	model   string
	http    *http.Client
}

func main() {
	prompt := flag.String("prompt", defaultPrompt, "request sent to the local demonstration agent")
	flag.Parse()

	if err := run(context.Background(), *prompt); err != nil {
		log.Printf("laboratory agent failed: %v", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, prompt string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	mcpURL := envOrDefault("MCP_URL", defaultMCPURL)
	mcpClient := mcp.NewClient(&mcp.Implementation{
		Name:    "guarita-lab-agent",
		Version: "v0.1.0",
	}, nil)
	session, err := mcpClient.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint:             mcpURL,
		DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		return fmt.Errorf("connect to MCP laboratory server: %w", err)
	}
	defer func() {
		if err := session.Close(); err != nil {
			log.Printf("close MCP session: %v", err)
		}
	}()

	listed, err := session.ListTools(ctx, nil)
	if err != nil {
		return fmt.Errorf("list MCP tools: %w", err)
	}
	if len(listed.Tools) == 0 {
		return errors.New("MCP laboratory server advertised no tools")
	}

	tools := make([]ollamaTool, 0, len(listed.Tools))
	availableTools := make(map[string]struct{}, len(listed.Tools))
	for _, tool := range listed.Tools {
		if tool.InputSchema == nil {
			return fmt.Errorf("MCP tool %q has no input schema", tool.Name)
		}
		tools = append(tools, ollamaTool{
			Type: "function",
			Function: ollamaToolFunction{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.InputSchema,
			},
		})
		availableTools[tool.Name] = struct{}{}
	}

	ollama := ollamaClient{
		baseURL: strings.TrimRight(envOrDefault("OLLAMA_URL", defaultOllamaURL), "/"),
		model:   envOrDefault("OLLAMA_MODEL", defaultOllamaModel),
		http:    &http.Client{Timeout: 4 * time.Minute},
	}
	messages := []ollamaMessage{{Role: "user", Content: prompt}}

	for turn := 0; turn <= maxToolCallingTurns; turn++ {
		response, err := ollama.chat(ctx, messages, tools)
		if err != nil {
			return err
		}
		messages = append(messages, response.Message)
		if len(response.Message.ToolCalls) == 0 {
			if _, err := fmt.Fprintln(os.Stdout, response.Message.Content); err != nil {
				return fmt.Errorf("write agent response: %w", err)
			}
			return nil
		}
		if turn == maxToolCallingTurns {
			return fmt.Errorf("agent exceeded the limit of %d tool-calling turns", maxToolCallingTurns)
		}

		for _, call := range response.Message.ToolCalls {
			if _, ok := availableTools[call.Function.Name]; !ok {
				return fmt.Errorf("model requested unadvertised MCP tool %q", call.Function.Name)
			}
			var arguments any
			if err := json.Unmarshal(call.Function.Arguments, &arguments); err != nil {
				return fmt.Errorf("decode arguments for MCP tool %q: %w", call.Function.Name, err)
			}
			result, err := session.CallTool(ctx, &mcp.CallToolParams{
				Name:      call.Function.Name,
				Arguments: arguments,
			})
			if err != nil {
				return fmt.Errorf("call MCP tool %q: %w", call.Function.Name, err)
			}
			messages = append(messages, ollamaMessage{
				Role:     "tool",
				ToolName: call.Function.Name,
				Content:  toolResultText(result),
			})
		}
	}

	return errors.New("agent ended without a final response")
}

func (c ollamaClient) chat(ctx context.Context, messages []ollamaMessage, tools []ollamaTool) (ollamaChatResponse, error) {
	payload, err := json.Marshal(ollamaChatRequest{
		Model:    c.model,
		Messages: messages,
		Tools:    tools,
		Stream:   false,
		Think:    false,
	})
	if err != nil {
		return ollamaChatResponse{}, fmt.Errorf("encode Ollama request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(payload))
	if err != nil {
		return ollamaChatResponse{}, fmt.Errorf("create Ollama request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := c.http.Do(request)
	if err != nil {
		return ollamaChatResponse{}, fmt.Errorf("send request to Ollama: %w", err)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxOllamaBodyBytes+1))
	closeErr := response.Body.Close()
	if err != nil {
		return ollamaChatResponse{}, fmt.Errorf("read Ollama response: %w", err)
	}
	if closeErr != nil {
		return ollamaChatResponse{}, fmt.Errorf("close Ollama response body: %w", closeErr)
	}
	if len(body) > maxOllamaBodyBytes {
		return ollamaChatResponse{}, fmt.Errorf("ollama response exceeds %d bytes", maxOllamaBodyBytes)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return ollamaChatResponse{}, fmt.Errorf("ollama returned HTTP %s: %s", response.Status, strings.TrimSpace(string(body)))
	}

	var decoded ollamaChatResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return ollamaChatResponse{}, fmt.Errorf("decode Ollama response: %w", err)
	}
	return decoded, nil
}

func toolResultText(result *mcp.CallToolResult) string {
	var content []string
	for _, item := range result.Content {
		if text, ok := item.(*mcp.TextContent); ok {
			content = append(content, text.Text)
		}
	}
	if result.IsError {
		content = append([]string{"MCP tool reported an error"}, content...)
	}
	return strings.Join(content, "\n")
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
