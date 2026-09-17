package engine

import "time"

// MatchSource 表示撮合请求的发起方式。
type MatchSource uint8

const (
	// MatchSourceAutomatic 表示引擎自动提出的撮合请求。
	MatchSourceAutomatic MatchSource = iota + 1
	// MatchSourceSellerAccept 表示卖方接受买方需求。
	MatchSourceSellerAccept
	// MatchSourceBuyerPurchase 表示买方购买卖方供给。
	MatchSourceBuyerPurchase
)

// String 返回撮合来源的稳定文本值。
func (s MatchSource) String() string {
	switch s {
	case MatchSourceAutomatic:
		return "automatic"
	case MatchSourceSellerAccept:
		return "seller_accept"
	case MatchSourceBuyerPurchase:
		return "buyer_purchase"
	default:
		return "unknown"
	}
}

func (s MatchSource) valid() bool {
	return s >= MatchSourceAutomatic && s <= MatchSourceBuyerPurchase
}

// MatchRequestStatus 表示撮合请求状态。
type MatchRequestStatus uint8

const (
	// MatchRequestStatusPending 表示请求等待处理。
	MatchRequestStatusPending MatchRequestStatus = iota + 1
	// MatchRequestStatusCompleted 表示请求已成交。
	MatchRequestStatusCompleted
	// MatchRequestStatusDeclined 表示请求被参与方拒绝。
	MatchRequestStatusDeclined
	// MatchRequestStatusCancelled 表示请求因其他操作取消。
	MatchRequestStatusCancelled
)

// String 返回撮合请求状态的稳定文本值。
func (s MatchRequestStatus) String() string {
	switch s {
	case MatchRequestStatusPending:
		return "pending"
	case MatchRequestStatusCompleted:
		return "completed"
	case MatchRequestStatusDeclined:
		return "declined"
	case MatchRequestStatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// MatchRequest 表示一对买卖挂单之间的撮合请求。
type MatchRequest struct {
	id              string
	buyListingID    string
	sellListingID   string
	tradePrice      Amount
	source          MatchSource
	buyerConfirmed  bool
	sellerConfirmed bool
	status          MatchRequestStatus
	orderRef        string
	createdAt       time.Time
	updatedAt       time.Time
}

// NewMatchRequest 创建状态为 pending 的撮合请求。
func NewMatchRequest(
	id, buyListingID, sellListingID string,
	tradePrice Amount,
	source MatchSource,
	createdAt time.Time,
) (*MatchRequest, error) {
	if err := required("match_request.id", id); err != nil {
		return nil, err
	}
	if err := required("match_request.buy_listing_id", buyListingID); err != nil {
		return nil, err
	}
	if err := required("match_request.sell_listing_id", sellListingID); err != nil {
		return nil, err
	}
	if buyListingID == sellListingID {
		return nil, &ValidationError{Field: "match_request.listing_ids", Reason: "must be different"}
	}
	if !tradePrice.valid() {
		return nil, &ValidationError{Field: "match_request.trade_price", Reason: "must be positive"}
	}
	if !source.valid() {
		return nil, &ValidationError{Field: "match_request.source", Reason: "must be a supported source"}
	}
	if createdAt.IsZero() {
		return nil, &ValidationError{Field: "match_request.created_at", Reason: "must not be zero"}
	}

	confirmed := source != MatchSourceAutomatic
	return &MatchRequest{
		id:              id,
		buyListingID:    buyListingID,
		sellListingID:   sellListingID,
		tradePrice:      tradePrice,
		source:          source,
		buyerConfirmed:  confirmed,
		sellerConfirmed: confirmed,
		status:          MatchRequestStatusPending,
		createdAt:       createdAt,
		updatedAt:       createdAt,
	}, nil
}

// ID 返回撮合请求 ID。
func (r MatchRequest) ID() string { return r.id }

// BuyListingID 返回买方挂单 ID。
func (r MatchRequest) BuyListingID() string { return r.buyListingID }

// SellListingID 返回卖方挂单 ID。
func (r MatchRequest) SellListingID() string { return r.sellListingID }

// TradePrice 返回成交价。
func (r MatchRequest) TradePrice() Amount { return r.tradePrice }

// Source 返回撮合来源。
func (r MatchRequest) Source() MatchSource { return r.source }

// BuyerConfirmed 返回买方是否已确认。
func (r MatchRequest) BuyerConfirmed() bool { return r.buyerConfirmed }

// SellerConfirmed 返回卖方是否已确认。
func (r MatchRequest) SellerConfirmed() bool { return r.sellerConfirmed }

// Status 返回撮合请求状态。
func (r MatchRequest) Status() MatchRequestStatus { return r.status }

// OrderRef 返回成交后的宿主订单引用。
func (r MatchRequest) OrderRef() string { return r.orderRef }

// CreatedAt 返回撮合请求创建时间。
func (r MatchRequest) CreatedAt() time.Time { return r.createdAt }

// UpdatedAt 返回撮合请求最近更新时间。
func (r MatchRequest) UpdatedAt() time.Time { return r.updatedAt }

// Confirm 记录指定交易方向参与方的确认；重复确认不改变请求。
func (r *MatchRequest) Confirm(side Side, at time.Time) error {
	if err := r.ensurePending(MatchRequestStatusPending.String()); err != nil {
		return err
	}
	if !side.valid() {
		return &ValidationError{Field: "match_request.confirmation_side", Reason: "must be buy or sell"}
	}
	if side == SideBuy && r.buyerConfirmed || side == SideSell && r.sellerConfirmed {
		return nil
	}
	if err := validateTimestamp("match_request.updated_at", at, r.updatedAt); err != nil {
		return err
	}

	if side == SideBuy {
		r.buyerConfirmed = true
	} else {
		r.sellerConfirmed = true
	}
	r.updatedAt = at
	return nil
}

// Complete 将双方已确认的待处理请求标记为已成交。
func (r *MatchRequest) Complete(orderRef string, at time.Time) error {
	if err := r.ensurePending(MatchRequestStatusCompleted.String()); err != nil {
		return err
	}
	if err := required("match_request.order_ref", orderRef); err != nil {
		return err
	}
	if !r.buyerConfirmed || !r.sellerConfirmed {
		return ErrConfirmationRequired
	}
	if err := validateTimestamp("match_request.updated_at", at, r.updatedAt); err != nil {
		return err
	}

	r.status = MatchRequestStatusCompleted
	r.orderRef = orderRef
	r.updatedAt = at
	return nil
}

// Decline 将待处理请求标记为已拒绝。
func (r *MatchRequest) Decline(at time.Time) error {
	return r.finish(MatchRequestStatusDeclined, at)
}

// Cancel 将待处理请求标记为已取消。
func (r *MatchRequest) Cancel(at time.Time) error {
	return r.finish(MatchRequestStatusCancelled, at)
}

func (r *MatchRequest) finish(status MatchRequestStatus, at time.Time) error {
	if err := r.ensurePending(status.String()); err != nil {
		return err
	}
	if err := validateTimestamp("match_request.updated_at", at, r.updatedAt); err != nil {
		return err
	}
	r.status = status
	r.updatedAt = at
	return nil
}

func (r *MatchRequest) ensurePending(to string) error {
	if r == nil {
		return &ValidationError{Field: "match_request", Reason: "must not be nil"}
	}
	if r.status != MatchRequestStatusPending {
		return &TransitionError{Entity: "match_request", From: r.status.String(), To: to}
	}
	return nil
}
