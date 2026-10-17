package app

import (
	"context"
	"errors"
	"testing"

	"github.com/aluiziodeveloper/guarita/internal/domain"
)

// Desafio do capítulo 4: toda decisão é registrada, e a falha do
// registro não muda a decisão.

type spyRecorder struct {
	decisions []domain.Decision
	errs      []error
	fail      error
}

func (r *spyRecorder) Record(
	_ context.Context, _ domain.Call, decision domain.Decision, err error,
) error {
	r.decisions = append(r.decisions, decision)
	r.errs = append(r.errs, err)
	return r.fail
}

func TestInterceptCall_registraCadaDecisao(t *testing.T) {
	for _, decision := range []domain.Decision{
		domain.DecisionAllow, domain.DecisionDeny,
		domain.DecisionRequireApproval, "execute",
	} {
		recorder := &spyRecorder{}
		uc := NewInterceptCall(&stubPolicy{decision: decision},
			&spyUpstream{}, WithAuditRecorder(recorder))

		_, err := uc.Run(context.Background(), validCall)

		if len(recorder.decisions) != 1 {
			t.Fatalf("%s: registros = %d; want 1", decision,
				len(recorder.decisions))
		}
		if !errors.Is(recorder.errs[0], err) && recorder.errs[0] != err {
			t.Fatalf("%s: registro = %v; want o erro de Run %v",
				decision, recorder.errs[0], err)
		}
	}
}

func TestInterceptCall_falhaDoRegistroNaoMudaADecisao(t *testing.T) {
	recorder := &spyRecorder{fail: errors.New("audit disk full")}
	upstream := &spyUpstream{}
	uc := NewInterceptCall(&stubPolicy{decision: domain.DecisionDeny},
		upstream, WithAuditRecorder(recorder))

	_, err := uc.Run(context.Background(), validCall)

	if !errors.Is(err, ErrPolicyDenied) || len(upstream.forwarded) != 0 {
		t.Fatalf("Run() = %v, repasses %d; want negação sem repasse",
			err, len(upstream.forwarded))
	}
}

func TestInterceptCall_semRegistroContinuaFuncionando(t *testing.T) {
	uc := NewInterceptCall(&stubPolicy{decision: domain.DecisionAllow},
		&spyUpstream{})
	if _, err := uc.Run(context.Background(), validCall); err != nil {
		t.Fatalf("Run() = %v; want nil", err)
	}
}
