package mcpproxy

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/aluiziodeveloper/guarita/internal/app"
	"github.com/aluiziodeveloper/guarita/internal/domain"
)

// Códigos de erro JSON-RPC. Os de -32000 a -32099 são reservados para
// erros definidos pela implementação; a Guarita usa -32001.
const (
	codeParseError     = -32700
	codeInvalidRequest = -32600
	codeInvalidParams  = -32602
	codeDenied         = -32001
)

// #region mcpproxy-authorizer
// Authorizer é o que a porta precisa do caso de uso: só a decisão. O
// repasse da mensagem fica com o Relay.
type Authorizer interface {
	Authorize(ctx context.Context, call domain.Call) error
}

// #endregion

var _ Authorizer = (*app.InterceptCall)(nil)

// Gate interpreta cada mensagem do agente e decide se ela segue.
type Gate struct {
	authorizer Authorizer
}

// NewGate recebe o caso de uso que decide as chamadas de ferramenta.
func NewGate(authorizer Authorizer) *Gate {
	return &Gate{authorizer: authorizer}
}

// #region mcpproxy-gate-admit
// Admit devolve true para repassar a mensagem ao servidor. Caso
// contrário, reply traz o erro JSON-RPC que volta ao agente.
func (g *Gate) Admit(
	ctx context.Context, line []byte,
) (forward bool, reply []byte) {
	message, err := ParseMessage(line)
	if err != nil {
		return false, errorReply(nil, parseErrorCode(err), err.Error())
	}
	id := message.ID
	call, isToolCall, err := toolCall(message)
	if err != nil {
		return false, errorReply(id, codeInvalidParams, err.Error())
	}
	if !isToolCall {
		return true, nil
	}
	if err := g.authorizer.Authorize(ctx, call); err != nil {
		return false, errorReply(id, deniedCode(err), err.Error())
	}
	return true, nil
}

// #endregion

func parseErrorCode(err error) int {
	if errors.Is(err, ErrVersion) || errors.Is(err, ErrShape) ||
		errors.Is(err, ErrBatch) {
		return codeInvalidRequest
	}
	return codeParseError
}

func deniedCode(err error) int {
	if errors.Is(err, domain.ErrInvalidArguments) ||
		errors.Is(err, domain.ErrEmptyToolName) {
		return codeInvalidParams
	}
	return codeDenied
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type errorResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Error   rpcError        `json:"error"`
}

// errorReply monta a resposta de erro, com id null quando a mensagem nem
// pôde ser lida.
func errorReply(id json.RawMessage, code int, text string) []byte {
	if len(id) == 0 {
		id = json.RawMessage("null")
	}
	reply, _ := json.Marshal(errorResponse{
		JSONRPC: "2.0", ID: id, Error: rpcError{code, text},
	})
	return append(reply, '\n')
}
