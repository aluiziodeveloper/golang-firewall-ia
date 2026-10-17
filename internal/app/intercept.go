// Package app reúne os casos de uso da Guarita e as portas que eles
// consomem. Ele depende só do domínio; os adaptadores dependem dele.
package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/aluiziodeveloper/guarita/internal/domain"
)

// #region app-ports
// PolicyEvaluator decide o que fazer com uma chamada já validada.
type PolicyEvaluator interface {
	Evaluate(
		ctx context.Context, call domain.Call,
	) (domain.Decision, error)
}

// Upstream repassa uma chamada permitida ao servidor MCP real.
type Upstream interface {
	Forward(
		ctx context.Context, call domain.Call,
	) (domain.Result, error)
}

// AuditRecorder registra a decisão tomada para cada chamada.
type AuditRecorder interface {
	Record(
		ctx context.Context, call domain.Call,
		decision domain.Decision, err error,
	) error
}

// #endregion

var (
	// ErrPolicyDenied indica que a política negou a chamada.
	ErrPolicyDenied = errors.New("call denied by policy")
	// ErrApprovalRequired indica que a chamada espera aprovação humana.
	ErrApprovalRequired = errors.New("call requires human approval")
	// ErrPolicyFailed indica que a política não produziu uma decisão
	// válida; a chamada é negada.
	ErrPolicyFailed = errors.New("policy could not decide")
)

// #region app-intercept-type
// InterceptCall é o caso de uso que fica entre o agente e o servidor.
type InterceptCall struct {
	policy   PolicyEvaluator
	upstream Upstream
	recorder AuditRecorder
}

// Option configura uma porta opcional do caso de uso.
type Option func(*InterceptCall)

// WithAuditRecorder registra cada decisão em recorder.
func WithAuditRecorder(recorder AuditRecorder) Option {
	return func(uc *InterceptCall) { uc.recorder = recorder }
}

// NewInterceptCall recebe as portas de que o caso de uso precisa.
func NewInterceptCall(
	policy PolicyEvaluator, upstream Upstream, options ...Option,
) *InterceptCall {
	uc := &InterceptCall{policy: policy, upstream: upstream}
	for _, option := range options {
		option(uc)
	}
	return uc
}

// #endregion

// #region app-intercept-run
// Run decide e registra a decisão. Uma falha do registro não muda a
// decisão: auditoria indisponível não pode liberar nem bloquear.
func (uc *InterceptCall) Run(
	ctx context.Context, call domain.Call,
) (domain.Result, error) {
	result, decision, err := uc.decide(ctx, call)
	if uc.recorder != nil {
		_ = uc.recorder.Record(ctx, call, decision, err)
	}
	return result, err
}

// decide valida, pergunta à política e só repassa com DecisionAllow.
// Qualquer outro resultado, inclusive erro, nega a chamada.
func (uc *InterceptCall) decide(
	ctx context.Context, call domain.Call,
) (domain.Result, domain.Decision, error) {
	deny := domain.DecisionDeny
	if err := call.Validate(); err != nil {
		return domain.Result{}, deny, fmt.Errorf("intercept call: %w", err)
	}
	decision, err := uc.policy.Evaluate(ctx, call)
	if err == nil {
		err = decision.Validate()
	}
	if err != nil {
		return domain.Result{}, deny, fmt.Errorf("intercept call: %w: %w",
			ErrPolicyFailed, err)
	}
	switch decision {
	case domain.DecisionAllow:
		result, err := uc.forward(ctx, call)
		return result, decision, err
	case domain.DecisionRequireApproval:
		return domain.Result{}, decision, ErrApprovalRequired
	case domain.DecisionDeny:
		return domain.Result{}, decision, ErrPolicyDenied
	}
	return domain.Result{}, deny, ErrPolicyFailed
}

// #endregion

func (uc *InterceptCall) forward(
	ctx context.Context, call domain.Call,
) (domain.Result, error) {
	result, err := uc.upstream.Forward(ctx, call)
	if err != nil {
		return domain.Result{}, fmt.Errorf("intercept call: forward: %w",
			err)
	}
	return result, nil
}
