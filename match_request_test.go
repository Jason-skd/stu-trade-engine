package engine

import (
	"errors"
	"testing"
	"time"
)

func TestNewMatchRequest(t *testing.T) {
	t.Parallel()

	// === 定义测试用例 ===
	price := mustAmount(t, 2000)
	valid := MatchRequestInput{
		ID: "request-1", BuyerID: "buyer-1", SellerID: "seller-1", ResourceRef: "resource-1",
		BuyListingID: "buy-1", SellListingID: "sell-1", TradePrice: price,
		Source: MatchSourceAutomatic, CreatedAt: testTime,
	}
	tests := []struct {
		name          string
		input         MatchRequestInput
		wantField     string
		wantConfirmed bool
	}{
		{name: "automatic", input: valid},
		{name: "seller accept", input: MatchRequestInput{ID: "request-1", BuyerID: "buyer-1", SellerID: "seller-1", ResourceRef: "resource-1", BuyListingID: "buy-1", TradePrice: price, Source: MatchSourceSellerAccept, CreatedAt: testTime}, wantConfirmed: true},
		{name: "buyer purchase", input: MatchRequestInput{ID: "request-1", BuyerID: "buyer-1", SellerID: "seller-1", ResourceRef: "resource-1", SellListingID: "sell-1", TradePrice: price, Source: MatchSourceBuyerPurchase, CreatedAt: testTime}, wantConfirmed: true},
		{name: "missing id", input: replaceMatchInput(valid, func(i *MatchRequestInput) { i.ID = "" }), wantField: "match_request.id"},
		{name: "missing buyer", input: replaceMatchInput(valid, func(i *MatchRequestInput) { i.BuyerID = "" }), wantField: "match_request.buyer_id"},
		{name: "same participant", input: replaceMatchInput(valid, func(i *MatchRequestInput) { i.SellerID = i.BuyerID }), wantField: "match_request.participant_ids"},
		{name: "missing resource", input: replaceMatchInput(valid, func(i *MatchRequestInput) { i.ResourceRef = "" }), wantField: "match_request.resource_ref"},
		{name: "missing buy listing", input: replaceMatchInput(valid, func(i *MatchRequestInput) { i.BuyListingID = "" }), wantField: "match_request.buy_listing_id"},
		{name: "missing sell listing", input: replaceMatchInput(valid, func(i *MatchRequestInput) { i.SellListingID = "" }), wantField: "match_request.sell_listing_id"},
		{name: "same listing", input: replaceMatchInput(valid, func(i *MatchRequestInput) { i.SellListingID = i.BuyListingID }), wantField: "match_request.listing_ids"},
		{name: "seller acceptance with sell listing", input: MatchRequestInput{ID: "request-1", BuyerID: "buyer-1", SellerID: "seller-1", ResourceRef: "resource-1", BuyListingID: "buy-1", SellListingID: "sell-1", TradePrice: price, Source: MatchSourceSellerAccept, CreatedAt: testTime}, wantField: "match_request.sell_listing_id"},
		{name: "buyer purchase with buy listing", input: MatchRequestInput{ID: "request-1", BuyerID: "buyer-1", SellerID: "seller-1", ResourceRef: "resource-1", BuyListingID: "buy-1", SellListingID: "sell-1", TradePrice: price, Source: MatchSourceBuyerPurchase, CreatedAt: testTime}, wantField: "match_request.buy_listing_id"},
		{name: "invalid price", input: replaceMatchInput(valid, func(i *MatchRequestInput) { i.TradePrice = Amount{} }), wantField: "match_request.trade_price"},
		{name: "invalid source", input: replaceMatchInput(valid, func(i *MatchRequestInput) { i.Source = MatchSource(99) }), wantField: "match_request.source"},
		{name: "zero creation time", input: replaceMatchInput(valid, func(i *MatchRequestInput) { i.CreatedAt = time.Time{} }), wantField: "match_request.created_at"},
	}

	// === 执行测试用例 ===
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request, err := NewMatchRequest(tt.input)
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
			if request.ID() != tt.input.ID || request.BuyListingID() != tt.input.BuyListingID || request.SellListingID() != tt.input.SellListingID {
				t.Fatalf("NewMatchRequest() identifiers = (%q, %q, %q)", request.ID(), request.BuyListingID(), request.SellListingID())
			}
			if request.BuyerID() != tt.input.BuyerID || request.SellerID() != tt.input.SellerID || request.ResourceRef() != tt.input.ResourceRef {
				t.Fatalf("participants and resource = (%q, %q, %q)", request.BuyerID(), request.SellerID(), request.ResourceRef())
			}
			if request.TradePrice() != tt.input.TradePrice || request.Source() != tt.input.Source {
				t.Fatalf("NewMatchRequest() trade values = (%v, %v)", request.TradePrice(), request.Source())
			}
			if request.BuyerConfirmed() != tt.wantConfirmed || request.SellerConfirmed() != tt.wantConfirmed {
				t.Fatalf("confirmations = (%t, %t), want both %t", request.BuyerConfirmed(), request.SellerConfirmed(), tt.wantConfirmed)
			}
			if request.Status() != MatchRequestStatusPending || request.OrderRef() != "" {
				t.Fatalf("state = (%v, %q), want pending without order", request.Status(), request.OrderRef())
			}
			if tt.wantConfirmed && (request.BuyerConfirmedAt() != testTime || request.SellerConfirmedAt() != testTime) {
				t.Fatalf("confirmation times = (%v, %v), want %v", request.BuyerConfirmedAt(), request.SellerConfirmedAt(), testTime)
			}
			if request.CreatedAt() != testTime || request.UpdatedAt() != testTime {
				t.Fatalf("timestamps = (%v, %v), want %v", request.CreatedAt(), request.UpdatedAt(), testTime)
			}
		})
	}
}

func TestMatchRequestConfirm(t *testing.T) {
	t.Parallel()

	// === 记录买方确认 ===
	request := mustMatchRequest(t, MatchSourceAutomatic)
	buyerAt := testTime.Add(time.Minute)
	if err := request.Confirm(SideBuy, buyerAt); err != nil {
		t.Fatalf("Confirm(buyer) error = %v", err)
	}
	if !request.BuyerConfirmed() || request.SellerConfirmed() {
		t.Fatalf("confirmations = (%t, %t), want (true, false)", request.BuyerConfirmed(), request.SellerConfirmed())
	}

	// === 验证重复确认 ===
	beforeDuplicate := *request
	if err := request.Confirm(SideBuy, testTime); err != nil {
		t.Fatalf("duplicate Confirm(buyer) error = %v", err)
	}
	if *request != beforeDuplicate {
		t.Fatalf("request changed after duplicate confirmation: got %+v, want %+v", *request, beforeDuplicate)
	}

	// === 记录卖方确认 ===
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

	// === 定义测试用例 ===
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

	// === 执行测试用例 ===
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

	// === 拒绝未确认请求 ===
	request := mustMatchRequest(t, MatchSourceAutomatic)
	before := *request
	if err := request.Complete("order-1", testTime.Add(time.Minute)); !errors.Is(err, ErrConfirmationRequired) {
		t.Fatalf("Complete() error = %v, want ErrConfirmationRequired", err)
	}
	if *request != before {
		t.Fatalf("request changed without confirmation: got %+v, want %+v", *request, before)
	}

	// === 确认双方参与者 ===
	if err := request.Confirm(SideBuy, testTime.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := request.Confirm(SideSell, testTime.Add(2*time.Minute)); err != nil {
		t.Fatal(err)
	}

	// === 拒绝空订单引用 ===
	beforeMissingOrder := *request
	if err := request.Complete("", testTime.Add(3*time.Minute)); err == nil {
		t.Fatal("Complete() error = nil, want validation error")
	} else {
		assertValidationError(t, err, "match_request.order_ref")
	}
	if *request != beforeMissingOrder {
		t.Fatalf("request changed without order reference: got %+v, want %+v", *request, beforeMissingOrder)
	}

	// === 完成撮合请求 ===
	completedAt := testTime.Add(3 * time.Minute)
	if err := request.Complete("order-1", completedAt); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if request.Status() != MatchRequestStatusCompleted || request.OrderRef() != "order-1" || request.UpdatedAt() != completedAt {
		t.Fatalf("completed request = %+v", *request)
	}
}

func TestMatchRequestCompleteRejectsInvalidTimeWithoutMutation(t *testing.T) {
	t.Parallel()

	request := mustMatchRequest(t, MatchSourceAutomatic)
	if err := request.Confirm(SideBuy, testTime.Add(time.Minute)); err != nil {
		t.Fatalf("Confirm(buyer) error = %v", err)
	}
	if err := request.Confirm(SideSell, testTime.Add(2*time.Minute)); err != nil {
		t.Fatalf("Confirm(seller) error = %v", err)
	}

	before := *request
	err := request.Complete("order-1", testTime.Add(time.Minute))
	assertValidationError(t, err, "match_request.updated_at")
	if *request != before {
		t.Fatalf("request changed after invalid completion time: got %+v, want %+v", *request, before)
	}
}

func TestDirectMatchRequestCanCompleteImmediately(t *testing.T) {
	t.Parallel()

	// === 定义测试用例 ===
	tests := []MatchSource{MatchSourceSellerAccept, MatchSourceBuyerPurchase}

	// === 执行测试用例 ===
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

	// === 定义测试用例 ===
	tests := []struct {
		name       string
		act        func(*MatchRequest, time.Time) error
		wantStatus MatchRequestStatus
	}{
		{name: "decline", act: (*MatchRequest).Decline, wantStatus: MatchRequestStatusDeclined},
		{name: "cancel", act: (*MatchRequest).Cancel, wantStatus: MatchRequestStatusCancelled},
	}

	// === 执行测试用例 ===
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
	input := MatchRequestInput{
		ID: "request-1", BuyerID: "buyer-1", SellerID: "seller-1", ResourceRef: "resource-1",
		BuyListingID: "buy-1", SellListingID: "sell-1", TradePrice: mustAmount(t, 2000),
		Source: source, CreatedAt: testTime,
	}
	if source == MatchSourceSellerAccept {
		input.SellListingID = ""
	}
	if source == MatchSourceBuyerPurchase {
		input.BuyListingID = ""
	}
	request, err := NewMatchRequest(input)
	if err != nil {
		t.Fatalf("NewMatchRequest() error = %v", err)
	}
	return request
}

func replaceMatchInput(input MatchRequestInput, replace func(*MatchRequestInput)) MatchRequestInput {
	replace(&input)
	return input
}
