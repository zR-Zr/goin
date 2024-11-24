package zerrors 

import (
	"errors"
	"fmt"
)

type ErrorCode interface {
	Code() int
	Message() string
	HTTPStatusCode() int
}

type ZError struct {
	Code     ErrorCode // 错误码
	StackErr error     // 原始错误
}

func (e *ZError) Error() string {
	return e.StackErr.Error()
}

func NewError(code ErrorCode, e error) *ZError {
	return &ZError{
		Code:     code,
		StackErr: e,
	}
}

func (e *ZError) HTTPStatusCode() int {
	return e.Code.HTTPStatusCode()
}

func (e *ZError) Stack() string {
	return fmt.Sprintf("%+v", e.StackErr)
}

// 校验错误
type ValidationError struct {
	*ZError
	Fields string
}

func NewValidationError(code ErrorCode, fields string, e error) error {
	return &ValidationError{
		ZError: NewError(code, e),
	}
}

type DatabaseError struct {
	*ZError
	SQL string
}

func NewDatabaseError(code ErrorCode, sql string, e error) error {
	return &DatabaseError{
		ZError: NewError(code, e),
		SQL:    sql,
	}
}

func IsCustomError(err error) bool {
	_, ok := err.(*ZError)
	return ok
}

func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

func IsDatabaseError(err error) bool {
	_, ok := err.(*DatabaseError)
	return ok
}

func AsZError(err error) *ZError {
	var ze *ZError
	if errors.As(err, &ze) {
		return ze
	}

	return nil
}

func AsValidationError(err error) *ValidationError {
	var ve *ValidationError
	if errors.As(err, &ve) {
		return ve
	}
	return nil
}

func AsDatabaseError(err error) *DatabaseError {
	var de *DatabaseError
	if errors.As(err, &de) {
		return de
	}
	return nil
}
