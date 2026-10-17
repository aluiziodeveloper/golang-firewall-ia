package mcpproxy

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aluiziodeveloper/guarita/internal/app"
	"github.com/aluiziodeveloper/guarita/internal/domain"
)

// authorizerFunc deixa uma função servir de Authorizer nos testes.
type authorizerFunc func(context.Context, domain.Call) error

func (f authorizerFunc) Authorize(ctx context.Context, c domain.Call) error {
	return f(ctx, c)
}

func allowOnly(tool string) Authorizer {
	return authorizerFunc(func(_ context.Context, call domain.Call) error {
		if call.ToolName == tool {
			return nil
		}
		return app.ErrPolicyDenied
	})
}

func replyCode(t *testing.T, reply []byte) (string, int) {
	t.Helper()
	var response errorResponse
	if err := json.Unmarshal(reply, &response); err != nil {
		t.Fatalf("resposta inválida %q: %v", reply, err)
	}
	return string(response.ID), response.Error.Code
}

// #region mcpproxy-gate-test
func TestGate_decideCadaMensagem(t *testing.T) {
	const head = `{"jsonrpc":"2.0",`
	tests := []struct {
		name    string
		line    string
		forward bool
		id      string
		code    int
	}{
		{
			"tools/list segue", head + `"id":1,"method":"tools/list"}`,
			true, "", 0,
		},
		{"permitida segue", head + `"id":2,"method":"tools/call",` +
			`"params":{"name":"read_file"}}`, true, "", 0},
		{"negada volta", head + `"id":3,"method":"tools/call",` +
			`"params":{"name":"delete_file"}}`, false, "3", codeDenied},
		{
			"contrabando volta", head + `"id":4,"method":"ping"}{}`,
			false, "null", codeParseError,
		},
		{
			"notificação tools/call", head + `"method":"tools/call",` +
				`"params":{"name":"delete_file"}}`, false, "null",
			codeInvalidParams,
		},
	}
	// #endregion
	gate := NewGate(allowOnly("read_file"))
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			forward, reply := gate.Admit(context.Background(),
				[]byte(test.line))
			if forward != test.forward {
				t.Fatalf("forward = %v; want %v", forward, test.forward)
			}
			if test.forward {
				return
			}
			id, code := replyCode(t, reply)
			if id != test.id || code != test.code {
				t.Fatalf("resposta id=%s code=%d; want id=%s code=%d",
					id, code, test.id, test.code)
			}
		})
	}
}
