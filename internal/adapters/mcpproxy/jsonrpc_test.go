package mcpproxy

import (
	"errors"
	"strings"
	"testing"
)

func TestParseMessage_aplicaAsRegras(t *testing.T) {
	const head = `{"jsonrpc":"2.0",`
	// #region mcpproxy-parse-test
	tests := []struct {
		name  string
		input string
		kind  Kind
		err   error
	}{
		{"pedido", head + `"id":1,"method":"ping"}`, KindRequest, nil},
		{
			"contrabando", head + `"id":1,"method":"a"}{}`,
			0, ErrTrailingData,
		},
		{
			"chave duplicada", head + `"id":1,"id":2,"method":"a"}`,
			0, ErrDuplicateKey,
		},
		{"lote", `[` + head + `"id":1,"method":"a"}]`, 0, ErrBatch},
		// #endregion
		{
			"notificação", head + `"method":"notifications/x"}`,
			KindNotification, nil,
		},
		{"resposta", head + `"id":"a","result":{}}`, KindResponse, nil},
		{"versão", `{"jsonrpc":"1.0","id":1,"method":"a"}`, 0, ErrVersion},
		{"sem forma", head + `"id":1}`, 0, ErrShape},
		{"cortado", head + `"id":`, 0, ErrMalformed},
		{"UTF-8 inválido", head + "\"method\":\"\xff\"}", 0, ErrNotUTF8},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			message, err := ParseMessage([]byte(test.input))
			if !errors.Is(err, test.err) {
				t.Fatalf("ParseMessage() error = %v; want %v",
					err, test.err)
			}
			if err == nil && message.Kind != test.kind {
				t.Fatalf("Kind = %s; want %s", message.Kind, test.kind)
			}
		})
	}
}

func TestParseMessage_recusaChaveDuplicadaAninhada(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"tools/call",` +
		`"params":{"name":"read_file","name":"delete_file"}}`
	if _, err := ParseMessage([]byte(input)); !errors.Is(err, ErrDuplicateKey) {
		t.Fatalf("ParseMessage() = %v; want ErrDuplicateKey", err)
	}
}

func TestParseMessage_recusaAninhamentoProfundo(t *testing.T) {
	deep := strings.Repeat("[", maxDepth+2) + strings.Repeat("]", maxDepth+2)
	input := `{"jsonrpc":"2.0","id":1,"method":"a","params":` + deep + `}`
	if _, err := ParseMessage([]byte(input)); !errors.Is(err, ErrTooDeep) {
		t.Fatalf("ParseMessage() = %v; want ErrTooDeep", err)
	}
}

func TestToolCall_traduzParaODominio(t *testing.T) {
	message, err := ParseMessage([]byte(`{"jsonrpc":"2.0","id":7,` +
		`"method":"tools/call","params":{"name":"read_file",` +
		`"arguments":{"path":"notes.txt"}}}`))
	if err != nil {
		t.Fatalf("ParseMessage() = %v", err)
	}
	call, ok, err := toolCall(message)
	if err != nil || !ok {
		t.Fatalf("toolCall() = %v, %v", ok, err)
	}
	if call.ToolName != "read_file" ||
		string(call.Arguments) != `{"path":"notes.txt"}` {
		t.Fatalf("Call = %+v", call)
	}
	if err := call.Validate(); err != nil {
		t.Fatalf("Validate() = %v", err)
	}
}
