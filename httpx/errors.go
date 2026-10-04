package httpx

import (
	"errors"
	"fmt"
	"net/http"
)

// Code is a stable machine-readable error code shared by HTTP and MCP.
type Code string

const (
	CodeInvalid      Code = "invalid"
	CodeUnauthorized Code = "unauthorized"
	CodeForbidden    Code = "forbidden"
	CodeNotFound     Code = "not_found"
	CodeConflict     Code = "conflict"
	CodeInternal     Code = "internal"
)

// Error is the single business error type for transport mapping.
type Error struct {
	Code    Code           `json:"code"`
	Message string         `json:"message"`
	Detail  map[string]any `json:"detail,omitempty"`
	wrapped error
}

func (e *Error) Error() string {
	if e.wrapped != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.wrapped)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.wrapped }

func newError(code Code, msg string, wrapped error) *Error {
	return &Error{Code: code, Message: msg, wrapped: wrapped}
}

func Invalid(msg string) *Error      { return newError(CodeInvalid, msg, nil) }
func Unauthorized(msg string) *Error { return newError(CodeUnauthorized, msg, nil) }
func Forbidden(msg string) *Error    { return newError(CodeForbidden, msg, nil) }
func NotFound(msg string) *Error     { return newError(CodeNotFound, msg, nil) }
func Conflict(msg string) *Error     { return newError(CodeConflict, msg, nil) }

func Internal(msg string, err error) *Error {
	return newError(CodeInternal, msg, err)
}

// Wrap annotates an existing error with a transport-stable code.
func Wrap(err error, code Code, msg string) *Error {
	return newError(code, msg, err)
}

// WithDetail attaches a structured detail field to a business error.
func WithDetail(err *Error, key string, val any) *Error {
	if err.Detail == nil {
		err.Detail = map[string]any{}
	}
	err.Detail[key] = val
	return err
}

// AsError extracts *Error from an error chain.
func AsError(err error) (*Error, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

// HTTPStatus maps err to an HTTP status code.
func HTTPStatus(err error) int {
	if e, ok := AsError(err); ok {
		return StatusOf(e.Code)
	}
	return http.StatusInternalServerError
}

// StatusOf maps a business code to HTTP status.
func StatusOf(c Code) int {
	switch c {
	case CodeInvalid:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
