package engine

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// PurchaseSellListingInput 是买方直接购买 Sell Listing 所需的宿主输入。
type PurchaseSellListingInput struct {
	// === runtime 字段 ===
	MatchRequestID string
	SellListingID  string
	BuyerID        string
	CreatedAt      time.Time
}

// AcceptBuyListingInput 是卖方直接接受 Buy Listing 所需的宿主输入。
type AcceptBuyListingInput struct {
	// === runtime 字段 ===
	MatchRequestID string
	BuyListingID   string
	SellerID       string
	ResourceRef    string
	CreatedAt      time.Time
}

// DirectTradeResult 返回直接成交后的通用 Listing 与 MatchRequest 状态。
type DirectTradeResult struct {
	// === runtime 字段 ===
	Listing      *Listing
	MatchRequest *MatchRequest
}

type directTradeInput struct {
	// === runtime 字段 ===
	matchRequestID string
	listingID      string
	initiatorID    string
	resourceRef    string
	source         MatchSource
	expectedSide   Side
	createdAt      time.Time
}

// PurchaseSellListing 在宿主工作单元内按卖方报价直接购买开放的 Sell Listing。
// 宿主正常拒绝撮合时返回可由 errors.Is 识别的 ErrMatchRejected。
func PurchaseSellListing(
	ctx context.Context,
	uow DirectTradeUnitOfWork,
	input PurchaseSellListingInput,
) (DirectTradeResult, error) {
	return completeDirectTrade(ctx, uow, directTradeInput{
		matchRequestID: input.MatchRequestID,
		listingID:      input.SellListingID,
		initiatorID:    input.BuyerID,
		source:         MatchSourceBuyerPurchase,
		expectedSide:   SideSell,
		createdAt:      input.CreatedAt,
	})
}

// AcceptBuyListing 在宿主工作单元内按买方报价接受开放的 Buy Listing。
// 宿主正常拒绝撮合时返回可由 errors.Is 识别的 ErrMatchRejected。
func AcceptBuyListing(
	ctx context.Context,
	uow DirectTradeUnitOfWork,
	input AcceptBuyListingInput,
) (DirectTradeResult, error) {
	return completeDirectTrade(ctx, uow, directTradeInput{
		matchRequestID: input.MatchRequestID,
		listingID:      input.BuyListingID,
		initiatorID:    input.SellerID,
		resourceRef:    input.ResourceRef,
		source:         MatchSourceSellerAccept,
		expectedSide:   SideBuy,
		createdAt:      input.CreatedAt,
	})
}

func completeDirectTrade(
	ctx context.Context,
	uow DirectTradeUnitOfWork,
	input directTradeInput,
) (DirectTradeResult, error) {
	// === 校验直接成交输入 ===
	if uow == nil {
		return DirectTradeResult{}, &ValidationError{Field: "direct_trade_unit_of_work", Reason: "must not be nil"}
	}
	if err := required("match_request.id", input.matchRequestID); err != nil {
		return DirectTradeResult{}, err
	}
	if err := required("listing.id", input.listingID); err != nil {
		return DirectTradeResult{}, err
	}
	if err := required("direct_trade.initiator_id", input.initiatorID); err != nil {
		return DirectTradeResult{}, err
	}
	if input.source == MatchSourceSellerAccept {
		if err := required("direct_trade.resource_ref", input.resourceRef); err != nil {
			return DirectTradeResult{}, err
		}
	}
	if input.createdAt.IsZero() {
		return DirectTradeResult{}, &ValidationError{Field: "match_request.created_at", Reason: "must not be zero"}
	}

	// === 执行并提交直接成交事务 ===
	var result DirectTradeResult
	callbackCompleted := false
	err := uow.WithTx(ctx, func(tx DirectTradeTransaction) error {
		if tx == nil {
			return &ValidationError{Field: "direct_trade_transaction", Reason: "must not be nil"}
		}
		completed, err := completeDirectTradeInTx(ctx, tx, input)
		if err != nil {
			return err
		}
		result = completed
		callbackCompleted = true
		return nil
	})
	if err != nil {
		return DirectTradeResult{}, err
	}
	if !callbackCompleted {
		return DirectTradeResult{}, &ValidationError{
			Field:  "direct_trade_unit_of_work",
			Reason: "must execute the transaction callback",
		}
	}
	return result, nil
}

func completeDirectTradeInTx(
	ctx context.Context,
	tx DirectTradeTransaction,
	input directTradeInput,
) (DirectTradeResult, error) {
	// === 锁定并恢复目标挂单 ===
	records, err := tx.ReadListingsForUpdate(ctx, []string{input.listingID})
	if err != nil {
		if errors.Is(err, ErrListingNotFound) {
			return DirectTradeResult{}, fmt.Errorf("read listing %q for direct trade: %w", input.listingID, ErrListingNotFound)
		}
		return DirectTradeResult{}, fmt.Errorf("read listing %q for direct trade: %w", input.listingID, err)
	}
	listings, err := restoreListingRecords([]string{input.listingID}, records)
	if err != nil {
		return DirectTradeResult{}, fmt.Errorf("restore listing %q for direct trade: %w", input.listingID, err)
	}
	listing := listings[0]
	if listing.Side() != input.expectedSide || listing.Status() != ListingStatusOpen {
		return DirectTradeResult{}, fmt.Errorf("trade listing %q: %w", input.listingID, ErrListingUnavailable)
	}
	if err := validateTimestamp("listing.updated_at", input.createdAt, listing.UpdatedAt()); err != nil {
		return DirectTradeResult{}, err
	}

	// === 归一化成交事实 ===
	buyerID, sellerID, resourceRef := input.initiatorID, listing.OwnerID(), listing.ResourceRef()
	buyListingID, sellListingID := "", listing.ID()
	if input.source == MatchSourceSellerAccept {
		buyerID, sellerID, resourceRef = listing.OwnerID(), input.initiatorID, input.resourceRef
		buyListingID, sellListingID = listing.ID(), ""
	}
	proposal := MatchProposal{
		Buyer: MatchParticipant{ParticipantID: buyerID},
		Seller: MatchParticipant{
			ParticipantID: sellerID,
			ListingID:     sellListingID,
			ResourceRef:   resourceRef,
		},
		TradePrice: listing.Price(), Source: input.source,
	}
	if input.source == MatchSourceSellerAccept {
		proposal.Buyer.ListingID = buyListingID
		proposal.Buyer.ResourceRef = listing.ResourceRef()
		proposal.Seller.ListingID = ""
	}
	matched, err := tx.ValidateMatch(ctx, proposal)
	if err != nil {
		return DirectTradeResult{}, fmt.Errorf("validate direct trade: %w", err)
	}
	if !matched {
		return DirectTradeResult{}, fmt.Errorf("validate direct trade: %w", ErrMatchRejected)
	}

	// === 创建已确认请求与宿主订单 ===
	request, err := NewMatchRequest(MatchRequestInput{
		ID: input.matchRequestID, BuyerID: buyerID, SellerID: sellerID,
		ResourceRef: resourceRef, BuyListingID: buyListingID, SellListingID: sellListingID,
		TradePrice: listing.Price(), Source: input.source, CreatedAt: input.createdAt,
	})
	if err != nil {
		return DirectTradeResult{}, err
	}
	if err := tx.CreateMatchRequest(ctx, request.Record()); err != nil {
		return DirectTradeResult{}, fmt.Errorf("create match request %q: %w", request.ID(), err)
	}
	orderRef, err := tx.CreateTrade(ctx, TradeInput{
		MatchRequestID: request.ID(), BuyerID: buyerID, SellerID: sellerID,
		ResourceRef: resourceRef, TradePrice: listing.Price(), Source: input.source,
	})
	if err != nil {
		return DirectTradeResult{}, fmt.Errorf("create trade for match request %q: %w", request.ID(), err)
	}
	if err := request.Complete(orderRef, input.createdAt); err != nil {
		return DirectTradeResult{}, err
	}
	if err := tx.UpdateMatchRequest(ctx, request.Record()); err != nil {
		return DirectTradeResult{}, fmt.Errorf("update match request %q: %w", request.ID(), err)
	}

	// === 完成目标挂单 ===
	if err := listing.Reserve(input.createdAt); err != nil {
		return DirectTradeResult{}, err
	}
	if err := listing.MarkMatched(input.createdAt); err != nil {
		return DirectTradeResult{}, err
	}
	if err := tx.UpdateListing(ctx, listing.Record()); err != nil {
		return DirectTradeResult{}, fmt.Errorf("update listing %q: %w", listing.ID(), err)
	}
	return DirectTradeResult{Listing: listing, MatchRequest: request}, nil
}
