package engine

import (
	"context"
	"errors"
	"testing"
)

func TestPublishListingCreatesOpenListing(t *testing.T) {
	t.Parallel()

	creator := &recordingListingCreator{}
	input := PublishListingInput{
		ID:          "listing-1",
		OwnerID:     "user-1",
		ResourceRef: "book-1",
		Side:        SideSell,
		Price:       mustAmount(t, 2500),
		CreatedAt:   testTime,
	}

	listing, err := PublishListing(context.Background(), creator, input)
	if err != nil {
		t.Fatalf("PublishListing() error = %v", err)
	}
	if listing != creator.created {
		t.Fatal("PublishListing() did not return the persisted listing")
	}
	if listing.Status() != ListingStatusOpen {
		t.Fatalf("Status() = %v, want %v", listing.Status(), ListingStatusOpen)
	}
	if listing.ID() != input.ID || listing.OwnerID() != input.OwnerID || listing.ResourceRef() != input.ResourceRef {
		t.Fatalf("listing identifiers = (%q, %q, %q)", listing.ID(), listing.OwnerID(), listing.ResourceRef())
	}
}

func TestPublishListingRejectsInvalidInputBeforePersistence(t *testing.T) {
	t.Parallel()

	creator := &recordingListingCreator{}
	listing, err := PublishListing(context.Background(), creator, PublishListingInput{})
	assertValidationError(t, err, "listing.id")
	if listing != nil || creator.called {
		t.Fatalf("result = (%v, called %t), want no listing and no persistence", listing, creator.called)
	}
}

func TestPublishListingPreservesPersistenceFailure(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("storage unavailable")
	creator := &recordingListingCreator{err: wantErr}
	listing, err := PublishListing(context.Background(), creator, PublishListingInput{
		ID:          "listing-1",
		OwnerID:     "user-1",
		ResourceRef: "book-1",
		Side:        SideSell,
		Price:       mustAmount(t, 2500),
		CreatedAt:   testTime,
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("PublishListing() error = %v, want wrapped persistence error", err)
	}
	if listing != nil {
		t.Fatalf("PublishListing() listing = %v, want nil", listing)
	}
}

func TestPublishListingRejectsNilCreator(t *testing.T) {
	t.Parallel()

	listing, err := PublishListing(context.Background(), nil, PublishListingInput{})
	assertValidationError(t, err, "listing_creator")
	if listing != nil {
		t.Fatalf("PublishListing() listing = %v, want nil", listing)
	}
}

type recordingListingCreator struct {
	created *Listing
	called  bool
	err     error
}

func (c *recordingListingCreator) CreateListing(_ context.Context, listing *Listing) error {
	c.called = true
	c.created = listing
	return c.err
}
