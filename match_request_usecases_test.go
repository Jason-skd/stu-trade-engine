package engine

import (
	"context"
	"errors"
	"slices"
	"testing"
)

func TestReadMatchRequestsRestoresInputOrder(t *testing.T) {
	t.Parallel()

	// === 读取逆序记录 ===
	reader := &recordingMatchRequestReader{records: []MatchRequestRecord{
		matchRequestRecord("request-2"),
		matchRequestRecord("request-1"),
	}}
	requests, err := ReadMatchRequests(context.Background(), reader, []string{"request-1", "request-2"})
	if err != nil {
		t.Fatalf("ReadMatchRequests() error = %v", err)
	}
	if len(requests) != 2 || requests[0].ID() != "request-1" || requests[1].ID() != "request-2" {
		t.Fatalf("ReadMatchRequests() ids = %v, want input order", matchRequestIDs(requests))
	}
	if !slices.Equal(reader.ids, []string{"request-1", "request-2"}) {
		t.Fatalf("ReadMatchRequests() dependency ids = %v", reader.ids)
	}
}

func TestReadMatchRequestsEmptyDoesNotCallDependency(t *testing.T) {
	t.Parallel()

	// === 读取空列表 ===
	reader := &recordingMatchRequestReader{}
	requests, err := ReadMatchRequests(context.Background(), reader, nil)
	if err != nil {
		t.Fatalf("ReadMatchRequests() error = %v", err)
	}
	if requests == nil || len(requests) != 0 || reader.called {
		t.Fatalf("ReadMatchRequests() = (%v, called %t), want non-nil empty result without read", requests, reader.called)
	}
}

func TestReadMatchRequestRestoresRecord(t *testing.T) {
	t.Parallel()

	// === 读取单个请求 ===
	reader := &recordingMatchRequestReader{records: []MatchRequestRecord{matchRequestRecord("request-1")}}
	request, err := ReadMatchRequest(context.Background(), reader, "request-1")
	if err != nil {
		t.Fatalf("ReadMatchRequest() error = %v", err)
	}
	if request.ID() != "request-1" || !slices.Equal(reader.ids, []string{"request-1"}) {
		t.Fatalf("ReadMatchRequest() = (%q, ids %v)", request.ID(), reader.ids)
	}
}

func TestReadMatchRequestsRejectsInvalidRequest(t *testing.T) {
	t.Parallel()

	// === 定义非法请求 ===
	tests := []struct {
		name      string
		reader    MatchRequestReader
		ids       []string
		wantField string
	}{
		{name: "nil reader", ids: []string{"request-1"}, wantField: "match_request_reader"},
		{name: "blank id", reader: &recordingMatchRequestReader{}, ids: []string{" "}, wantField: "match_request.id"},
		{name: "duplicate id", reader: &recordingMatchRequestReader{}, ids: []string{"request-1", "request-1"}, wantField: "match_request.ids"},
	}

	// === 验证请求校验 ===
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ReadMatchRequests(context.Background(), tt.reader, tt.ids)
			assertValidationError(t, err, tt.wantField)
		})
	}
}

func TestReadMatchRequestsRejectsInvalidResponse(t *testing.T) {
	t.Parallel()

	// === 定义非法响应 ===
	tests := []struct {
		name      string
		records   []MatchRequestRecord
		wantField string
		wantErr   error
	}{
		{name: "missing record", wantErr: ErrMatchRequestNotFound},
		{name: "unexpected record", records: []MatchRequestRecord{matchRequestRecord("request-2")}, wantField: "match_request.records"},
		{name: "extra record", records: []MatchRequestRecord{matchRequestRecord("request-1"), matchRequestRecord("request-2")}, wantField: "match_request.records"},
		{name: "duplicate record", records: []MatchRequestRecord{matchRequestRecord("request-1"), matchRequestRecord("request-1")}, wantField: "match_request.records"},
		{name: "corrupt record", records: []MatchRequestRecord{{ID: "request-1"}}, wantField: "match_request.trade_price"},
	}

	// === 验证响应校验 ===
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ReadMatchRequests(
				context.Background(),
				&recordingMatchRequestReader{records: tt.records},
				[]string{"request-1"},
			)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("ReadMatchRequests() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			assertValidationError(t, err, tt.wantField)
		})
	}
}

func TestReadMatchRequestPreservesDependencyFailure(t *testing.T) {
	t.Parallel()

	// === 读取依赖失败 ===
	wantErr := errors.New("storage unavailable")
	_, err := ReadMatchRequest(context.Background(), &recordingMatchRequestReader{err: wantErr}, "request-1")
	if !errors.Is(err, wantErr) {
		t.Fatalf("ReadMatchRequest() error = %v, want wrapped dependency error", err)
	}
}

type recordingMatchRequestReader struct {
	records []MatchRequestRecord
	ids     []string
	called  bool
	err     error
}

func (r *recordingMatchRequestReader) ReadMatchRequests(_ context.Context, ids []string) ([]MatchRequestRecord, error) {
	r.called = true
	r.ids = slices.Clone(ids)
	return r.records, r.err
}

func matchRequestRecord(id string) MatchRequestRecord {
	return MatchRequestRecord{
		ID: id, BuyerID: "buyer-1", SellerID: "seller-1", ResourceRef: "resource-1",
		SellListingID: "sell-1", TradePriceCents: 2000, Source: MatchSourceBuyerPurchase,
		BuyerConfirmedAt: testTime, SellerConfirmedAt: testTime, Status: MatchRequestStatusPending,
		CreatedAt: testTime, UpdatedAt: testTime,
	}
}

func matchRequestIDs(requests []*MatchRequest) []string {
	ids := make([]string, len(requests))
	for i, request := range requests {
		ids[i] = request.ID()
	}
	return ids
}
