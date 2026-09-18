package engine

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"
)

func TestDirectTradePathsCommitExpectedProposalAndState(t *testing.T) {
	t.Parallel()

	// === 定义直接成交路径 ===
	tests := []struct {
		name            string
		side            Side
		ownerID         string
		listingResource string
		run             func(*recordingDirectTradeUnitOfWork) (DirectTradeResult, error)
		wantProposal    MatchProposal
		wantResource    string
		wantBuyListing  string
		wantSellListing string
	}{
		{
			name: "buyer purchases sell listing", side: SideSell,
			ownerID: "seller-1", listingResource: "book-1",
			run: func(uow *recordingDirectTradeUnitOfWork) (DirectTradeResult, error) {
				return PurchaseSellListing(context.Background(), uow, validPurchaseInput())
			},
			wantProposal: MatchProposal{
				Buyer: MatchParticipant{ParticipantID: "buyer-1"},
				Seller: MatchParticipant{
					ParticipantID: "seller-1", ListingID: "listing-1", ResourceRef: "book-1",
				},
				TradePrice: mustAmount(t, 2000), Source: MatchSourceBuyerPurchase,
			},
			wantResource: "book-1", wantSellListing: "listing-1",
		},
		{
			name: "seller accepts buy listing", side: SideBuy,
			ownerID: "buyer-1", listingResource: "wanted-book-1",
			run: func(uow *recordingDirectTradeUnitOfWork) (DirectTradeResult, error) {
				return AcceptBuyListing(context.Background(), uow, AcceptBuyListingInput{
					MatchRequestID: "request-1", BuyListingID: "listing-1", SellerID: "seller-1",
					ResourceRef: "delivered-book-1", CreatedAt: testTime.Add(time.Minute),
				})
			},
			wantProposal: MatchProposal{
				Buyer: MatchParticipant{
					ParticipantID: "buyer-1", ListingID: "listing-1", ResourceRef: "wanted-book-1",
				},
				Seller:     MatchParticipant{ParticipantID: "seller-1", ResourceRef: "delivered-book-1"},
				TradePrice: mustAmount(t, 2000), Source: MatchSourceSellerAccept,
			},
			wantResource: "delivered-book-1", wantBuyListing: "listing-1",
		},
	}

	// === 验证事务与成交事实 ===
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uow := newRecordingDirectTradeUnitOfWork(tt.side, tt.ownerID, tt.listingResource)
			result, err := tt.run(uow)
			if err != nil {
				t.Fatalf("direct trade error = %v", err)
			}
			if uow.withTxCalls != 1 || !uow.readInsideTx || !uow.committed {
				t.Fatalf("transaction evidence = calls:%d read-inside:%t committed:%t", uow.withTxCalls, uow.readInsideTx, uow.committed)
			}
			if !slices.Equal(uow.lockIDs, []string{"listing-1"}) {
				t.Fatalf("locked ids = %v", uow.lockIDs)
			}
			wantCalls := []string{
				"with_tx", "read_listings_for_update", "validate_match", "create_match_request",
				"create_trade", "update_match_request", "update_listing", "commit",
			}
			if !slices.Equal(uow.state.calls, wantCalls) {
				t.Fatalf("calls = %v, want %v", uow.state.calls, wantCalls)
			}
			if uow.state.proposal != tt.wantProposal {
				t.Fatalf("proposal = %+v, want %+v", uow.state.proposal, tt.wantProposal)
			}
			assertCompletedDirectTrade(t, result, uow.state, tt.wantProposal, tt.wantResource, tt.wantBuyListing, tt.wantSellListing)
		})
	}
}

func TestDirectTradeInterpretsAllValidatorOutcomes(t *testing.T) {
	t.Parallel()

	// === 定义校验返回组合 ===
	validatorErr := errors.New("validator unavailable")
	tests := []struct {
		name       string
		matched    bool
		err        error
		wantErr    error
		wantCommit bool
	}{
		{name: "allowed", matched: true, wantCommit: true},
		{name: "business mismatch", wantErr: ErrMatchRejected},
		{name: "execution failure", err: validatorErr, wantErr: validatorErr},
		{name: "true with execution failure", matched: true, err: validatorErr, wantErr: validatorErr},
	}

	// === 验证校验语义 ===
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uow := newRecordingDirectTradeUnitOfWork(SideSell, "seller-1", "book-1")
			uow.state.matchResult, uow.state.validationErr = tt.matched, tt.err
			result, err := PurchaseSellListing(context.Background(), uow, validPurchaseInput())
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("PurchaseSellListing() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantCommit {
				if result.Listing == nil || !uow.committed {
					t.Fatalf("allowed result = %+v, committed = %t", result, uow.committed)
				}
				return
			}
			if result != (DirectTradeResult{}) || uow.committed {
				t.Fatalf("rejected result = %+v, committed = %t", result, uow.committed)
			}
			assertNoDirectTradeWrites(t, uow.state)
		})
	}
}

func TestDirectTradeRejectsUnavailableListingBeforeWrites(t *testing.T) {
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

	// === 验证拒绝无写入 ===
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uow := newRecordingDirectTradeUnitOfWork(tt.side, "seller-1", "book-1")
			uow.state.listing.Status = tt.status
			_, err := PurchaseSellListing(context.Background(), uow, validPurchaseInput())
			if !errors.Is(err, ErrListingUnavailable) {
				t.Fatalf("PurchaseSellListing() error = %v, want ErrListingUnavailable", err)
			}
			if uow.state.listing.Status != tt.status || len(uow.state.requests) != 0 || len(uow.state.orders) != 0 {
				t.Fatalf("rejected state = listing:%v requests:%d orders:%d", uow.state.listing.Status, len(uow.state.requests), len(uow.state.orders))
			}
		})
	}
}

func TestDirectTradeRejectsInvalidLockedRecords(t *testing.T) {
	t.Parallel()

	// === 定义非法锁读取结果 ===
	tests := []struct {
		name      string
		records   []ListingRecord
		wantErr   error
		wantField string
	}{
		{name: "missing record", wantErr: ErrListingNotFound},
		{name: "unexpected record", records: []ListingRecord{listingRecord("listing-2")}, wantField: "listing.records"},
		{name: "duplicate record", records: []ListingRecord{listingRecord("listing-1"), listingRecord("listing-1")}, wantField: "listing.records"},
		{name: "extra record", records: []ListingRecord{listingRecord("listing-1"), listingRecord("listing-2")}, wantField: "listing.records"},
		{name: "corrupt record", records: []ListingRecord{{ID: "listing-1"}}, wantField: "listing.owner_id"},
	}

	// === 验证锁读取契约 ===
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uow := newRecordingDirectTradeUnitOfWork(SideSell, "seller-1", "book-1")
			uow.state.useLockedRecords = true
			uow.state.lockedRecords = tt.records
			_, err := PurchaseSellListing(context.Background(), uow, validPurchaseInput())
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("PurchaseSellListing() error = %v, want %v", err, tt.wantErr)
				}
			} else {
				assertValidationError(t, err, tt.wantField)
			}
			assertNoDirectTradeWrites(t, uow.state)
		})
	}
}

func TestDirectTradeRejectsStaleTimeBeforeWrites(t *testing.T) {
	t.Parallel()

	// === 提交过期成交时间 ===
	uow := newRecordingDirectTradeUnitOfWork(SideSell, "seller-1", "book-1")
	input := validPurchaseInput()
	input.CreatedAt = testTime.Add(-time.Nanosecond)
	_, err := PurchaseSellListing(context.Background(), uow, input)
	assertValidationError(t, err, "listing.updated_at")
	assertNoDirectTradeWrites(t, uow.state)
}

func TestDirectTradeFailuresRollBackAndPreserveCause(t *testing.T) {
	t.Parallel()

	// === 定义事务与依赖失败点 ===
	stages := []string{
		"with_tx", "read_listings_for_update", "create_match_request", "create_trade",
		"update_match_request", "update_listing", "commit",
	}
	for _, stage := range stages {
		t.Run(stage, func(t *testing.T) {
			t.Parallel()
			cause := errors.New(stage + " failed")
			uow := newRecordingDirectTradeUnitOfWork(SideSell, "seller-1", "book-1")
			uow.state.failStage, uow.state.failure = stage, cause
			result, err := PurchaseSellListing(context.Background(), uow, validPurchaseInput())
			if !errors.Is(err, cause) {
				t.Fatalf("PurchaseSellListing() error = %v, want %v", err, cause)
			}
			if result != (DirectTradeResult{}) || uow.committed {
				t.Fatalf("failed result = %+v, committed = %t", result, uow.committed)
			}
			assertNoDirectTradeWrites(t, uow.state)
		})
	}
}

func TestDirectTradeRollsBackEmptyOrderReference(t *testing.T) {
	t.Parallel()

	// === 返回非法订单引用 ===
	uow := newRecordingDirectTradeUnitOfWork(SideSell, "seller-1", "book-1")
	uow.state.orderRef = ""
	result, err := PurchaseSellListing(context.Background(), uow, validPurchaseInput())
	assertValidationError(t, err, "match_request.order_ref")
	if result != (DirectTradeResult{}) {
		t.Fatalf("PurchaseSellListing() result = %+v", result)
	}
	assertNoDirectTradeWrites(t, uow.state)
}

func TestDirectTradeRepeatedSubmissionCreatesOneTrade(t *testing.T) {
	t.Parallel()

	// === 重复购买同一挂单 ===
	uow := newRecordingDirectTradeUnitOfWork(SideSell, "seller-1", "book-1")
	if _, err := PurchaseSellListing(context.Background(), uow, validPurchaseInput()); err != nil {
		t.Fatal(err)
	}
	if _, err := PurchaseSellListing(context.Background(), uow, validPurchaseInput()); !errors.Is(err, ErrListingUnavailable) {
		t.Fatalf("repeated trade error = %v, want ErrListingUnavailable", err)
	}
	if len(uow.state.requests) != 1 || len(uow.state.orders) != 1 {
		t.Fatalf("repeated trade persisted requests:%d orders:%d", len(uow.state.requests), len(uow.state.orders))
	}
}

func TestConcurrentDirectTradesCreateOneTrade(t *testing.T) {
	t.Parallel()

	// === 竞争同一目标挂单 ===
	uow := newRecordingDirectTradeUnitOfWork(SideSell, "seller-1", "book-1")
	errorsByBuyer := make(chan error, 2)
	for _, buyerID := range []string{"buyer-1", "buyer-2"} {
		go func() {
			input := validPurchaseInput()
			input.MatchRequestID = "request-" + buyerID
			input.BuyerID = buyerID
			_, err := PurchaseSellListing(context.Background(), uow, input)
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
	if successes != 1 || unavailable != 1 || len(uow.state.requests) != 1 || len(uow.state.orders) != 1 {
		t.Fatalf("concurrent results = success:%d unavailable:%d requests:%d orders:%d", successes, unavailable, len(uow.state.requests), len(uow.state.orders))
	}
}

func TestDirectTradeValidatesInputBeforeWithTx(t *testing.T) {
	t.Parallel()

	// === 定义无效调用 ===
	validUOW := newRecordingDirectTradeUnitOfWork(SideSell, "seller-1", "book-1")
	tests := []struct {
		name      string
		uow       DirectTradeUnitOfWork
		input     PurchaseSellListingInput
		wantField string
	}{
		{name: "nil unit of work", input: validPurchaseInput(), wantField: "direct_trade_unit_of_work"},
		{name: "missing request", uow: validUOW, input: replacePurchaseInput(validPurchaseInput(), func(i *PurchaseSellListingInput) { i.MatchRequestID = "" }), wantField: "match_request.id"},
		{name: "missing listing", uow: validUOW, input: replacePurchaseInput(validPurchaseInput(), func(i *PurchaseSellListingInput) { i.SellListingID = "" }), wantField: "listing.id"},
		{name: "missing buyer", uow: validUOW, input: replacePurchaseInput(validPurchaseInput(), func(i *PurchaseSellListingInput) { i.BuyerID = "" }), wantField: "direct_trade.initiator_id"},
		{name: "zero time", uow: validUOW, input: replacePurchaseInput(validPurchaseInput(), func(i *PurchaseSellListingInput) { i.CreatedAt = time.Time{} }), wantField: "match_request.created_at"},
	}

	// === 验证事务调用边界 ===
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := PurchaseSellListing(context.Background(), tt.uow, tt.input)
			assertValidationError(t, err, tt.wantField)
		})
	}
	if validUOW.withTxCalls != 0 {
		t.Fatalf("WithTx called during input validation: %d", validUOW.withTxCalls)
	}
}

func TestAcceptBuyListingRequiresDeliveryResourceBeforeWithTx(t *testing.T) {
	t.Parallel()

	// === 省略真实交付资源 ===
	uow := newRecordingDirectTradeUnitOfWork(SideBuy, "buyer-1", "wanted-book-1")
	_, err := AcceptBuyListing(context.Background(), uow, AcceptBuyListingInput{
		MatchRequestID: "request-1", BuyListingID: "listing-1",
		SellerID: "seller-1", CreatedAt: testTime.Add(time.Minute),
	})
	assertValidationError(t, err, "direct_trade.resource_ref")
	if uow.withTxCalls != 0 {
		t.Fatalf("WithTx called during input validation: %d", uow.withTxCalls)
	}
}

func TestDirectTradeRejectsInvalidTransactionCallback(t *testing.T) {
	t.Parallel()

	// === 定义损坏的工作单元 ===
	tests := []struct {
		name      string
		uow       DirectTradeUnitOfWork
		wantField string
	}{
		{
			name: "callback not executed",
			uow: directTradeUnitOfWorkFunc(func(context.Context, func(DirectTradeTransaction) error) error {
				return nil
			}),
			wantField: "direct_trade_unit_of_work",
		},
		{
			name: "nil transaction",
			uow: directTradeUnitOfWorkFunc(func(_ context.Context, run func(DirectTradeTransaction) error) error {
				return run(nil)
			}),
			wantField: "direct_trade_transaction",
		},
	}

	// === 验证协议损坏不返回成功 ===
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := PurchaseSellListing(context.Background(), tt.uow, validPurchaseInput())
			assertValidationError(t, err, tt.wantField)
			if result != (DirectTradeResult{}) {
				t.Fatalf("PurchaseSellListing() result = %+v", result)
			}
		})
	}
}

type directTradeState struct {
	listing          ListingRecord
	requests         map[string]MatchRequestRecord
	orders           map[string]string
	proposal         MatchProposal
	tradeInput       TradeInput
	calls            []string
	failStage        string
	failure          error
	orderRef         string
	matchResult      bool
	validationErr    error
	useLockedRecords bool
	lockedRecords    []ListingRecord
}

func newDirectTradeState(side Side, ownerID, resourceRef string) *directTradeState {
	return &directTradeState{
		listing: ListingRecord{
			ID: "listing-1", OwnerID: ownerID, ResourceRef: resourceRef,
			Side: side, PriceCents: 2000, Status: ListingStatusOpen,
			CreatedAt: testTime, UpdatedAt: testTime,
		},
		requests: make(map[string]MatchRequestRecord), orders: make(map[string]string),
		orderRef: "order-request-1", matchResult: true,
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
	clone.lockedRecords = slices.Clone(s.lockedRecords)
	return &clone
}

func (s *directTradeState) fail(stage string) error {
	if s.failStage == stage {
		return s.failure
	}
	return nil
}

type recordingDirectTradeUnitOfWork struct {
	mu           sync.Mutex
	state        *directTradeState
	withTxCalls  int
	callbackOpen bool
	readInsideTx bool
	committed    bool
	lockIDs      []string
}

func newRecordingDirectTradeUnitOfWork(side Side, ownerID, resourceRef string) *recordingDirectTradeUnitOfWork {
	return &recordingDirectTradeUnitOfWork{state: newDirectTradeState(side, ownerID, resourceRef)}
}

func (uow *recordingDirectTradeUnitOfWork) WithTx(
	ctx context.Context,
	run func(DirectTradeTransaction) error,
) error {
	// === 串行并暂存事务写入 ===
	uow.mu.Lock()
	defer uow.mu.Unlock()
	uow.withTxCalls++
	uow.state.calls = append(uow.state.calls, "with_tx")
	if err := uow.state.fail("with_tx"); err != nil {
		return err
	}
	staged := uow.state.clone()
	uow.callbackOpen = true
	err := run(&recordingDirectTradeTransaction{state: staged, uow: uow})
	uow.callbackOpen = false
	uow.state.calls = append(uow.state.calls, staged.calls...)
	uow.state.proposal = staged.proposal
	uow.state.tradeInput = staged.tradeInput
	if err != nil {
		return err
	}

	// === 提交或回滚暂存状态 ===
	uow.state.calls = append(uow.state.calls, "commit")
	if err := uow.state.fail("commit"); err != nil {
		return err
	}
	uow.state.listing, uow.state.requests, uow.state.orders = staged.listing, staged.requests, staged.orders
	uow.committed = true
	return nil
}

type recordingDirectTradeTransaction struct {
	state *directTradeState
	uow   *recordingDirectTradeUnitOfWork
}

type directTradeUnitOfWorkFunc func(context.Context, func(DirectTradeTransaction) error) error

func (f directTradeUnitOfWorkFunc) WithTx(ctx context.Context, run func(DirectTradeTransaction) error) error {
	return f(ctx, run)
}

func (tx *recordingDirectTradeTransaction) ReadListingsForUpdate(_ context.Context, ids []string) ([]ListingRecord, error) {
	tx.state.calls = append(tx.state.calls, "read_listings_for_update")
	tx.uow.readInsideTx = tx.uow.callbackOpen
	tx.uow.lockIDs = slices.Clone(ids)
	if err := tx.state.fail("read_listings_for_update"); err != nil {
		return nil, err
	}
	if tx.state.useLockedRecords {
		return slices.Clone(tx.state.lockedRecords), nil
	}
	return []ListingRecord{tx.state.listing}, nil
}

func (tx *recordingDirectTradeTransaction) ValidateMatch(_ context.Context, proposal MatchProposal) (bool, error) {
	tx.state.calls = append(tx.state.calls, "validate_match")
	tx.state.proposal = proposal
	return tx.state.matchResult, tx.state.validationErr
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
	tx.state.tradeInput = input
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

func assertCompletedDirectTrade(
	t *testing.T,
	result DirectTradeResult,
	state *directTradeState,
	proposal MatchProposal,
	wantResourceRef, wantBuyListing, wantSellListing string,
) {
	t.Helper()
	if result.Listing == nil || result.MatchRequest == nil {
		t.Fatalf("direct trade result = %+v", result)
	}
	if result.Listing.Status() != ListingStatusMatched || state.listing.Status != ListingStatusMatched {
		t.Fatalf("listing statuses = (%v, %v), want matched", result.Listing.Status(), state.listing.Status)
	}
	request := result.MatchRequest
	if request.Status() != MatchRequestStatusCompleted || request.OrderRef() != "order-request-1" {
		t.Fatalf("request completion = (%v, %q)", request.Status(), request.OrderRef())
	}
	if request.Source() != proposal.Source || request.BuyerID() != proposal.Buyer.ParticipantID || request.SellerID() != proposal.Seller.ParticipantID {
		t.Fatalf("request trade facts = (%v, %q, %q)", request.Source(), request.BuyerID(), request.SellerID())
	}
	if request.ResourceRef() != wantResourceRef || request.BuyListingID() != wantBuyListing || request.SellListingID() != wantSellListing {
		t.Fatalf("request associations = (%q, %q, %q)", request.ResourceRef(), request.BuyListingID(), request.SellListingID())
	}
	if request.TradePrice() != proposal.TradePrice || request.BuyerConfirmedAt() != testTime.Add(time.Minute) || request.SellerConfirmedAt() != testTime.Add(time.Minute) {
		t.Fatalf("request price and confirmations = (%d, %v, %v)", request.TradePrice().Cents(), request.BuyerConfirmedAt(), request.SellerConfirmedAt())
	}
	if state.tradeInput.ResourceRef != wantResourceRef || state.tradeInput.MatchRequestID != request.ID() {
		t.Fatalf("trade input = %+v", state.tradeInput)
	}
	if len(state.requests) != 1 || len(state.orders) != 1 {
		t.Fatalf("persisted trade = requests:%d orders:%d", len(state.requests), len(state.orders))
	}
}

func assertNoDirectTradeWrites(t *testing.T, state *directTradeState) {
	t.Helper()
	if state.listing.Status != ListingStatusOpen || len(state.requests) != 0 || len(state.orders) != 0 {
		t.Fatalf("direct trade wrote state: listing:%v requests:%d orders:%d", state.listing.Status, len(state.requests), len(state.orders))
	}
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
