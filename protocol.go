package engine

import "context"

// ListingCreator 是发布挂单所需的最小持久化能力。
type ListingCreator interface {
	CreateListing(context.Context, *Listing) error
}
