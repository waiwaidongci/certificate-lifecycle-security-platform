package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

type Code string

const (
	CodeInvalidArgument Code = "INVALID_ARGUMENT"
	CodeNotFound        Code = "NOT_FOUND"
	CodeConflict        Code = "CONFLICT"
	CodeUnavailable     Code = "UNAVAILABLE"
	CodeInternal        Code = "INTERNAL"
	CodeUnauthenticated Code = "UNAUTHENTICATED"
	CodeRateLimited     Code = "RATE_LIMITED"
)

type Error struct {
	Code    Code
	Message string
	Details map[string]any
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func New(code Code, message string) *Error {
	return &Error{Code: code, Message: message}
}

func Wrap(code Code, message string, err error) *Error {
	return &Error{Code: code, Message: message, Err: err}
}

func Invalid(message string) *Error {
	return New(CodeInvalidArgument, message)
}

func NotFound(message string) *Error {
	return New(CodeNotFound, message)
}

func Conflict(message string) *Error {
	return New(CodeConflict, message)
}

func Unavailable(message string) *Error {
	return New(CodeUnavailable, message)
}

func Internal(err error) *Error {
	return &Error{Code: CodeInternal, Message: "internal error", Err: err}
}

func As(err error) (*Error, bool) {
	var target *Error
	if errors.As(err, &target) {
		return target, true
	}
	return nil, false
}

func HTTPStatus(code Code) int {
	switch code {
	case CodeInvalidArgument:
		return http.StatusBadRequest
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	case CodeUnavailable:
		return http.StatusServiceUnavailable
	case CodeUnauthenticated:
		return http.StatusUnauthorized
	case CodeRateLimited:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
