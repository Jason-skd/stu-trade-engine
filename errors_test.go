package engine

import (
	"errors"
	"testing"
)

func TestDomainErrorContracts(t *testing.T) {
	t.Parallel()

	// === 定义测试用例 ===
	tests := []struct {
		name   string
		err    error
		target error
		want   string
	}{
		{
			name:   "validation error",
			err:    &ValidationError{Field: "listing.id", Reason: "must not be empty"},
			target: ErrInvalidArgument,
			want:   "validate listing.id: must not be empty",
		},
		{
			name:   "transition error",
			err:    &TransitionError{Entity: "listing", From: "open", To: "matched"},
			target: ErrInvalidTransition,
			want:   "transition listing from open to matched: invalid state transition",
		},
	}

	// === 执行测试用例 ===
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.err.Error(); got != tt.want {
				t.Fatalf("Error() = %q, want %q", got, tt.want)
			}
			if !errors.Is(tt.err, tt.target) {
				t.Fatalf("errors.Is(%v, %v) = false, want true", tt.err, tt.target)
			}
		})
	}
}

func TestStableErrorIdentities(t *testing.T) {
	t.Parallel()

	if ErrListingNotFound.Error() != "listing not found" {
		t.Fatalf("ErrListingNotFound = %q", ErrListingNotFound)
	}
	if ErrListingNotOwned.Error() != "listing not owned" {
		t.Fatalf("ErrListingNotOwned = %q", ErrListingNotOwned)
	}
	if ErrMatchRejected.Error() != "match rejected" {
		t.Fatalf("ErrMatchRejected = %q", ErrMatchRejected)
	}
}
