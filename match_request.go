package engine

import "time"

// MatchSource 表示撮合请求的发起方式。零值不是有效来源。
type MatchSource uint8

const (
	// MatchSourceAutomatic 表示引擎自动提出的撮合请求。
	MatchSourceAutomatic MatchSource = iota + 1
	// MatchSourceSellerAccept 表示卖方接受买方需求。
	MatchSourceSellerAccept
	// MatchSourceBuyerPurchase 表示买方购买卖方供给。
	MatchSourceBuyerPurchase
)

// String 返回撮合来源的稳定文本值；无效来源返回 unknown。
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

// MatchRequestStatus 表示撮合请求状态。零值不是有效状态。
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

// String 返回撮合请求状态的稳定文本值；无效状态返回 unknown。
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

func (s MatchRequestStatus) valid() bool {
	return s >= MatchRequestStatusPending && s <= MatchRequestStatusCancelled
}

// MatchRequestInput 是创建撮合请求所需的通用事实。
// 自动撮合必须引用买卖双方 Listing；卖方接单只引用 Buy Listing；买方购买只引用 Sell Listing。
type MatchRequestInput struct {
	// === runtime 字段 ===
	ID            string
	BuyerID       string
	SellerID      string
	ResourceRef   string
	BuyListingID  string
	SellListingID string
	TradePrice    Amount
	Source        MatchSource
	CreatedAt     time.Time
}

// MatchRequestRecord 是 MatchRequest 的逻辑持久化表示。
type MatchRequestRecord struct {
	// === persistent 字段 ===
	ID                string
	BuyerID           string
	SellerID          string
	ResourceRef       string
	BuyListingID      string
	SellListingID     string
	TradePriceCents   int64
	Source            MatchSource
	BuyerConfirmedAt  time.Time
	SellerConfirmedAt time.Time
	Status            MatchRequestStatus
	OrderRef          string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// MatchRequest 表示一次买卖撮合。零值无效，请求必须通过 NewMatchRequest 创建
// 或通过 RestoreMatchRequest 恢复。
type MatchRequest struct {
	// === persistent 字段 ===
	id                string
	buyerID           string
	sellerID          string
	resourceRef       string
	buyListingID      string
	sellListingID     string
	tradePrice        Amount
	source            MatchSource
	buyerConfirmedAt  time.Time
	sellerConfirmedAt time.Time
	status            MatchRequestStatus
	orderRef          string
	createdAt         time.Time
	updatedAt         time.Time
}

// NewMatchRequest 创建状态为 pending 的撮合请求。直接成交请求在创建时已包含双方确认，
// 自动撮合请求初始时双方均未确认。
func NewMatchRequest(input MatchRequestInput) (*MatchRequest, error) {
	// === 校验请求事实 ===
	if err := validateMatchRequestIdentity(
		input.ID, input.BuyerID, input.SellerID, input.ResourceRef,
		input.BuyListingID, input.SellListingID, input.TradePrice, input.Source,
	); err != nil {
		return nil, err
	}
	if input.CreatedAt.IsZero() {
		return nil, &ValidationError{Field: "match_request.created_at", Reason: "must not be zero"}
	}

	// === 创建待处理请求 ===
	request := &MatchRequest{
		id: input.ID, buyerID: input.BuyerID, sellerID: input.SellerID,
		resourceRef: input.ResourceRef, buyListingID: input.BuyListingID,
		sellListingID: input.SellListingID, tradePrice: input.TradePrice,
		source: input.Source, status: MatchRequestStatusPending,
		createdAt: input.CreatedAt, updatedAt: input.CreatedAt,
	}
	if input.Source != MatchSourceAutomatic {
		request.buyerConfirmedAt = input.CreatedAt
		request.sellerConfirmedAt = input.CreatedAt
	}
	return request, nil
}

// RestoreMatchRequest 从逻辑持久化表示恢复并校验 MatchRequest。
func RestoreMatchRequest(record MatchRequestRecord) (*MatchRequest, error) {
	// === 校验持久化事实 ===
	price, err := NewAmount(record.TradePriceCents)
	if err != nil {
		return nil, &ValidationError{Field: "match_request.trade_price", Reason: "must be positive"}
	}
	if err := validateMatchRequestIdentity(
		record.ID, record.BuyerID, record.SellerID, record.ResourceRef,
		record.BuyListingID, record.SellListingID, price, record.Source,
	); err != nil {
		return nil, err
	}
	if !record.Status.valid() {
		return nil, &ValidationError{Field: "match_request.status", Reason: "must be pending, completed, declined, or cancelled"}
	}
	if record.CreatedAt.IsZero() {
		return nil, &ValidationError{Field: "match_request.created_at", Reason: "must not be zero"}
	}
	if record.UpdatedAt.IsZero() || record.UpdatedAt.Before(record.CreatedAt) {
		return nil, &ValidationError{Field: "match_request.updated_at", Reason: "must not be before created_at"}
	}
	if err := validateConfirmationTime("match_request.buyer_confirmed_at", record.BuyerConfirmedAt, record.CreatedAt, record.UpdatedAt); err != nil {
		return nil, err
	}
	if err := validateConfirmationTime("match_request.seller_confirmed_at", record.SellerConfirmedAt, record.CreatedAt, record.UpdatedAt); err != nil {
		return nil, err
	}
	if record.Source != MatchSourceAutomatic && (record.BuyerConfirmedAt.IsZero() || record.SellerConfirmedAt.IsZero()) {
		return nil, &ValidationError{Field: "match_request.confirmations", Reason: "direct trades require both confirmations"}
	}
	if record.Status == MatchRequestStatusCompleted {
		if record.BuyerConfirmedAt.IsZero() || record.SellerConfirmedAt.IsZero() {
			return nil, &ValidationError{Field: "match_request.confirmations", Reason: "completed requests require both confirmations"}
		}
		if err := required("match_request.order_ref", record.OrderRef); err != nil {
			return nil, err
		}
	} else if record.OrderRef != "" {
		return nil, &ValidationError{Field: "match_request.order_ref", Reason: "must be empty before completion"}
	}

	// === 恢复请求实体 ===
	return &MatchRequest{
		id: record.ID, buyerID: record.BuyerID, sellerID: record.SellerID,
		resourceRef: record.ResourceRef, buyListingID: record.BuyListingID,
		sellListingID: record.SellListingID, tradePrice: price, source: record.Source,
		buyerConfirmedAt: record.BuyerConfirmedAt, sellerConfirmedAt: record.SellerConfirmedAt,
		status: record.Status, orderRef: record.OrderRef,
		createdAt: record.CreatedAt, updatedAt: record.UpdatedAt,
	}, nil
}

// Record 返回可交给宿主适配器持久化的 MatchRequest 表示。
func (r MatchRequest) Record() MatchRequestRecord {
	return MatchRequestRecord{
		ID: r.id, BuyerID: r.buyerID, SellerID: r.sellerID, ResourceRef: r.resourceRef,
		BuyListingID: r.buyListingID, SellListingID: r.sellListingID,
		TradePriceCents: r.tradePrice.Cents(), Source: r.source,
		BuyerConfirmedAt: r.buyerConfirmedAt, SellerConfirmedAt: r.sellerConfirmedAt,
		Status: r.status, OrderRef: r.orderRef,
		CreatedAt: r.createdAt, UpdatedAt: r.updatedAt,
	}
}

// ID 返回撮合请求 ID。
func (r MatchRequest) ID() string { return r.id }

// BuyerID 返回买方参与者 ID。
func (r MatchRequest) BuyerID() string { return r.buyerID }

// SellerID 返回卖方参与者 ID。
func (r MatchRequest) SellerID() string { return r.sellerID }

// ResourceRef 返回实际成交资源的宿主引用。
func (r MatchRequest) ResourceRef() string { return r.resourceRef }

// BuyListingID 返回真实 Buy Listing ID；直接购买时为空。
func (r MatchRequest) BuyListingID() string { return r.buyListingID }

// SellListingID 返回真实 Sell Listing ID；直接接单时为空。
func (r MatchRequest) SellListingID() string { return r.sellListingID }

// TradePrice 返回成交价。
func (r MatchRequest) TradePrice() Amount { return r.tradePrice }

// Source 返回撮合来源。
func (r MatchRequest) Source() MatchSource { return r.source }

// BuyerConfirmed 返回买方是否已确认。
func (r MatchRequest) BuyerConfirmed() bool { return !r.buyerConfirmedAt.IsZero() }

// SellerConfirmed 返回卖方是否已确认。
func (r MatchRequest) SellerConfirmed() bool { return !r.sellerConfirmedAt.IsZero() }

// BuyerConfirmedAt 返回买方确认时间；未确认时为零值。
func (r MatchRequest) BuyerConfirmedAt() time.Time { return r.buyerConfirmedAt }

// SellerConfirmedAt 返回卖方确认时间；未确认时为零值。
func (r MatchRequest) SellerConfirmedAt() time.Time { return r.sellerConfirmedAt }

// Status 返回撮合请求状态。
func (r MatchRequest) Status() MatchRequestStatus { return r.status }

// OrderRef 返回成交后的宿主订单引用。
func (r MatchRequest) OrderRef() string { return r.orderRef }

// CreatedAt 返回撮合请求创建时间。
func (r MatchRequest) CreatedAt() time.Time { return r.createdAt }

// UpdatedAt 返回撮合请求最近更新时间。
func (r MatchRequest) UpdatedAt() time.Time { return r.updatedAt }

// Confirm 记录指定交易方向参与方的确认。已经确认的参与方再次确认是幂等操作。
func (r *MatchRequest) Confirm(side Side, at time.Time) error {
	// === 校验确认操作 ===
	if err := r.ensurePending(MatchRequestStatusPending.String()); err != nil {
		return err
	}
	if !side.valid() {
		return &ValidationError{Field: "match_request.confirmation_side", Reason: "must be buy or sell"}
	}
	if side == SideBuy && r.BuyerConfirmed() || side == SideSell && r.SellerConfirmed() {
		return nil
	}
	if err := validateTimestamp("match_request.updated_at", at, r.updatedAt); err != nil {
		return err
	}

	// === 记录参与方确认 ===
	if side == SideBuy {
		r.buyerConfirmedAt = at
	} else {
		r.sellerConfirmedAt = at
	}
	r.updatedAt = at
	return nil
}

// Complete 将双方已确认的待处理请求标记为已成交。
func (r *MatchRequest) Complete(orderRef string, at time.Time) error {
	// === 校验完成条件 ===
	if err := r.ensurePending(MatchRequestStatusCompleted.String()); err != nil {
		return err
	}
	if err := required("match_request.order_ref", orderRef); err != nil {
		return err
	}
	if !r.BuyerConfirmed() || !r.SellerConfirmed() {
		return ErrConfirmationRequired
	}
	if err := validateTimestamp("match_request.updated_at", at, r.updatedAt); err != nil {
		return err
	}

	// === 完成撮合请求 ===
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
	// === 校验终态转换 ===
	if err := r.ensurePending(status.String()); err != nil {
		return err
	}
	if err := validateTimestamp("match_request.updated_at", at, r.updatedAt); err != nil {
		return err
	}

	// === 更新请求终态 ===
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

func validateMatchRequestIdentity(
	id, buyerID, sellerID, resourceRef, buyListingID, sellListingID string,
	tradePrice Amount,
	source MatchSource,
) error {
	// === 校验必填事实 ===
	for _, value := range []struct{ field, value string }{
		{field: "match_request.id", value: id},
		{field: "match_request.buyer_id", value: buyerID},
		{field: "match_request.seller_id", value: sellerID},
		{field: "match_request.resource_ref", value: resourceRef},
	} {
		if err := required(value.field, value.value); err != nil {
			return err
		}
	}
	if buyerID == sellerID {
		return &ValidationError{Field: "match_request.participant_ids", Reason: "must be different"}
	}
	if !tradePrice.valid() {
		return &ValidationError{Field: "match_request.trade_price", Reason: "must be positive"}
	}
	if !source.valid() {
		return &ValidationError{Field: "match_request.source", Reason: "must be a supported source"}
	}

	// === 校验来源关联 ===
	switch source {
	case MatchSourceAutomatic:
		if err := required("match_request.buy_listing_id", buyListingID); err != nil {
			return err
		}
		if err := required("match_request.sell_listing_id", sellListingID); err != nil {
			return err
		}
		if buyListingID == sellListingID {
			return &ValidationError{Field: "match_request.listing_ids", Reason: "must be different"}
		}
	case MatchSourceSellerAccept:
		if err := required("match_request.buy_listing_id", buyListingID); err != nil {
			return err
		}
		if sellListingID != "" {
			return &ValidationError{Field: "match_request.sell_listing_id", Reason: "must be empty for seller acceptance"}
		}
	case MatchSourceBuyerPurchase:
		if err := required("match_request.sell_listing_id", sellListingID); err != nil {
			return err
		}
		if buyListingID != "" {
			return &ValidationError{Field: "match_request.buy_listing_id", Reason: "must be empty for buyer purchase"}
		}
	}
	return nil
}

func validateConfirmationTime(field string, value, createdAt, updatedAt time.Time) error {
	if value.IsZero() {
		return nil
	}
	if value.Before(createdAt) || value.After(updatedAt) {
		return &ValidationError{Field: field, Reason: "must be between created_at and updated_at"}
	}
	return nil
}
