package errs

import "fmt"

type ErrorCode string

const (
	ErrorCodeDuplicate    ErrorCode = "DUPLICATE_ENTRY"
	ErrorCodeNotFound     ErrorCode = "NOT_FOUND"
	ErrorCodeInternal     ErrorCode = "INTERNAL_ERROR"
	ErrorCodeValidation   ErrorCode = "VALIDATION_ERROR"
	ErrorCodeUnauthorized ErrorCode = "UNAUTHORIZED_ERROR"
	ErrorCodeBadRequest   ErrorCode = "BAD_REQUEST"
)

type AppError struct {
	Code ErrorCode `json:"code"`
	Msg  string    `json:"message"`
	Err  error     `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s]: %s", e.Code, e.Msg)
}

func NewAppErr(err error, msg string, code ErrorCode) error {
	if msg == "" && err != nil {
		msg = err.Error()
	}
	return &AppError{
		Code: code,
		Msg:  msg,
		Err:  err,
	}
}

func NewUnauthorizedErr(err error, msg string) error {
	return NewAppErr(err, msg, ErrorCodeUnauthorized)
}

func NewInternalServerErr(err error, msg string) error {
	return NewAppErr(err, msg, ErrorCodeInternal)
}

func NewNotFoundErr(err error, msg string) error {
	return NewAppErr(err, msg, ErrorCodeNotFound)
}

func NewBadRequestErr(err error, msg string) error {
	return NewAppErr(err, msg, ErrorCodeBadRequest)
}
