package engine

import "time"

// Amount 表示以人民币分为单位的正整数金额。零值无效，金额必须通过 NewAmount 创建。
type Amount struct {
	cents int64
}

// NewAmount 创建金额。cents 必须为正数，否则返回可由 errors.As 识别的 ValidationError。
func NewAmount(cents int64) (Amount, error) {
	if cents <= 0 {
		return Amount{}, &ValidationError{Field: "amount", Reason: "must be positive"}
	}
	return Amount{cents: cents}, nil
}

// Cents 返回金额的人民币分值。
func (a Amount) Cents() int64 {
	return a.cents
}

func (a Amount) valid() bool {
	return a.cents > 0
}

// Side 表示挂单的交易方向。零值不是有效方向。
type Side uint8

const (
	// SideBuy 表示买方需求。
	SideBuy Side = iota + 1
	// SideSell 表示卖方供给。
	SideSell
)

// String 返回交易方向的稳定文本值；无效方向返回 unknown。
func (s Side) String() string {
	switch s {
	case SideBuy:
		return "buy"
	case SideSell:
		return "sell"
	default:
		return "unknown"
	}
}

func (s Side) valid() bool {
	return s == SideBuy || s == SideSell
}

// ListingStatus 表示挂单状态。零值不是有效状态。
type ListingStatus uint8

const (
	// ListingStatusOpen 表示挂单可参与撮合。
	ListingStatusOpen ListingStatus = iota + 1
	// ListingStatusReserved 表示挂单已被一个待处理请求预留。
	ListingStatusReserved
	// ListingStatusMatched 表示挂单已成交。
	ListingStatusMatched
	// ListingStatusWithdrawn 表示挂单已撤回。
	ListingStatusWithdrawn
)

// String 返回挂单状态的稳定文本值；无效状态返回 unknown。
func (s ListingStatus) String() string {
	switch s {
	case ListingStatusOpen:
		return "open"
	case ListingStatusReserved:
		return "reserved"
	case ListingStatusMatched:
		return "matched"
	case ListingStatusWithdrawn:
		return "withdrawn"
	default:
		return "unknown"
	}
}

func (s ListingStatus) valid() bool {
	return s >= ListingStatusOpen && s <= ListingStatusWithdrawn
}

// Listing 表示一个通用买方需求或卖方供给。零值无效，挂单必须通过 NewListing 创建。
//
// 状态转换要求时间非零且不早于 UpdatedAt。非法时间或空接收者返回 ValidationError，
// 非法状态转换返回 TransitionError；失败不会修改挂单。
type Listing struct {
	id          string
	ownerID     string
	resourceRef string
	side        Side
	price       Amount
	status      ListingStatus
	createdAt   time.Time
	updatedAt   time.Time
}

// ListingRecord 是 Listing 的逻辑持久化表示。宿主适配器负责存取该表示，
// 业务代码必须通过 RestoreListing 恢复并校验实体。
type ListingRecord struct {
	ID          string
	OwnerID     string
	ResourceRef string
	Side        Side
	PriceCents  int64
	Status      ListingStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewListing 创建状态为 open 的挂单。ID、所有者 ID 和资源引用必须包含非空白字符，
// 方向和金额必须有效，创建时间不得为零；非法参数返回 ValidationError。
func NewListing(id, ownerID, resourceRef string, side Side, price Amount, createdAt time.Time) (*Listing, error) {
	// === 校验挂单字段 ===
	if err := required("listing.id", id); err != nil {
		return nil, err
	}
	if err := required("listing.owner_id", ownerID); err != nil {
		return nil, err
	}
	if err := required("listing.resource_ref", resourceRef); err != nil {
		return nil, err
	}
	if !side.valid() {
		return nil, &ValidationError{Field: "listing.side", Reason: "must be buy or sell"}
	}
	if !price.valid() {
		return nil, &ValidationError{Field: "listing.price", Reason: "must be positive"}
	}
	if createdAt.IsZero() {
		return nil, &ValidationError{Field: "listing.created_at", Reason: "must not be zero"}
	}

	// === 创建开放挂单 ===
	return &Listing{
		id:          id,
		ownerID:     ownerID,
		resourceRef: resourceRef,
		side:        side,
		price:       price,
		status:      ListingStatusOpen,
		createdAt:   createdAt,
		updatedAt:   createdAt,
	}, nil
}

// RestoreListing 从逻辑持久化表示恢复并校验 Listing。损坏的持久化数据返回 ValidationError。
func RestoreListing(record ListingRecord) (*Listing, error) {
	// === 校验持久化字段 ===
	if err := required("listing.id", record.ID); err != nil {
		return nil, err
	}
	if err := required("listing.owner_id", record.OwnerID); err != nil {
		return nil, err
	}
	if err := required("listing.resource_ref", record.ResourceRef); err != nil {
		return nil, err
	}
	if !record.Side.valid() {
		return nil, &ValidationError{Field: "listing.side", Reason: "must be buy or sell"}
	}
	price, err := NewAmount(record.PriceCents)
	if err != nil {
		return nil, &ValidationError{Field: "listing.price", Reason: "must be positive"}
	}
	if !record.Status.valid() {
		return nil, &ValidationError{Field: "listing.status", Reason: "must be open, reserved, matched, or withdrawn"}
	}
	if record.CreatedAt.IsZero() {
		return nil, &ValidationError{Field: "listing.created_at", Reason: "must not be zero"}
	}
	if record.UpdatedAt.IsZero() {
		return nil, &ValidationError{Field: "listing.updated_at", Reason: "must not be zero"}
	}
	if record.UpdatedAt.Before(record.CreatedAt) {
		return nil, &ValidationError{Field: "listing.updated_at", Reason: "must not be before the creation time"}
	}

	// === 恢复挂单状态 ===
	return &Listing{
		id: record.ID, ownerID: record.OwnerID, resourceRef: record.ResourceRef,
		side: record.Side, price: price, status: record.Status,
		createdAt: record.CreatedAt, updatedAt: record.UpdatedAt,
	}, nil
}

// Record 返回可交给宿主适配器持久化的 Listing 表示。
func (l Listing) Record() ListingRecord {
	return ListingRecord{
		ID: l.id, OwnerID: l.ownerID, ResourceRef: l.resourceRef,
		Side: l.side, PriceCents: l.price.Cents(), Status: l.status,
		CreatedAt: l.createdAt, UpdatedAt: l.updatedAt,
	}
}

// ID 返回挂单 ID。
func (l Listing) ID() string { return l.id }

// OwnerID 返回挂单所有者 ID。
func (l Listing) OwnerID() string { return l.ownerID }

// ResourceRef 返回宿主资源引用。
func (l Listing) ResourceRef() string { return l.resourceRef }

// Side 返回挂单方向。
func (l Listing) Side() Side { return l.side }

// Price 返回挂单价格。
func (l Listing) Price() Amount { return l.price }

// Status 返回挂单状态。
func (l Listing) Status() ListingStatus { return l.status }

// CreatedAt 返回挂单创建时间。
func (l Listing) CreatedAt() time.Time { return l.createdAt }

// UpdatedAt 返回挂单最近更新时间。
func (l Listing) UpdatedAt() time.Time { return l.updatedAt }

// Reserve 将开放挂单预留给待处理撮合请求。
func (l *Listing) Reserve(at time.Time) error {
	return l.transition(ListingStatusOpen, ListingStatusReserved, at)
}

// Release 将预留挂单重新开放。
func (l *Listing) Release(at time.Time) error {
	return l.transition(ListingStatusReserved, ListingStatusOpen, at)
}

// MarkMatched 将预留挂单标记为已成交。
func (l *Listing) MarkMatched(at time.Time) error {
	return l.transition(ListingStatusReserved, ListingStatusMatched, at)
}

// Withdraw 撤回开放挂单。
func (l *Listing) Withdraw(at time.Time) error {
	return l.transition(ListingStatusOpen, ListingStatusWithdrawn, at)
}

func (l *Listing) transition(from, to ListingStatus, at time.Time) error {
	// === 校验状态转换 ===
	if l == nil {
		return &ValidationError{Field: "listing", Reason: "must not be nil"}
	}
	if l.status != from {
		return &TransitionError{Entity: "listing", From: l.status.String(), To: to.String()}
	}
	if err := validateTimestamp("listing.updated_at", at, l.updatedAt); err != nil {
		return err
	}

	// === 更新挂单状态 ===
	l.status = to
	l.updatedAt = at
	return nil
}

func validateTimestamp(field string, at, previous time.Time) error {
	if at.IsZero() {
		return &ValidationError{Field: field, Reason: "must not be zero"}
	}
	if at.Before(previous) {
		return &ValidationError{Field: field, Reason: "must not be before the current update time"}
	}
	return nil
}
