package engine

import (
	"errors"
	"testing"
	"time"
)

func TestNewMatchRequest(t *testing.T) {
	t.Parallel()

	price := mustAmount(t, 2000)
	tests := []struct {
		name          string
		id            string
		buyListingID  string
		sellListingID string
		tradePrice    Amount
		source        MatchSource
		createdAt     time.Time
		wantField     string
		wantConfirmed bool
	}{
		{name: "automatic", id: "request-1", buyListingID: "buy-1", sellListingID: "sell-1", tradePrice: price, source: MatchSourceAutomatic, createdAt: testTime},
		{name: "seller accept", id: "request-1", buyListingID: "buy-1", sellListingID: "sell-1", tradePrice: price, source: MatchSourceSellerAccept, createdAt: testTime, wantConfirmed: true},
		{name: "buyer purchase", id: "request-1", buyListingID: "buy-1", sellListingID: "sell-1", tradePrice: price, source: MatchSourceBuyerPurchase, createdAt: testTime, wantConfirmed: true},
		{name: "missing id", buyListingID: "buy-1", sellListingID: "sell-1", tradePrice: price, source: MatchSourceAutomatic, createdAt: testTime, wantField: "match_request.id"},
		{name: "missing buy listing", id: "request-1", sellListingID: "sell-1", tradePrice: price, source: MatchSourceAutomatic, createdAt: testTime, wantField: "match_request.buy_listing_id"},
		{name: "missing sell listing", id: "request-1", buyListingID: "buy-1", tradePrice: price, source: MatchSourceAutomatic, createdAt: testTime, wantField: "match_request.sell_listing_id"},
		{name: "same listing", id: "request-1", buyListingID: "listing-1", sellListingID: "listing-1", tradePrice: price, source: MatchSourceAutomatic, createdAt: testTime, wantField: "match_request.listing_ids"},
		{name: "invalid price", id: "request-1", buyListingID: "buy-1", sellListingID: "sell-1", source: MatchSourceAutomatic, createdAt: testTime, wantField: "match_request.trade_price"},
		{name: "invalid source", id: "request-1", buyListingID: "buy-1", sellListingID: "sell-1", tradePrice: price, source: MatchSource(99), createdAt: testTime, wantField: "match_request.source"},
		{name: "zero creation time", id: "request-1", buyListingID: "buy-1", sellListingID: "sell-1", tradePrice: price, source: MatchSourceAutomatic, wantField: "match_request.created_at"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request, err := NewMatchRequest(tt.id, tt.buyListingID, tt.sellListingID, tt.tradePrice, tt.source, tt.createdAt)
			if tt.wantField != "" {
				assertValidationError(t, err, tt.wantField)
				if request != nil {
					t.Fatal("NewMatchRequest() returned a request after validation failure")
				}
				return
			}
			if err != nil {
				t.Fatalf("NewMatchRequest() error = %v", err)
			}
			if request.ID() != tt.id || request.BuyListingID() != tt.buyListingID || request.SellListingID() != tt.sellListingID {
				t.Fatalf("NewMatchRequest() identifiers = (%q, %q, %q)", request.ID(), request.BuyListingID(), request.SellListingID())
			}
			if request.TradePrice() != tt.tradePrice || request.Source() != tt.source {
				t.Fatalf("NewMatchRequest() trade values = (%v, %v)", request.TradePrice(), request.Source())
			}
			if request.BuyerConfirmed() != tt.wantConfirmed || request.SellerConfirmed() != tt.wantConfirmed {
				t.Fatalf("confirmations = (%t, %t), want both %t", request.BuyerConfirmed(), request.SellerConfirmed(), tt.wantConfirmed)
			}
			if request.Status() != MatchRequestStatusPending || request.OrderRef() != "" {
				t.Fatalf("state = (%v, %q), want pending without order", request.Status(), request.OrderRef())
			}
			if request.CreatedAt() != testTime || request.UpdatedAt() != testTime {
				t.Fatalf("timestamps = (%v, %v), want %v", request.CreatedAt(), request.UpdatedAt(), testTime)
			}
		})
	}
}

func TestMatchRequestConfirm(t *testing.T) {
	t.Parallel()

	request := mustMatchRequest(t, MatchSourceAutomatic)
	buyerAt := testTime.Add(time.Minute)
	if err := request.Confirm(SideBuy, buyerAt); err != nil {
		t.Fatalf("Confirm(buyer) error = %v", err)
	}
	if !request.BuyerConfirmed() || request.SellerConfirmed() {
		t.Fatalf("confirmations = (%t, %t), want (true, false)", request.BuyerConfirmed(), request.SellerConfirmed())
	}
	beforeDuplicate := *request
	if err := request.Confirm(SideBuy, testTime); err != nil {
		t.Fatalf("duplicate Confirm(buyer) error = %v", err)
	}
	if *request != beforeDuplicate {
		t.Fatalf("request changed after duplicate confirmation: got %+v, want %+v", *request, beforeDuplicate)
	}

	sellerAt := buyerAt.Add(time.Minute)
	if err := request.Confirm(SideSell, sellerAt); err != nil {
		t.Fatalf("Confirm(seller) error = %v", err)
	}
	if !request.BuyerConfirmed() || !request.SellerConfirmed() || request.UpdatedAt() != sellerAt {
		t.Fatalf("request not fully confirmed at %v: %+v", sellerAt, *request)
	}
}

func TestMatchRequestConfirmRejectsInvalidInputWithoutMutation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		side      Side
		at        time.Time
		wantField string
	}{
		{name: "invalid side", side: Side(99), at: testTime, wantField: "match_request.confirmation_side"},
		{name: "zero time", side: SideBuy, wantField: "match_request.updated_at"},
		{name: "time before update", side: SideBuy, at: testTime.Add(-time.Nanosecond), wantField: "match_request.updated_at"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := mustMatchRequest(t, MatchSourceAutomatic)
			before := *request
			err := request.Confirm(tt.side, tt.at)
			assertValidationError(t, err, tt.wantField)
			if *request != before {
				t.Fatalf("request changed after failed confirmation: got %+v, want %+v", *request, before)
			}
		})
	}
}

func TestMatchRequestComplete(t *testing.T) {
	t.Parallel()

	request := mustMatchRequest(t, MatchSourceAutomatic)
	before := *request
	if err := request.Complete("order-1", testTime.Add(time.Minute)); !errors.Is(err, ErrConfirmationRequired) {
		t.Fatalf("Complete() error = %v, want ErrConfirmationRequired", err)
	}
	if *request != before {
		t.Fatalf("request changed without confirmation: got %+v, want %+v", *request, before)
	}

	if err := request.Confirm(SideBuy, testTime.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := request.Confirm(SideSell, testTime.Add(2*time.Minute)); err != nil {
		t.Fatal(err)
	}
	beforeMissingOrder := *request
	if err := request.Complete("", testTime.Add(3*time.Minute)); err == nil {
		t.Fatal("Complete() error = nil, want validation error")
	} else {
		assertValidationError(t, err, "match_request.order_ref")
	}
	if *request != beforeMissingOrder {
		t.Fatalf("request changed without order reference: got %+v, want %+v", *request, beforeMissingOrder)
	}

	completedAt := testTime.Add(3 * time.Minute)
	if err := request.Complete("order-1", completedAt); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if request.Status() != MatchRequestStatusCompleted || request.OrderRef() != "order-1" || request.UpdatedAt() != completedAt {
		t.Fatalf("completed request = %+v", *request)
	}
}

func TestDirectMatchRequestCanCompleteImmediately(t *testing.T) {
	t.Parallel()

	tests := []MatchSource{MatchSourceSellerAccept, MatchSourceBuyerPurchase}
	for _, source := range tests {
		t.Run(source.String(), func(t *testing.T) {
			t.Parallel()

			request := mustMatchRequest(t, source)
			if err := request.Complete("order-1", testTime); err != nil {
				t.Fatalf("Complete() error = %v", err)
			}
			if request.Status() != MatchRequestStatusCompleted || request.OrderRef() != "order-1" {
				t.Fatalf("completed request = %+v", *request)
			}
		})
	}
}

func TestMatchRequestTerminalTransitions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		act        func(*MatchRequest, time.Time) error
		wantStatus MatchRequestStatus
	}{
		{name: "decline", act: (*MatchRequest).Decline, wantStatus: MatchRequestStatusDeclined},
		{name: "cancel", act: (*MatchRequest).Cancel, wantStatus: MatchRequestStatusCancelled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := mustMatchRequest(t, MatchSourceAutomatic)
			at := testTime.Add(time.Minute)
			if err := tt.act(request, at); err != nil {
				t.Fatalf("terminal transition error = %v", err)
			}
			if request.Status() != tt.wantStatus || request.UpdatedAt() != at {
				t.Fatalf("terminal request = %+v, want status %v at %v", *request, tt.wantStatus, at)
			}
			before := *request
			err := request.Cancel(at.Add(time.Minute))
			assertTransitionError(t, err, "match_request")
			if *request != before {
				t.Fatalf("terminal request changed: got %+v, want %+v", *request, before)
			}
		})
	}
}

func TestMatchRequestRejectsInvalidTerminalTimeWithoutMutation(t *testing.T) {
	t.Parallel()

	request := mustMatchRequest(t, MatchSourceAutomatic)
	before := *request
	err := request.Decline(testTime.Add(-time.Nanosecond))
	assertValidationError(t, err, "match_request.updated_at")
	if *request != before {
		t.Fatalf("request changed after invalid time: got %+v, want %+v", *request, before)
	}
}

func TestNilMatchRequestOperations(t *testing.T) {
	t.Parallel()

	var request *MatchRequest
	assertValidationError(t, request.Confirm(SideBuy, testTime), "match_request")
	assertValidationError(t, request.Complete("order-1", testTime), "match_request")
	assertValidationError(t, request.Decline(testTime), "match_request")
	assertValidationError(t, request.Cancel(testTime), "match_request")
}

func mustMatchRequest(t *testing.T, source MatchSource) *MatchRequest {
	t.Helper()
	request, err := NewMatchRequest("request-1", "buy-1", "sell-1", mustAmount(t, 2000), source, testTime)
	if err != nil {
		t.Fatalf("NewMatchRequest() error = %v", err)
	}
	return request
}
