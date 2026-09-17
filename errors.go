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
)

// ValidationError 描述无法通过领域校验的字段。
type ValidationError struct {
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
type TransitionError struct {
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
