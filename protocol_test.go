package engine

import (
	"context"
	"errors"
	"testing"
)

func TestMatchValidatorImplementations(t *testing.T) {
	t.Parallel()

	// === 验证结构体实现 ===
	wantErr := errors.New("validation failed")
	proposal := MatchProposal{Buyer: MatchParticipant{ParticipantID: "buyer-1"}}
	validator := &recordingMatchValidator{matched: true, err: wantErr}
	matched, err := validator.ValidateMatch(context.Background(), proposal)
	if !matched || !errors.Is(err, wantErr) || validator.proposal != proposal {
		t.Fatalf("struct validator = (%t, %v, %+v)", matched, err, validator.proposal)
	}

	// === 验证函数适配器 ===
	called := false
	var asInterface MatchValidator = MatchValidatorFunc(func(_ context.Context, got MatchProposal) (bool, error) {
		called = true
		if got != proposal {
			t.Fatalf("proposal = %+v, want %+v", got, proposal)
		}
		return true, nil
	})
	matched, err = asInterface.ValidateMatch(context.Background(), proposal)
	if !called || !matched || err != nil {
		t.Fatalf("function validator = called:%t matched:%t error:%v", called, matched, err)
	}
}

func TestNilMatchValidatorFuncReturnsValidationError(t *testing.T) {
	t.Parallel()

	// === 调用 nil 函数适配器 ===
	var validator MatchValidatorFunc
	matched, err := validator.ValidateMatch(context.Background(), MatchProposal{})
	if matched {
		t.Fatal("nil MatchValidatorFunc matched")
	}
	assertValidationError(t, err, "match_validator")
}

type recordingMatchValidator struct {
	matched  bool
	err      error
	proposal MatchProposal
}

func (v *recordingMatchValidator) ValidateMatch(_ context.Context, proposal MatchProposal) (bool, error) {
	v.proposal = proposal
	return v.matched, v.err
}
