package engine

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrInvalidArgument 表示调用参数无法构成合法领域值。
	ErrInvalidArgument = errors.New("invalid argument")
	// ErrInvalidTransition 表示当前状态不允许目标状态转换。
	ErrInvalidTransition = errors.New("invalid state transition")
	// ErrConfirmationRequired 表示完成撮合请求前仍缺少参与方确认。
	ErrConfirmationRequired = errors.New("confirmation required")
	// ErrListingNotFound 表示请求的挂单不存在。
	ErrListingNotFound = errors.New("listing not found")
	// ErrListingNotOwned 表示调用者不是挂单所有者。
	ErrListingNotOwned = errors.New("listing not owned")
	// ErrListingUnavailable 表示目标挂单方向不符或已不可成交。
	ErrListingUnavailable = errors.New("listing unavailable")
	// ErrMatchRejected 表示宿主业务规则判定撮合不成立。
	ErrMatchRejected = errors.New("match rejected")
)

// ValidationError 描述无法通过领域校验的字段。
// Field 是稳定的字段标识，Reason 是供宿主记录的英文内部上下文。
type ValidationError struct {
	// === runtime 字段 ===
	Field  string
	Reason string
}

// Error 返回稳定的字段校验上下文。
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validate %s: %s", e.Field, e.Reason)
}

// Unwrap 允许调用方通过 errors.Is 判断 ErrInvalidArgument。
func (e *ValidationError) Unwrap() error {
	return ErrInvalidArgument
}

// TransitionError 描述被拒绝的状态转换。
// Entity、From 和 To 提供稳定的领域状态上下文。
type TransitionError struct {
	// === runtime 字段 ===
	Entity string
	From   string
	To     string
}

// Error 返回稳定的状态转换上下文。
func (e *TransitionError) Error() string {
	return fmt.Sprintf("transition %s from %s to %s: %v", e.Entity, e.From, e.To, ErrInvalidTransition)
}

// Unwrap 允许调用方通过 errors.Is 判断 ErrInvalidTransition。
func (e *TransitionError) Unwrap() error {
	return ErrInvalidTransition
}

func required(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{Field: field, Reason: "must not be empty"}
	}
	return nil
}
