package engine

import (
	"testing"
	"time"
)

func TestRestoreMatchRequestRoundTrip(t *testing.T) {
	t.Parallel()

	// === 定义来源记录 ===
	tests := []MatchSource{MatchSourceAutomatic, MatchSourceSellerAccept, MatchSourceBuyerPurchase}

	// === 验证持久化往返 ===
	for _, source := range tests {
		t.Run(source.String(), func(t *testing.T) {
			t.Parallel()
			request := mustMatchRequest(t, source)
			if source == MatchSourceAutomatic {
				if err := request.Confirm(SideBuy, testTime.Add(time.Minute)); err != nil {
					t.Fatal(err)
				}
				if err := request.Confirm(SideSell, testTime.Add(2*time.Minute)); err != nil {
					t.Fatal(err)
				}
			}
			if err := request.Complete("order-1", request.UpdatedAt()); err != nil {
				t.Fatal(err)
			}
			record := request.Record()
			restored, err := RestoreMatchRequest(record)
			if err != nil {
				t.Fatalf("RestoreMatchRequest() error = %v", err)
			}
			if restored.Record() != record {
				t.Fatalf("Record() = %+v, want %+v", restored.Record(), record)
			}
		})
	}
}

func TestRestoreMatchRequestRejectsInvalidRecord(t *testing.T) {
	t.Parallel()

	// === 构造合法基线 ===
	base := mustMatchRequest(t, MatchSourceAutomatic).Record()
	tests := []struct {
		name      string
		change    func(*MatchRequestRecord)
		wantField string
	}{
		{name: "missing buyer", change: func(r *MatchRequestRecord) { r.BuyerID = "" }, wantField: "match_request.buyer_id"},
		{name: "same participant", change: func(r *MatchRequestRecord) { r.SellerID = r.BuyerID }, wantField: "match_request.participant_ids"},
		{name: "invalid price", change: func(r *MatchRequestRecord) { r.TradePriceCents = 0 }, wantField: "match_request.trade_price"},
		{name: "invalid source", change: func(r *MatchRequestRecord) { r.Source = 0 }, wantField: "match_request.source"},
		{name: "invalid status", change: func(r *MatchRequestRecord) { r.Status = 0 }, wantField: "match_request.status"},
		{name: "zero created at", change: func(r *MatchRequestRecord) { r.CreatedAt = time.Time{} }, wantField: "match_request.created_at"},
		{name: "updated before created", change: func(r *MatchRequestRecord) { r.UpdatedAt = r.CreatedAt.Add(-time.Nanosecond) }, wantField: "match_request.updated_at"},
		{name: "confirmation before created", change: func(r *MatchRequestRecord) { r.BuyerConfirmedAt = r.CreatedAt.Add(-time.Nanosecond) }, wantField: "match_request.buyer_confirmed_at"},
		{name: "confirmation after updated", change: func(r *MatchRequestRecord) { r.BuyerConfirmedAt = r.UpdatedAt.Add(time.Nanosecond) }, wantField: "match_request.buyer_confirmed_at"},
		{name: "seller confirmation before created", change: func(r *MatchRequestRecord) { r.SellerConfirmedAt = r.CreatedAt.Add(-time.Nanosecond) }, wantField: "match_request.seller_confirmed_at"},
		{name: "seller confirmation after updated", change: func(r *MatchRequestRecord) { r.SellerConfirmedAt = r.UpdatedAt.Add(time.Nanosecond) }, wantField: "match_request.seller_confirmed_at"},
		{name: "seller acceptance missing buy listing", change: func(r *MatchRequestRecord) {
			r.Source, r.BuyListingID, r.SellListingID = MatchSourceSellerAccept, "", ""
		}, wantField: "match_request.buy_listing_id"},
		{name: "buyer purchase missing sell listing", change: func(r *MatchRequestRecord) {
			r.Source, r.BuyListingID, r.SellListingID = MatchSourceBuyerPurchase, "", ""
		}, wantField: "match_request.sell_listing_id"},
		{name: "pending with order", change: func(r *MatchRequestRecord) { r.OrderRef = "order-1" }, wantField: "match_request.order_ref"},
		{name: "completed without confirmation", change: func(r *MatchRequestRecord) { r.Status = MatchRequestStatusCompleted; r.OrderRef = "order-1" }, wantField: "match_request.confirmations"},
		{name: "completed without order", change: func(r *MatchRequestRecord) {
			r.Status = MatchRequestStatusCompleted
			r.BuyerConfirmedAt, r.SellerConfirmedAt = r.CreatedAt, r.CreatedAt
		}, wantField: "match_request.order_ref"},
		{name: "direct trade without confirmations", change: func(r *MatchRequestRecord) {
			r.Source, r.SellListingID = MatchSourceSellerAccept, ""
		}, wantField: "match_request.confirmations"},
	}

	// === 执行损坏记录用例 ===
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			record := base
			tt.change(&record)
			request, err := RestoreMatchRequest(record)
			assertValidationError(t, err, tt.wantField)
			if request != nil {
				t.Fatalf("RestoreMatchRequest() = %+v, want nil", request)
			}
		})
	}
}
