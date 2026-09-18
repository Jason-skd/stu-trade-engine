package engine

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ReadListing 按 ID 读取并恢复一个 Listing。
func ReadListing(ctx context.Context, reader ListingReader, id string) (*Listing, error) {
	listings, err := ReadListings(ctx, reader, []string{id})
	if err != nil {
		return nil, err
	}
	return listings[0], nil
}

// ReadListings 按输入 ID 顺序批量读取并恢复 Listing。
func ReadListings(ctx context.Context, reader ListingReader, ids []string) ([]*Listing, error) {
	// === 校验读取请求 ===
	if reader == nil {
		return nil, &ValidationError{Field: "listing_reader", Reason: "must not be nil"}
	}
	if len(ids) == 0 {
		return []*Listing{}, nil
	}

	wanted := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if err := required("listing.id", id); err != nil {
			return nil, err
		}
		if _, duplicate := wanted[id]; duplicate {
			return nil, &ValidationError{Field: "listing.ids", Reason: "must not contain duplicates"}
		}
		wanted[id] = struct{}{}
	}

	// === 读取持久化记录 ===
	records, err := reader.ReadListings(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("read listings: %w", err)
	}

	// === 恢复挂单实体 ===
	byID := make(map[string]*Listing, len(records))
	for _, record := range records {
		if _, expected := wanted[record.ID]; !expected {
			return nil, &ValidationError{Field: "listing.records", Reason: "must contain only requested listings"}
		}
		if _, duplicate := byID[record.ID]; duplicate {
			return nil, &ValidationError{Field: "listing.records", Reason: "must not contain duplicates"}
		}
		listing, err := RestoreListing(record)
		if err != nil {
			return nil, fmt.Errorf("restore listing %q: %w", record.ID, err)
		}
		byID[record.ID] = listing
	}

	// === 重建请求顺序 ===
	listings := make([]*Listing, 0, len(ids))
	for _, id := range ids {
		listing, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("read listing %q: %w", id, ErrListingNotFound)
		}
		listings = append(listings, listing)
	}
	return listings, nil
}

// WithdrawListingInput 是撤回挂单所需的宿主输入。
type WithdrawListingInput struct {
	ID        string
	OwnerID   string
	UpdatedAt time.Time
}

// WithdrawListing 在调用方提供的事务内锁定并撤回 open Listing。
// 已由同一所有者撤回的 Listing 幂等成功。
func WithdrawListing(ctx context.Context, withdrawer ListingWithdrawer, input WithdrawListingInput) (*Listing, error) {
	// === 校验撤回请求 ===
	if withdrawer == nil {
		return nil, &ValidationError{Field: "listing_withdrawer", Reason: "must not be nil"}
	}
	if err := required("listing.id", input.ID); err != nil {
		return nil, err
	}
	if err := required("listing.owner_id", input.OwnerID); err != nil {
		return nil, err
	}
	if input.UpdatedAt.IsZero() {
		return nil, &ValidationError{Field: "listing.updated_at", Reason: "must not be zero"}
	}

	// === 锁定并恢复挂单 ===
	record, err := withdrawer.ReadListingForUpdate(ctx, input.ID)
	if err != nil {
		if errors.Is(err, ErrListingNotFound) {
			return nil, fmt.Errorf("read listing %q for update: %w", input.ID, ErrListingNotFound)
		}
		return nil, fmt.Errorf("read listing %q for update: %w", input.ID, err)
	}
	listing, err := RestoreListing(record)
	if err != nil {
		return nil, fmt.Errorf("restore listing %q: %w", input.ID, err)
	}
	if listing.ID() != input.ID {
		return nil, &ValidationError{Field: "listing.id", Reason: "must match the requested id"}
	}

	// === 校验撤回权限 ===
	if listing.OwnerID() != input.OwnerID {
		return nil, fmt.Errorf("withdraw listing %q: %w", input.ID, ErrListingNotOwned)
	}
	if listing.Status() == ListingStatusWithdrawn {
		return listing, nil
	}

	// === 撤回并持久化挂单 ===
	if err := listing.Withdraw(input.UpdatedAt); err != nil {
		return nil, err
	}
	if err := withdrawer.UpdateListing(ctx, listing.Record()); err != nil {
		return nil, fmt.Errorf("update listing %q: %w", input.ID, err)
	}
	return listing, nil
}
