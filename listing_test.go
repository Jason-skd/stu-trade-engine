package engine

import (
	"errors"
	"testing"
	"time"
)

var testTime = time.Date(2026, time.September, 17, 12, 0, 0, 0, time.UTC)

func TestNewAmount(t *testing.T) {
	t.Parallel()

	// === 定义测试用例 ===
	tests := []struct {
		name    string
		cents   int64
		wantErr bool
	}{
		{name: "negative", cents: -1, wantErr: true},
		{name: "zero", cents: 0, wantErr: true},
		{name: "one cent", cents: 1},
		{name: "large amount", cents: 1<<63 - 1},
	}

	// === 执行测试用例 ===
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			amount, err := NewAmount(tt.cents)
			if tt.wantErr {
				assertValidationError(t, err, "amount")
				return
			}
			if err != nil {
				t.Fatalf("NewAmount() error = %v", err)
			}
			if got := amount.Cents(); got != tt.cents {
				t.Fatalf("Cents() = %d, want %d", got, tt.cents)
			}
		})
	}
}

func TestNewListing(t *testing.T) {
	t.Parallel()

	// === 定义测试用例 ===
	price := mustAmount(t, 2500)
	tests := []struct {
		name        string
		id          string
		ownerID     string
		resourceRef string
		side        Side
		price       Amount
		createdAt   time.Time
		wantField   string
	}{
		{name: "valid buy", id: "buy-1", ownerID: "user-1", resourceRef: "book-1", side: SideBuy, price: price, createdAt: testTime},
		{name: "valid sell", id: "sell-1", ownerID: "user-2", resourceRef: "book-2", side: SideSell, price: price, createdAt: testTime},
		{name: "missing id", ownerID: "user-1", resourceRef: "book-1", side: SideBuy, price: price, createdAt: testTime, wantField: "listing.id"},
		{name: "blank id", id: " \t", ownerID: "user-1", resourceRef: "book-1", side: SideBuy, price: price, createdAt: testTime, wantField: "listing.id"},
		{name: "missing owner", id: "buy-1", resourceRef: "book-1", side: SideBuy, price: price, createdAt: testTime, wantField: "listing.owner_id"},
		{name: "missing resource", id: "buy-1", ownerID: "user-1", side: SideBuy, price: price, createdAt: testTime, wantField: "listing.resource_ref"},
		{name: "invalid side", id: "buy-1", ownerID: "user-1", resourceRef: "book-1", side: Side(99), price: price, createdAt: testTime, wantField: "listing.side"},
		{name: "invalid price", id: "buy-1", ownerID: "user-1", resourceRef: "book-1", side: SideBuy, createdAt: testTime, wantField: "listing.price"},
		{name: "zero creation time", id: "buy-1", ownerID: "user-1", resourceRef: "book-1", side: SideBuy, price: price, wantField: "listing.created_at"},
	}

	// === 执行测试用例 ===
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			listing, err := NewListing(tt.id, tt.ownerID, tt.resourceRef, tt.side, tt.price, tt.createdAt)
			if tt.wantField != "" {
				assertValidationError(t, err, tt.wantField)
				if listing != nil {
					t.Fatal("NewListing() returned a listing after validation failure")
				}
				return
			}
			if err != nil {
				t.Fatalf("NewListing() error = %v", err)
			}
			if listing.ID() != tt.id || listing.OwnerID() != tt.ownerID || listing.ResourceRef() != tt.resourceRef {
				t.Fatalf("NewListing() identifiers = (%q, %q, %q)", listing.ID(), listing.OwnerID(), listing.ResourceRef())
			}
			if listing.Side() != tt.side || listing.Price() != tt.price {
				t.Fatalf("NewListing() trade values = (%v, %v)", listing.Side(), listing.Price())
			}
			if listing.Status() != ListingStatusOpen {
				t.Fatalf("Status() = %v, want %v", listing.Status(), ListingStatusOpen)
			}
			if listing.CreatedAt() != testTime || listing.UpdatedAt() != testTime {
				t.Fatalf("timestamps = (%v, %v), want %v", listing.CreatedAt(), listing.UpdatedAt(), testTime)
			}
		})
	}
}

func TestListingAllowedTransitions(t *testing.T) {
	t.Parallel()

	// === 定义测试用例 ===
	tests := []struct {
		name        string
		transitions []func(*Listing, time.Time) error
		wantStatus  ListingStatus
	}{
		{name: "reserve", transitions: []func(*Listing, time.Time) error{(*Listing).Reserve}, wantStatus: ListingStatusReserved},
		{name: "reserve and release", transitions: []func(*Listing, time.Time) error{(*Listing).Reserve, (*Listing).Release}, wantStatus: ListingStatusOpen},
		{name: "reserve and match", transitions: []func(*Listing, time.Time) error{(*Listing).Reserve, (*Listing).MarkMatched}, wantStatus: ListingStatusMatched},
		{name: "withdraw", transitions: []func(*Listing, time.Time) error{(*Listing).Withdraw}, wantStatus: ListingStatusWithdrawn},
	}

	// === 执行测试用例 ===
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			listing := mustListing(t)
			at := testTime
			for _, transition := range tt.transitions {
				at = at.Add(time.Minute)
				if err := transition(listing, at); err != nil {
					t.Fatalf("transition error = %v", err)
				}
			}
			if listing.Status() != tt.wantStatus {
				t.Fatalf("Status() = %v, want %v", listing.Status(), tt.wantStatus)
			}
			if listing.UpdatedAt() != at {
				t.Fatalf("UpdatedAt() = %v, want %v", listing.UpdatedAt(), at)
			}
		})
	}
}

func TestListingRejectedTransitionsDoNotMutate(t *testing.T) {
	t.Parallel()

	// === 定义测试用例 ===
	tests := []struct {
		name    string
		prepare func(*Listing)
		act     func(*Listing, time.Time) error
	}{
		{name: "open to matched", act: (*Listing).MarkMatched},
		{name: "open to open", act: (*Listing).Release},
		{name: "reserved to withdrawn", prepare: reserveListing, act: (*Listing).Withdraw},
		{name: "reserved to reserved", prepare: reserveListing, act: (*Listing).Reserve},
		{name: "matched is terminal", prepare: matchListing, act: (*Listing).Release},
		{name: "withdrawn is terminal", prepare: withdrawListing, act: (*Listing).Reserve},
	}

	// === 执行测试用例 ===
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			listing := mustListing(t)
			if tt.prepare != nil {
				tt.prepare(listing)
			}
			before := *listing
			err := tt.act(listing, before.updatedAt.Add(time.Minute))
			assertTransitionError(t, err, "listing")
			if *listing != before {
				t.Fatalf("listing changed after failed transition: got %+v, want %+v", *listing, before)
			}
		})
	}
}

func TestListingRejectsInvalidTransitionTimeWithoutMutation(t *testing.T) {
	t.Parallel()

	// === 定义测试用例 ===
	tests := []struct {
		name string
		at   time.Time
	}{
		{name: "zero", at: time.Time{}},
		{name: "before update", at: testTime.Add(-time.Nanosecond)},
	}

	// === 执行测试用例 ===
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			listing := mustListing(t)
			before := *listing
			err := listing.Reserve(tt.at)
			assertValidationError(t, err, "listing.updated_at")
			if *listing != before {
				t.Fatalf("listing changed after invalid time: got %+v, want %+v", *listing, before)
			}
		})
	}
}

func TestListingAllowsEqualTransitionTime(t *testing.T) {
	t.Parallel()

	listing := mustListing(t)
	if err := listing.Reserve(testTime); err != nil {
		t.Fatalf("Reserve() error = %v", err)
	}
	if listing.UpdatedAt() != testTime {
		t.Fatalf("UpdatedAt() = %v, want %v", listing.UpdatedAt(), testTime)
	}
}

func TestNilListingTransition(t *testing.T) {
	t.Parallel()

	var listing *Listing
	assertValidationError(t, listing.Reserve(testTime), "listing")
}

func mustAmount(t *testing.T, cents int64) Amount {
	t.Helper()
	amount, err := NewAmount(cents)
	if err != nil {
		t.Fatalf("NewAmount(%d) error = %v", cents, err)
	}
	return amount
}

func mustListing(t *testing.T) *Listing {
	t.Helper()
	listing, err := NewListing("listing-1", "user-1", "book-1", SideBuy, mustAmount(t, 2500), testTime)
	if err != nil {
		t.Fatalf("NewListing() error = %v", err)
	}
	return listing
}

func reserveListing(listing *Listing) {
	if err := listing.Reserve(testTime.Add(time.Minute)); err != nil {
		panic(err)
	}
}

func matchListing(listing *Listing) {
	reserveListing(listing)
	if err := listing.MarkMatched(testTime.Add(2 * time.Minute)); err != nil {
		panic(err)
	}
}

func withdrawListing(listing *Listing) {
	if err := listing.Withdraw(testTime.Add(time.Minute)); err != nil {
		panic(err)
	}
}

func assertValidationError(t *testing.T, err error, wantField string) {
	t.Helper()

	// === 校验错误标识 ===
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("error = %v, want errors.Is(_, ErrInvalidArgument)", err)
	}

	// === 校验错误详情 ===
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("error = %v, want *ValidationError", err)
	}
	if validationErr.Field != wantField {
		t.Fatalf("ValidationError.Field = %q, want %q", validationErr.Field, wantField)
	}
}

func assertTransitionError(t *testing.T, err error, wantEntity string) {
	t.Helper()

	// === 校验错误标识 ===
	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("error = %v, want errors.Is(_, ErrInvalidTransition)", err)
	}

	// === 校验错误详情 ===
	var transitionErr *TransitionError
	if !errors.As(err, &transitionErr) {
		t.Fatalf("error = %v, want *TransitionError", err)
	}
	if transitionErr.Entity != wantEntity {
		t.Fatalf("TransitionError.Entity = %q, want %q", transitionErr.Entity, wantEntity)
	}
}
