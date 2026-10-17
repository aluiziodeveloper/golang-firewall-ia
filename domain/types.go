// Package domain defines the Guarita's core call and policy types.
package domain

import "encoding/json"

// Call is a request to invoke a named tool with its raw JSON arguments.
type Call struct {
	ToolName  string
	Arguments json.RawMessage
	// CorrelationID liga a chamada aos registros que ela produzir.
	CorrelationID string
}

// Tool describes a tool exposed by an MCP server.
type Tool struct {
	Name        string
	Description string
	InputSchema json.RawMessage
}

// Decision is the outcome selected by the Guarita's policy.
type Decision string

const (
	// DecisionAllow permits the call.
	DecisionAllow Decision = "allow"
	// DecisionDeny rejects the call.
	DecisionDeny Decision = "deny"
	// DecisionRequireApproval pauses the call for human approval.
	DecisionRequireApproval Decision = "require_approval"
)
