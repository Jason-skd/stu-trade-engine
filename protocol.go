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
