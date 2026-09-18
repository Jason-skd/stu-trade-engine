package engine

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"
)

func TestDirectTradePathsCompleteTargetListing(t *testing.T) {
	t.Parallel()

	// === 定义直接成交路径 ===
	tests := []struct {
		name            string
		side            Side
		ownerID         string
		resourceRef     string
		run             func(*directTradeState) (DirectTradeResult, error)
		wantSource      MatchSource
		wantBuyerID     string
		wantSellerID    string
		wantResourceRef string
		wantBuyListing  string
		wantSellListing string
	}{
		{
			name: "buyer purchases sell listing", side: SideSell, ownerID: "seller-1", resourceRef: "book-1",
			run: func(state *directTradeState) (DirectTradeResult, error) {
				return runDirectTradeTransaction(state, func(tx *recordingDirectTradeTransaction) (DirectTradeResult, error) {
					return PurchaseSellListing(context.Background(), tx, PurchaseSellListingInput{
						MatchRequestID: "request-1", SellListingID: "listing-1", BuyerID: "buyer-1", CreatedAt: testTime.Add(time.Minute),
					})
				})
			},
			wantSource: MatchSourceBuyerPurchase, wantBuyerID: "buyer-1", wantSellerID: "seller-1",
			wantResourceRef: "book-1", wantSellListing: "listing-1",
		},
		{
			name: "seller accepts buy listing", side: SideBuy, ownerID: "buyer-1", resourceRef: "wanted-book-1",
			run: func(state *directTradeState) (DirectTradeResult, error) {
				return runDirectTradeTransaction(state, func(tx *recordingDirectTradeTransaction) (DirectTradeResult, error) {
					return AcceptBuyListing(context.Background(), tx, AcceptBuyListingInput{
						MatchRequestID: "request-1", BuyListingID: "listing-1", SellerID: "seller-1",
						ResourceRef: "delivered-book-1", CreatedAt: testTime.Add(time.Minute),
					})
				})
			},
			wantSource: MatchSourceSellerAccept, wantBuyerID: "buyer-1", wantSellerID: "seller-1",
			wantResourceRef: "delivered-book-1", wantBuyListing: "listing-1",
		},
	}

	// === 验证共享成交结果 ===
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			state := newDirectTradeState(tt.side, tt.ownerID, tt.resourceRef)
			result, err := tt.run(state)
			if err != nil {
				t.Fatalf("direct trade error = %v", err)
			}
			if result.Listing.Status() != ListingStatusMatched || state.listing.Status != ListingStatusMatched {
				t.Fatalf("listing statuses = (%v, %v), want matched", result.Listing.Status(), state.listing.Status)
			}
			request := result.MatchRequest
			if request.Status() != MatchRequestStatusCompleted || request.OrderRef() != "order-request-1" {
				t.Fatalf("request completion = (%v, %q)", request.Status(), request.OrderRef())
			}
			if request.Source() != tt.wantSource || request.BuyerID() != tt.wantBuyerID || request.SellerID() != tt.wantSellerID {
				t.Fatalf("request trade facts = (%v, %q, %q)", request.Source(), request.BuyerID(), request.SellerID())
			}
			if request.ResourceRef() != tt.wantResourceRef || request.BuyListingID() != tt.wantBuyListing || request.SellListingID() != tt.wantSellListing {
				t.Fatalf("request associations = (%q, %q, %q)", request.ResourceRef(), request.BuyListingID(), request.SellListingID())
			}
			if request.TradePrice().Cents() != 2000 || request.BuyerConfirmedAt() != testTime.Add(time.Minute) || request.SellerConfirmedAt() != testTime.Add(time.Minute) {
				t.Fatalf("price and confirmations = (%d, %v, %v)", request.TradePrice().Cents(), request.BuyerConfirmedAt(), request.SellerConfirmedAt())
			}
			wantCalls := []string{"read_listing", "evaluate_match", "create_match_request", "create_trade", "update_match_request", "update_listing"}
			if !slices.Equal(state.calls, wantCalls) {
				t.Fatalf("calls = %v, want %v", state.calls, wantCalls)
			}
			if len(state.requests) != 1 || len(state.orders) != 1 || state.proposal != (MatchProposal{
				BuyerID: tt.wantBuyerID, SellerID: tt.wantSellerID, ResourceRef: tt.wantResourceRef,
				BuyListingID: tt.wantBuyListing, SellListingID: tt.wantSellListing,
				TradePrice: mustAmount(t, 2000), Source: tt.wantSource,
			}) {
				t.Fatalf("persisted trade = requests:%d orders:%d proposal:%+v", len(state.requests), len(state.orders), state.proposal)
			}
		})
	}
}

func TestDirectTradeRejectsUnavailableListingBeforeSideEffects(t *testing.T) {
	t.Parallel()

	// === 定义不可成交目标 ===
	tests := []struct {
		name   string
		side   Side
		status ListingStatus
	}{
		{name: "wrong side", side: SideBuy, status: ListingStatusOpen},
		{name: "reserved", side: SideSell, status: ListingStatusReserved},
		{name: "matched", side: SideSell, status: ListingStatusMatched},
		{name: "withdrawn", side: SideSell, status: ListingStatusWithdrawn},
	}

	// === 验证拒绝无副作用 ===
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			state := newDirectTradeState(tt.side, "seller-1", "book-1")
			state.listing.Status = tt.status
			_, err := runDirectTradeTransaction(state, func(tx *recordingDirectTradeTransaction) (DirectTradeResult, error) {
				return PurchaseSellListing(context.Background(), tx, validPurchaseInput())
			})
			if !errors.Is(err, ErrListingUnavailable) {
				t.Fatalf("PurchaseSellListing() error = %v, want ErrListingUnavailable", err)
			}
			if len(state.requests) != 0 || len(state.orders) != 0 || !slices.Equal(state.calls, []string{"read_listing"}) {
				t.Fatalf("state changed after rejection: requests:%d orders:%d calls:%v", len(state.requests), len(state.orders), state.calls)
			}
		})
	}
}

func TestDirectTradePreservesPolicyRejection(t *testing.T) {
	t.Parallel()

	policyRejection := errors.New("self trade rejected")
	state := newDirectTradeState(SideSell, "buyer-1", "book-1")
	state.failStage, state.failure = "evaluate_match", policyRejection
	_, err := runDirectTradeTransaction(state, func(tx *recordingDirectTradeTransaction) (DirectTradeResult, error) {
		return PurchaseSellListing(context.Background(), tx, validPurchaseInput())
	})
	if !errors.Is(err, policyRejection) {
		t.Fatalf("PurchaseSellListing() error = %v, want policy cause", err)
	}
	if state.listing.Status != ListingStatusOpen || len(state.requests) != 0 || len(state.orders) != 0 {
		t.Fatalf("state changed after policy rejection: %+v", state)
	}
}

func TestDirectTradeRejectsStaleTimeBeforeSideEffects(t *testing.T) {
	t.Parallel()

	state := newDirectTradeState(SideSell, "seller-1", "book-1")
	input := validPurchaseInput()
	input.CreatedAt = testTime.Add(-time.Nanosecond)
	_, err := runDirectTradeTransaction(state, func(tx *recordingDirectTradeTransaction) (DirectTradeResult, error) {
		return PurchaseSellListing(context.Background(), tx, input)
	})
	assertValidationError(t, err, "listing.updated_at")
	if !slices.Equal(state.calls, []string{"read_listing"}) || len(state.requests) != 0 || len(state.orders) != 0 {
		t.Fatalf("stale trade changed state: calls:%v requests:%d orders:%d", state.calls, len(state.requests), len(state.orders))
	}
}

func TestDirectTradeDependencyFailuresRollBack(t *testing.T) {
	t.Parallel()

	// === 定义失败注入点 ===
	stages := []string{"read_listing", "create_match_request", "create_trade", "update_match_request", "update_listing"}
	for _, stage := range stages {
		t.Run(stage, func(t *testing.T) {
			t.Parallel()
			cause := errors.New(stage + " failed")
			state := newDirectTradeState(SideSell, "seller-1", "book-1")
			state.failStage, state.failure = stage, cause
			_, err := runDirectTradeTransaction(state, func(tx *recordingDirectTradeTransaction) (DirectTradeResult, error) {
				return PurchaseSellListing(context.Background(), tx, validPurchaseInput())
			})
			if !errors.Is(err, cause) {
				t.Fatalf("PurchaseSellListing() error = %v, want %v", err, cause)
			}
			if state.listing.Status != ListingStatusOpen || len(state.requests) != 0 || len(state.orders) != 0 {
				t.Fatalf("state persisted after %s failure: listing:%v requests:%d orders:%d", stage, state.listing.Status, len(state.requests), len(state.orders))
			}
		})
	}
}

func TestDirectTradeRollsBackEmptyOrderReference(t *testing.T) {
	t.Parallel()

	state := newDirectTradeState(SideSell, "seller-1", "book-1")
	state.orderRef = ""
	_, err := runDirectTradeTransaction(state, func(tx *recordingDirectTradeTransaction) (DirectTradeResult, error) {
		return PurchaseSellListing(context.Background(), tx, validPurchaseInput())
	})
	assertValidationError(t, err, "match_request.order_ref")
	if state.listing.Status != ListingStatusOpen || len(state.requests) != 0 || len(state.orders) != 0 {
		t.Fatalf("state persisted without order reference: %+v", state)
	}
}

func TestDirectTradeRepeatedSubmissionCreatesOneTrade(t *testing.T) {
	t.Parallel()

	state := newDirectTradeState(SideSell, "seller-1", "book-1")
	run := func() (DirectTradeResult, error) {
		return runDirectTradeTransaction(state, func(tx *recordingDirectTradeTransaction) (DirectTradeResult, error) {
			return PurchaseSellListing(context.Background(), tx, validPurchaseInput())
		})
	}
	if _, err := run(); err != nil {
		t.Fatal(err)
	}
	if _, err := run(); !errors.Is(err, ErrListingUnavailable) {
		t.Fatalf("repeated trade error = %v, want ErrListingUnavailable", err)
	}
	if len(state.requests) != 1 || len(state.orders) != 1 {
		t.Fatalf("repeated trade persisted requests:%d orders:%d", len(state.requests), len(state.orders))
	}
}

func TestConcurrentDirectTradesCreateOneTrade(t *testing.T) {
	t.Parallel()

	// === 竞争同一目标挂单 ===
	state := newDirectTradeState(SideSell, "seller-1", "book-1")
	var transaction sync.Mutex
	errorsByBuyer := make(chan error, 2)
	for _, buyerID := range []string{"buyer-1", "buyer-2"} {
		go func() {
			transaction.Lock()
			defer transaction.Unlock()
			input := validPurchaseInput()
			input.MatchRequestID = "request-" + buyerID
			input.BuyerID = buyerID
			_, err := runDirectTradeTransaction(state, func(tx *recordingDirectTradeTransaction) (DirectTradeResult, error) {
				return PurchaseSellListing(context.Background(), tx, input)
			})
			errorsByBuyer <- err
		}()
	}

	// === 验证唯一成交 ===
	successes, unavailable := 0, 0
	for range 2 {
		err := <-errorsByBuyer
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ErrListingUnavailable):
			unavailable++
		default:
			t.Fatalf("concurrent trade error = %v", err)
		}
	}
	if successes != 1 || unavailable != 1 || len(state.requests) != 1 || len(state.orders) != 1 {
		t.Fatalf("concurrent results = success:%d unavailable:%d requests:%d orders:%d", successes, unavailable, len(state.requests), len(state.orders))
	}
}

func TestDirectTradeValidatesInputBeforeDependencies(t *testing.T) {
	t.Parallel()

	// === 定义无效调用 ===
	validTx := &recordingDirectTradeTransaction{state: newDirectTradeState(SideSell, "seller-1", "book-1")}
	tests := []struct {
		name      string
		tx        DirectTradeTransaction
		input     PurchaseSellListingInput
		wantField string
	}{
		{name: "nil transaction", input: validPurchaseInput(), wantField: "direct_trade_transaction"},
		{name: "missing request", tx: validTx, input: replacePurchaseInput(validPurchaseInput(), func(i *PurchaseSellListingInput) { i.MatchRequestID = "" }), wantField: "match_request.id"},
		{name: "missing listing", tx: validTx, input: replacePurchaseInput(validPurchaseInput(), func(i *PurchaseSellListingInput) { i.SellListingID = "" }), wantField: "listing.id"},
		{name: "missing buyer", tx: validTx, input: replacePurchaseInput(validPurchaseInput(), func(i *PurchaseSellListingInput) { i.BuyerID = "" }), wantField: "direct_trade.initiator_id"},
		{name: "zero time", tx: validTx, input: replacePurchaseInput(validPurchaseInput(), func(i *PurchaseSellListingInput) { i.CreatedAt = time.Time{} }), wantField: "match_request.created_at"},
	}

	// === 验证调用边界 ===
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := PurchaseSellListing(context.Background(), tt.tx, tt.input)
			assertValidationError(t, err, tt.wantField)
		})
	}
	if len(validTx.state.calls) != 0 {
		t.Fatalf("dependencies called during input validation: %v", validTx.state.calls)
	}
}

func TestAcceptBuyListingRequiresDeliveryResource(t *testing.T) {
	t.Parallel()

	tx := &recordingDirectTradeTransaction{state: newDirectTradeState(SideBuy, "buyer-1", "wanted-book-1")}
	_, err := AcceptBuyListing(context.Background(), tx, AcceptBuyListingInput{
		MatchRequestID: "request-1", BuyListingID: "listing-1",
		SellerID: "seller-1", CreatedAt: testTime.Add(time.Minute),
	})
	assertValidationError(t, err, "direct_trade.resource_ref")
	if len(tx.state.calls) != 0 {
		t.Fatalf("dependencies called during input validation: %v", tx.state.calls)
	}
}

type directTradeState struct {
	listing   ListingRecord
	requests  map[string]MatchRequestRecord
	orders    map[string]string
	proposal  MatchProposal
	calls     []string
	failStage string
	failure   error
	orderRef  string
}

func newDirectTradeState(side Side, ownerID, resourceRef string) *directTradeState {
	return &directTradeState{
		listing: ListingRecord{
			ID: "listing-1", OwnerID: ownerID, ResourceRef: resourceRef,
			Side: side, PriceCents: 2000, Status: ListingStatusOpen,
			CreatedAt: testTime, UpdatedAt: testTime,
		},
		requests: make(map[string]MatchRequestRecord), orders: make(map[string]string),
		orderRef: "order-request-1",
	}
}

func (s *directTradeState) clone() *directTradeState {
	clone := *s
	clone.requests = make(map[string]MatchRequestRecord, len(s.requests))
	for id, record := range s.requests {
		clone.requests[id] = record
	}
	clone.orders = make(map[string]string, len(s.orders))
	for id, order := range s.orders {
		clone.orders[id] = order
	}
	clone.calls = nil
	return &clone
}

type recordingDirectTradeTransaction struct{ state *directTradeState }

func (tx *recordingDirectTradeTransaction) ReadListingForUpdate(context.Context, string) (ListingRecord, error) {
	tx.state.calls = append(tx.state.calls, "read_listing")
	if err := tx.state.fail("read_listing"); err != nil {
		return ListingRecord{}, err
	}
	return tx.state.listing, nil
}

func (tx *recordingDirectTradeTransaction) EvaluateMatch(_ context.Context, proposal MatchProposal) error {
	tx.state.calls = append(tx.state.calls, "evaluate_match")
	tx.state.proposal = proposal
	return tx.state.fail("evaluate_match")
}

func (tx *recordingDirectTradeTransaction) CreateMatchRequest(_ context.Context, record MatchRequestRecord) error {
	tx.state.calls = append(tx.state.calls, "create_match_request")
	if err := tx.state.fail("create_match_request"); err != nil {
		return err
	}
	tx.state.requests[record.ID] = record
	return nil
}

func (tx *recordingDirectTradeTransaction) CreateTrade(_ context.Context, input TradeInput) (string, error) {
	tx.state.calls = append(tx.state.calls, "create_trade")
	if err := tx.state.fail("create_trade"); err != nil {
		return "", err
	}
	if orderRef, exists := tx.state.orders[input.MatchRequestID]; exists {
		return orderRef, nil
	}
	tx.state.orders[input.MatchRequestID] = tx.state.orderRef
	return tx.state.orderRef, nil
}

func (tx *recordingDirectTradeTransaction) UpdateMatchRequest(_ context.Context, record MatchRequestRecord) error {
	tx.state.calls = append(tx.state.calls, "update_match_request")
	if err := tx.state.fail("update_match_request"); err != nil {
		return err
	}
	tx.state.requests[record.ID] = record
	return nil
}

func (tx *recordingDirectTradeTransaction) UpdateListing(_ context.Context, record ListingRecord) error {
	tx.state.calls = append(tx.state.calls, "update_listing")
	if err := tx.state.fail("update_listing"); err != nil {
		return err
	}
	tx.state.listing = record
	return nil
}

func (s *directTradeState) fail(stage string) error {
	if s.failStage == stage {
		return s.failure
	}
	return nil
}

func runDirectTradeTransaction(
	state *directTradeState,
	run func(*recordingDirectTradeTransaction) (DirectTradeResult, error),
) (DirectTradeResult, error) {
	staged := state.clone()
	result, err := run(&recordingDirectTradeTransaction{state: staged})
	state.calls = append(state.calls, staged.calls...)
	state.proposal = staged.proposal
	if err != nil {
		return DirectTradeResult{}, err
	}
	state.listing, state.requests, state.orders = staged.listing, staged.requests, staged.orders
	return result, nil
}

func validPurchaseInput() PurchaseSellListingInput {
	return PurchaseSellListingInput{
		MatchRequestID: "request-1", SellListingID: "listing-1",
		BuyerID: "buyer-1", CreatedAt: testTime.Add(time.Minute),
	}
}

func replacePurchaseInput(input PurchaseSellListingInput, replace func(*PurchaseSellListingInput)) PurchaseSellListingInput {
	replace(&input)
	return input
}
