package engine

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// PurchaseSellListingInput 是买方直接购买 Sell Listing 所需的宿主输入。
type PurchaseSellListingInput struct {
	MatchRequestID string
	SellListingID  string
	BuyerID        string
	CreatedAt      time.Time
}

// AcceptBuyListingInput 是卖方直接接受 Buy Listing 所需的宿主输入。
type AcceptBuyListingInput struct {
	MatchRequestID string
	BuyListingID   string
	SellerID       string
	ResourceRef    string
	CreatedAt      time.Time
}

// DirectTradeResult 返回直接成交后的通用 Listing 与 MatchRequest 状态。
type DirectTradeResult struct {
	Listing      *Listing
	MatchRequest *MatchRequest
}

type directTradeInput struct {
	matchRequestID string
	listingID      string
	initiatorID    string
	resourceRef    string
	source         MatchSource
	expectedSide   Side
	createdAt      time.Time
}

// PurchaseSellListing 在调用方提供的事务内按卖方报价直接购买开放的 Sell Listing。
func PurchaseSellListing(
	ctx context.Context,
	tx DirectTradeTransaction,
	input PurchaseSellListingInput,
) (DirectTradeResult, error) {
	return completeDirectTrade(ctx, tx, directTradeInput{
		matchRequestID: input.MatchRequestID,
		listingID:      input.SellListingID,
		initiatorID:    input.BuyerID,
		source:         MatchSourceBuyerPurchase,
		expectedSide:   SideSell,
		createdAt:      input.CreatedAt,
	})
}

// AcceptBuyListing 在调用方提供的事务内按买方报价接受开放的 Buy Listing。
func AcceptBuyListing(
	ctx context.Context,
	tx DirectTradeTransaction,
	input AcceptBuyListingInput,
) (DirectTradeResult, error) {
	return completeDirectTrade(ctx, tx, directTradeInput{
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
	tx DirectTradeTransaction,
	input directTradeInput,
) (DirectTradeResult, error) {
	// === 校验直接成交输入 ===
	if tx == nil {
		return DirectTradeResult{}, &ValidationError{Field: "direct_trade_transaction", Reason: "must not be nil"}
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

	// === 锁定并恢复目标挂单 ===
	record, err := tx.ReadListingForUpdate(ctx, input.listingID)
	if err != nil {
		if errors.Is(err, ErrListingNotFound) {
			return DirectTradeResult{}, fmt.Errorf("read listing %q for direct trade: %w", input.listingID, ErrListingNotFound)
		}
		return DirectTradeResult{}, fmt.Errorf("read listing %q for direct trade: %w", input.listingID, err)
	}
	listing, err := RestoreListing(record)
	if err != nil {
		return DirectTradeResult{}, fmt.Errorf("restore listing %q for direct trade: %w", input.listingID, err)
	}
	if listing.ID() != input.listingID {
		return DirectTradeResult{}, &ValidationError{Field: "listing.id", Reason: "must match the requested id"}
	}
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
		BuyerID: buyerID, SellerID: sellerID, ResourceRef: resourceRef,
		BuyListingID: buyListingID, SellListingID: sellListingID,
		TradePrice: listing.Price(), Source: input.source,
	}
	if err := tx.EvaluateMatch(ctx, proposal); err != nil {
		return DirectTradeResult{}, fmt.Errorf("evaluate direct trade: %w", err)
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
