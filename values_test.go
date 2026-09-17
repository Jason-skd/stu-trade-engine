package engine

import (
	"fmt"
	"testing"
)

func TestStableValueStrings(t *testing.T) {
	t.Parallel()

	// === 定义测试用例 ===
	tests := []struct {
		name  string
		value fmt.Stringer
		want  string
	}{
		{name: "buy side", value: SideBuy, want: "buy"},
		{name: "sell side", value: SideSell, want: "sell"},
		{name: "unknown side", value: Side(0), want: "unknown"},
		{name: "open listing", value: ListingStatusOpen, want: "open"},
		{name: "reserved listing", value: ListingStatusReserved, want: "reserved"},
		{name: "matched listing", value: ListingStatusMatched, want: "matched"},
		{name: "withdrawn listing", value: ListingStatusWithdrawn, want: "withdrawn"},
		{name: "unknown listing status", value: ListingStatus(0), want: "unknown"},
		{name: "automatic match", value: MatchSourceAutomatic, want: "automatic"},
		{name: "seller accept", value: MatchSourceSellerAccept, want: "seller_accept"},
		{name: "buyer purchase", value: MatchSourceBuyerPurchase, want: "buyer_purchase"},
		{name: "unknown match source", value: MatchSource(0), want: "unknown"},
		{name: "pending request", value: MatchRequestStatusPending, want: "pending"},
		{name: "completed request", value: MatchRequestStatusCompleted, want: "completed"},
		{name: "declined request", value: MatchRequestStatusDeclined, want: "declined"},
		{name: "cancelled request", value: MatchRequestStatusCancelled, want: "cancelled"},
		{name: "unknown request status", value: MatchRequestStatus(0), want: "unknown"},
	}

	// === 执行测试用例 ===
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.value.String(); got != tt.want {
				t.Fatalf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
