package engine

import (
	"testing"
	"time"
)

func TestRestoreListingRoundTrip(t *testing.T) {
	t.Parallel()

	// === 定义挂单状态 ===
	statuses := []ListingStatus{
		ListingStatusOpen,
		ListingStatusReserved,
		ListingStatusMatched,
		ListingStatusWithdrawn,
	}

	// === 验证持久化往返 ===
	for _, status := range statuses {
		status := status
		t.Run(status.String(), func(t *testing.T) {
			t.Parallel()

			record := listingRecord("listing-1")
			record.Status = status
			record.UpdatedAt = testTime.Add(time.Minute)
			listing, err := RestoreListing(record)
			if err != nil {
				t.Fatalf("RestoreListing() error = %v", err)
			}
			if got := listing.Record(); got != record {
				t.Fatalf("Record() = %+v, want %+v", got, record)
			}
		})
	}
}

func TestRestoreListingRejectsInvalidRecord(t *testing.T) {
	t.Parallel()

	// === 定义非法记录 ===
	tests := []struct {
		name      string
		change    func(*ListingRecord)
		wantField string
	}{
		{name: "missing id", change: func(r *ListingRecord) { r.ID = "" }, wantField: "listing.id"},
		{name: "missing owner", change: func(r *ListingRecord) { r.OwnerID = "" }, wantField: "listing.owner_id"},
		{name: "missing resource", change: func(r *ListingRecord) { r.ResourceRef = "" }, wantField: "listing.resource_ref"},
		{name: "invalid side", change: func(r *ListingRecord) { r.Side = Side(99) }, wantField: "listing.side"},
		{name: "invalid price", change: func(r *ListingRecord) { r.PriceCents = 0 }, wantField: "listing.price"},
		{name: "invalid status", change: func(r *ListingRecord) { r.Status = ListingStatus(99) }, wantField: "listing.status"},
		{name: "zero created at", change: func(r *ListingRecord) { r.CreatedAt = time.Time{} }, wantField: "listing.created_at"},
		{name: "zero updated at", change: func(r *ListingRecord) { r.UpdatedAt = time.Time{} }, wantField: "listing.updated_at"},
		{name: "updated before created", change: func(r *ListingRecord) { r.UpdatedAt = r.CreatedAt.Add(-time.Nanosecond) }, wantField: "listing.updated_at"},
	}

	// === 验证记录校验 ===
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			record := listingRecord("listing-1")
			tt.change(&record)
			listing, err := RestoreListing(record)
			assertValidationError(t, err, tt.wantField)
			if listing != nil {
				t.Fatalf("RestoreListing() = %v, want nil", listing)
			}
		})
	}
}

func listingRecord(id string) ListingRecord {
	return ListingRecord{
		ID: id, OwnerID: "user-1", ResourceRef: "book-1", Side: SideSell,
		PriceCents: 2500, Status: ListingStatusOpen,
		CreatedAt: testTime, UpdatedAt: testTime,
	}
}
