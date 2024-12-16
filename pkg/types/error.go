package types

import "fmt"

type ErrorCode string

const (
	ErrorCodeDuplicate  ErrorCode = "DUPLICATE_ENTRY"
	ErrorCodeNotFound   ErrorCode = "NOT_FOUND"
	ErrorCodeInternal   ErrorCode = "INTERNAL_ERROR"
	ErrorCodeValidation ErrorCode = "VALIDATION_ERROR"
)

type APIError struct {
	Code ErrorCode `json:"code"`
	Msg  string    `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Msg)
}
