package engine

import (
	"context"
	"fmt"
	"time"
)

// PublishListingInput 是发布挂单所需的宿主输入。
type PublishListingInput struct {
	ID          string
	OwnerID     string
	ResourceRef string
	Side        Side
	Price       Amount
	CreatedAt   time.Time
}

// PublishListing 校验并持久化一个初始状态为 open 的挂单。
func PublishListing(ctx context.Context, creator ListingCreator, input PublishListingInput) (*Listing, error) {
	// === 构造挂单 ===
	if creator == nil {
		return nil, &ValidationError{Field: "listing_creator", Reason: "must not be nil"}
	}
	listing, err := NewListing(
		input.ID,
		input.OwnerID,
		input.ResourceRef,
		input.Side,
		input.Price,
		input.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// === 持久化挂单 ===
	if err := creator.CreateListing(ctx, listing); err != nil {
		return nil, fmt.Errorf("create listing: %w", err)
	}
	return listing, nil
}
