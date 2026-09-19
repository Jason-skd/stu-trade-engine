package engine

import (
	"context"
	"fmt"
)

// ReadMatchRequest 按 ID 读取并恢复一个 MatchRequest。
func ReadMatchRequest(ctx context.Context, reader MatchRequestReader, id string) (*MatchRequest, error) {
	requests, err := ReadMatchRequests(ctx, reader, []string{id})
	if err != nil {
		return nil, err
	}
	return requests[0], nil
}

// ReadMatchRequests 按输入 ID 顺序批量读取并恢复 MatchRequest。
func ReadMatchRequests(ctx context.Context, reader MatchRequestReader, ids []string) ([]*MatchRequest, error) {
	// === 校验读取请求 ===
	if reader == nil {
		return nil, &ValidationError{Field: "match_request_reader", Reason: "must not be nil"}
	}
	if len(ids) == 0 {
		return []*MatchRequest{}, nil
	}
	if err := validateMatchRequestIDs(ids); err != nil {
		return nil, err
	}

	// === 读取持久化记录 ===
	records, err := reader.ReadMatchRequests(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("read match requests: %w", err)
	}

	return restoreMatchRequestRecords(ids, records)
}

func validateMatchRequestIDs(ids []string) error {
	// === 校验请求标识 ===
	wanted := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if err := required("match_request.id", id); err != nil {
			return err
		}
		if _, duplicate := wanted[id]; duplicate {
			return &ValidationError{Field: "match_request.ids", Reason: "must not contain duplicates"}
		}
		wanted[id] = struct{}{}
	}
	return nil
}

func restoreMatchRequestRecords(ids []string, records []MatchRequestRecord) ([]*MatchRequest, error) {
	// === 校验并恢复返回记录 ===
	wanted := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		wanted[id] = struct{}{}
	}
	byID := make(map[string]*MatchRequest, len(records))
	for _, record := range records {
		if _, expected := wanted[record.ID]; !expected {
			return nil, &ValidationError{Field: "match_request.records", Reason: "must contain only requested match requests"}
		}
		if _, duplicate := byID[record.ID]; duplicate {
			return nil, &ValidationError{Field: "match_request.records", Reason: "must not contain duplicates"}
		}
		request, err := RestoreMatchRequest(record)
		if err != nil {
			return nil, fmt.Errorf("restore match request %q: %w", record.ID, err)
		}
		byID[record.ID] = request
	}

	// === 按请求顺序返回实体 ===
	requests := make([]*MatchRequest, 0, len(ids))
	for _, id := range ids {
		request, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("read match request %q: %w", id, ErrMatchRequestNotFound)
		}
		requests = append(requests, request)
	}
	return requests, nil
}
