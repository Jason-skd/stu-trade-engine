package engine

import "context"

// ListingCreator 是发布挂单所需的最小持久化能力。
type ListingCreator interface {
	CreateListing(context.Context, *Listing) error
}

// ListingReader 是按 ID 批量读取挂单持久化表示所需的最小能力。
type ListingReader interface {
	ReadListings(context.Context, []string) ([]ListingRecord, error)
}

// ListingWithdrawer 是在宿主事务内撤回挂单所需的最小锁定与保存能力。
type ListingWithdrawer interface {
	ReadListingForUpdate(context.Context, string) (ListingRecord, error)
	UpdateListing(context.Context, ListingRecord) error
}

// MatchProposal 是宿主规则判断一次撮合所需的通用事实。
type MatchProposal struct {
	BuyerID       string
	SellerID      string
	ResourceRef   string
	BuyListingID  string
	SellListingID string
	TradePrice    Amount
	Source        MatchSource
}

// TradeInput 是宿主创建唯一订单所需的通用成交事实。
type TradeInput struct {
	MatchRequestID string
	BuyerID        string
	SellerID       string
	ResourceRef    string
	TradePrice     Amount
	Source         MatchSource
}

// MatchPolicy 是事务内授权宿主撮合规则所需的最小能力。
type MatchPolicy interface {
	EvaluateMatch(context.Context, MatchProposal) error
}

// TradeCreator 是事务内以 MatchRequest ID 创建宿主订单所需的最小能力。
// 对同一 MatchRequest ID 的重复调用不得创建第二张订单，并返回同一订单引用。
type TradeCreator interface {
	CreateTrade(context.Context, TradeInput) (string, error)
}

// DirectTradeTransaction 是一次直接成交在同一宿主事务内消费的完整能力。
type DirectTradeTransaction interface {
	MatchPolicy
	TradeCreator
	ReadListingForUpdate(context.Context, string) (ListingRecord, error)
	UpdateListing(context.Context, ListingRecord) error
	CreateMatchRequest(context.Context, MatchRequestRecord) error
	UpdateMatchRequest(context.Context, MatchRequestRecord) error
}
