package engine

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestReadListingsRestoresInputOrder(t *testing.T) {
	t.Parallel()

	// === 读取逆序记录 ===
	reader := &recordingListingReader{records: []ListingRecord{listingRecord("listing-2"), listingRecord("listing-1")}}
	listings, err := ReadListings(context.Background(), reader, []string{"listing-1", "listing-2"})
	if err != nil {
		t.Fatalf("ReadListings() error = %v", err)
	}
	if len(listings) != 2 || listings[0].ID() != "listing-1" || listings[1].ID() != "listing-2" {
		t.Fatalf("ReadListings() ids = %v, want input order", listingIDs(listings))
	}
	if len(reader.ids) != 2 || reader.ids[0] != "listing-1" || reader.ids[1] != "listing-2" {
		t.Fatalf("ReadListings() dependency ids = %v", reader.ids)
	}
}

func TestReadListingsEmptyDoesNotCallDependency(t *testing.T) {
	t.Parallel()

	// === 读取空列表 ===
	reader := &recordingListingReader{}
	listings, err := ReadListings(context.Background(), reader, nil)
	if err != nil {
		t.Fatalf("ReadListings() error = %v", err)
	}
	if listings == nil || len(listings) != 0 || reader.called {
		t.Fatalf("ReadListings() = (%v, called %t), want non-nil empty result without read", listings, reader.called)
	}
}

func TestReadListingsRejectsInvalidRequest(t *testing.T) {
	t.Parallel()

	// === 定义非法请求 ===
	tests := []struct {
		name      string
		reader    ListingReader
		ids       []string
		wantField string
	}{
		{name: "nil reader", ids: []string{"listing-1"}, wantField: "listing_reader"},
		{name: "blank id", reader: &recordingListingReader{}, ids: []string{" "}, wantField: "listing.id"},
		{name: "duplicate id", reader: &recordingListingReader{}, ids: []string{"listing-1", "listing-1"}, wantField: "listing.ids"},
	}

	// === 验证请求校验 ===
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ReadListings(context.Background(), tt.reader, tt.ids)
			assertValidationError(t, err, tt.wantField)
		})
	}
}

func TestReadListingsRejectsInvalidResponse(t *testing.T) {
	t.Parallel()

	// === 定义非法响应 ===
	tests := []struct {
		name      string
		records   []ListingRecord
		wantField string
		wantErr   error
	}{
		{name: "missing record", wantErr: ErrListingNotFound},
		{name: "unexpected record", records: []ListingRecord{listingRecord("listing-2")}, wantField: "listing.records"},
		{name: "duplicate record", records: []ListingRecord{listingRecord("listing-1"), listingRecord("listing-1")}, wantField: "listing.records"},
		{name: "corrupt record", records: []ListingRecord{{ID: "listing-1"}}, wantField: "listing.owner_id"},
	}

	// === 验证响应校验 ===
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ReadListings(context.Background(), &recordingListingReader{records: tt.records}, []string{"listing-1"})
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("ReadListings() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			assertValidationError(t, err, tt.wantField)
		})
	}
}

func TestReadListingPreservesDependencyFailure(t *testing.T) {
	t.Parallel()

	// === 读取依赖失败 ===
	wantErr := errors.New("storage unavailable")
	_, err := ReadListing(context.Background(), &recordingListingReader{err: wantErr}, "listing-1")
	if !errors.Is(err, wantErr) {
		t.Fatalf("ReadListing() error = %v, want wrapped dependency error", err)
	}
}

func TestWithdrawListingWithdrawsOpenListing(t *testing.T) {
	t.Parallel()

	// === 撤回开放挂单 ===
	withdrawer := &recordingListingWithdrawer{record: listingRecord("listing-1")}
	listing, err := WithdrawListing(context.Background(), withdrawer, withdrawInput())
	if err != nil {
		t.Fatalf("WithdrawListing() error = %v", err)
	}
	if listing.Status() != ListingStatusWithdrawn || listing.UpdatedAt() != testTime.Add(time.Minute) {
		t.Fatalf("WithdrawListing() = (%v, %v)", listing.Status(), listing.UpdatedAt())
	}
	if withdrawer.updated == nil || *withdrawer.updated != listing.Record() {
		t.Fatal("WithdrawListing() did not persist the withdrawn listing")
	}
}

func TestWithdrawListingIsIdempotentForOwner(t *testing.T) {
	t.Parallel()

	// === 准备已撤回挂单 ===
	record := listingRecord("listing-1")
	record.Status = ListingStatusWithdrawn
	record.UpdatedAt = testTime.Add(time.Minute)
	withdrawer := &recordingListingWithdrawer{record: record}

	// === 重复撤回挂单 ===
	listing, err := WithdrawListing(context.Background(), withdrawer, withdrawInput())
	if err != nil {
		t.Fatalf("WithdrawListing() error = %v", err)
	}
	if listing.Record() != record || withdrawer.updated != nil {
		t.Fatalf("WithdrawListing() changed an already withdrawn listing")
	}
}

func TestWithdrawListingRejectsUnavailableState(t *testing.T) {
	t.Parallel()

	// === 验证不可撤回状态 ===
	for _, status := range []ListingStatus{ListingStatusReserved, ListingStatusMatched} {
		status := status
		t.Run(status.String(), func(t *testing.T) {
			t.Parallel()
			record := listingRecord("listing-1")
			record.Status = status
			withdrawer := &recordingListingWithdrawer{record: record}
			_, err := WithdrawListing(context.Background(), withdrawer, withdrawInput())
			assertTransitionError(t, err, "listing")
			if withdrawer.updated != nil {
				t.Fatal("WithdrawListing() persisted a rejected transition")
			}
		})
	}
}

func TestWithdrawListingRejectsNonOwner(t *testing.T) {
	t.Parallel()

	// === 尝试越权撤回 ===
	withdrawer := &recordingListingWithdrawer{record: listingRecord("listing-1")}
	input := withdrawInput()
	input.OwnerID = "user-2"
	_, err := WithdrawListing(context.Background(), withdrawer, input)
	if !errors.Is(err, ErrListingNotOwned) || withdrawer.updated != nil {
		t.Fatalf("WithdrawListing() error = %v, updated = %v", err, withdrawer.updated)
	}
}

func TestWithdrawListingRejectsMismatchedLockedRecord(t *testing.T) {
	t.Parallel()

	// === 读取不匹配记录 ===
	withdrawer := &recordingListingWithdrawer{record: listingRecord("listing-2")}
	_, err := WithdrawListing(context.Background(), withdrawer, withdrawInput())
	assertValidationError(t, err, "listing.id")
	if withdrawer.updated != nil {
		t.Fatal("WithdrawListing() persisted a mismatched record")
	}
}

func TestWithdrawListingPreservesDependencyFailures(t *testing.T) {
	t.Parallel()

	// === 校验读取错误 ===
	readErr := errors.New("read failed")
	_, err := WithdrawListing(context.Background(), &recordingListingWithdrawer{readErr: readErr}, withdrawInput())
	if !errors.Is(err, readErr) {
		t.Fatalf("WithdrawListing() read error = %v", err)
	}

	// === 校验更新错误 ===
	updateErr := errors.New("update failed")
	withdrawer := &recordingListingWithdrawer{record: listingRecord("listing-1"), updateErr: updateErr}
	listing, err := WithdrawListing(context.Background(), withdrawer, withdrawInput())
	if !errors.Is(err, updateErr) || listing != nil {
		t.Fatalf("WithdrawListing() = (%v, %v), want nil and wrapped update error", listing, err)
	}
}

func TestWithdrawListingRejectsInvalidInputBeforeRead(t *testing.T) {
	t.Parallel()

	// === 定义非法输入 ===
	tests := []struct {
		name       string
		withdrawer ListingWithdrawer
		input      WithdrawListingInput
		wantField  string
	}{
		{name: "nil withdrawer", input: withdrawInput(), wantField: "listing_withdrawer"},
		{name: "missing id", withdrawer: &recordingListingWithdrawer{}, input: WithdrawListingInput{OwnerID: "user-1", UpdatedAt: testTime}, wantField: "listing.id"},
		{name: "missing owner", withdrawer: &recordingListingWithdrawer{}, input: WithdrawListingInput{ID: "listing-1", UpdatedAt: testTime}, wantField: "listing.owner_id"},
		{name: "missing time", withdrawer: &recordingListingWithdrawer{}, input: WithdrawListingInput{ID: "listing-1", OwnerID: "user-1"}, wantField: "listing.updated_at"},
	}

	// === 验证读取前校验 ===
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := WithdrawListing(context.Background(), tt.withdrawer, tt.input)
			assertValidationError(t, err, tt.wantField)
		})
	}
}

type recordingListingReader struct {
	records []ListingRecord
	ids     []string
	called  bool
	err     error
}

func (r *recordingListingReader) ReadListings(_ context.Context, ids []string) ([]ListingRecord, error) {
	r.called = true
	r.ids = append([]string(nil), ids...)
	return r.records, r.err
}

type recordingListingWithdrawer struct {
	record    ListingRecord
	updated   *ListingRecord
	readErr   error
	updateErr error
}

func (w *recordingListingWithdrawer) ReadListingForUpdate(_ context.Context, _ string) (ListingRecord, error) {
	return w.record, w.readErr
}

func (w *recordingListingWithdrawer) UpdateListing(_ context.Context, record ListingRecord) error {
	w.updated = &record
	return w.updateErr
}

func withdrawInput() WithdrawListingInput {
	return WithdrawListingInput{ID: "listing-1", OwnerID: "user-1", UpdatedAt: testTime.Add(time.Minute)}
}

func listingIDs(listings []*Listing) []string {
	ids := make([]string, len(listings))
	for i, listing := range listings {
		ids[i] = listing.ID()
	}
	return ids
}
