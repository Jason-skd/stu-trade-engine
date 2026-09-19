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

// MatchRequestReader 是按 ID 批量读取撮合请求持久化表示所需的最小能力。
type MatchRequestReader interface {
	ReadMatchRequests(context.Context, []string) ([]MatchRequestRecord, error)
}

// ListingLocker 要求宿主按给定 ID 顺序取得真实挂单锁。
// 返回记录必须与请求 ID 数量及集合完全一致，且不得包含重复记录。
type ListingLocker interface {
	ReadListingsForUpdate(context.Context, []string) ([]ListingRecord, error)
}

// ListingWithdrawer 是在宿主事务内撤回挂单所需的最小锁定与保存能力。
type ListingWithdrawer interface {
	ListingLocker
	UpdateListing(context.Context, ListingRecord) error
}

// MatchParticipant 是宿主校验一方参与者、相关挂单与资源所需的通用事实。
// 不适用于该撮合路径的 ListingID 或 ResourceRef 保持为空。
type MatchParticipant struct {
	// === runtime 字段 ===
	ParticipantID string
	ListingID     string
	ResourceRef   string
}

// MatchProposal 是宿主规则判断一次撮合所需的完整通用事实。
type MatchProposal struct {
	// === runtime 字段 ===
	Buyer      MatchParticipant
	Seller     MatchParticipant
	TradePrice Amount
	Source     MatchSource
}

// TradeInput 是宿主创建唯一订单所需的通用成交事实。
type TradeInput struct {
	// === runtime 字段 ===
	MatchRequestID string
	BuyerID        string
	SellerID       string
	ResourceRef    string
	TradePrice     Amount
	Source         MatchSource
}

// MatchValidator 区分撮合通过、正常业务不匹配和校验执行失败。
// true、nil 表示通过，false、nil 表示不匹配；error 非 nil 时忽略 bool。
type MatchValidator interface {
	ValidateMatch(context.Context, MatchProposal) (bool, error)
}

// MatchValidatorFunc 将函数适配为 MatchValidator。
type MatchValidatorFunc func(context.Context, MatchProposal) (bool, error)

// ValidateMatch 调用撮合校验函数。nil 函数返回 ValidationError。
func (f MatchValidatorFunc) ValidateMatch(ctx context.Context, proposal MatchProposal) (bool, error) {
	if f == nil {
		return false, &ValidationError{Field: "match_validator", Reason: "must not be nil"}
	}
	return f(ctx, proposal)
}

// TradeCreator 是事务内以 MatchRequest ID 创建宿主订单所需的最小能力。
// 对同一 MatchRequest ID 的重复调用不得创建第二张订单，并返回同一订单引用。
type TradeCreator interface {
	CreateTrade(context.Context, TradeInput) (string, error)
}

// DirectTradeUnitOfWork 由宿主开启、提交或回滚一次直接成交事务。
// WithTx 仅在回调成功且物理事务提交成功后返回 nil。
type DirectTradeUnitOfWork interface {
	WithTx(context.Context, func(DirectTradeTransaction) error) error
}

// DirectTradeTransaction 是 engine 在直接成交事务回调内消费的完整能力。
type DirectTradeTransaction interface {
	ListingLocker
	MatchValidator
	TradeCreator
	UpdateListing(context.Context, ListingRecord) error
	CreateMatchRequest(context.Context, MatchRequestRecord) error
	UpdateMatchRequest(context.Context, MatchRequestRecord) error
}
