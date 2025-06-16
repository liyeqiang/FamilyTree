package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode 错误码类型
type ErrorCode string

const (
	// 通用错误码
	ErrCodeInvalidInput  ErrorCode = "INVALID_INPUT"
	ErrCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrCodeAlreadyExists ErrorCode = "ALREADY_EXISTS"
	ErrCodeInternalError ErrorCode = "INTERNAL_ERROR"
	ErrCodeUnauthorized  ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden     ErrorCode = "FORBIDDEN"

	// 业务错误码
	ErrCodeInvalidRelation  ErrorCode = "INVALID_RELATION"
	ErrCodeCircularRelation ErrorCode = "CIRCULAR_RELATION"
	ErrCodeGenderMismatch   ErrorCode = "GENDER_MISMATCH"
	ErrCodeHasChildren      ErrorCode = "HAS_CHILDREN"
	ErrCodeInFamily         ErrorCode = "IN_FAMILY"
)

// AppError 应用错误结构
type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// HTTPStatus 返回对应的HTTP状态码
func (e *AppError) HTTPStatus() int {
	switch e.Code {
	case ErrCodeInvalidInput, ErrCodeInvalidRelation, ErrCodeCircularRelation, ErrCodeGenderMismatch:
		return http.StatusBadRequest
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeAlreadyExists:
		return http.StatusConflict
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeHasChildren, ErrCodeInFamily:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// New 创建新的应用错误
func New(code ErrorCode, message string, details ...string) *AppError {
	err := &AppError{
		Code:    code,
		Message: message,
	}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}

// Wrap 包装现有错误
func Wrap(err error, code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: err.Error(),
	}
}

// 预定义的常用错误
var (
	ErrInvalidID        = New(ErrCodeInvalidInput, "无效的ID")
	ErrNotFound         = New(ErrCodeNotFound, "资源不存在")
	ErrEmptyName        = New(ErrCodeInvalidInput, "姓名不能为空")
	ErrSelfParent       = New(ErrCodeInvalidRelation, "不能将自己设为父母")
	ErrSameParents      = New(ErrCodeInvalidRelation, "父亲和母亲不能是同一个人")
	ErrCircularRelation = New(ErrCodeCircularRelation, "检测到循环关系")
	ErrGenderMismatch   = New(ErrCodeGenderMismatch, "性别不匹配")
	ErrHasChildren      = New(ErrCodeHasChildren, "该个人有子女记录，不能删除")
	ErrInFamily         = New(ErrCodeInFamily, "该个人仍存在于家庭关系中，不能删除")
)
