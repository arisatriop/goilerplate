package utils

import (
	"fmt"
	"runtime"
)

type ClientError struct {
	Code    int
	Message string
	Err     error
}

func ClientErr(code int, msg string, errs ...error) *ClientError {
	var err error
	if len(errs) > 0 {
		err = errs[0]
	}
	return &ClientError{
		Code:    code,
		Message: msg,
		Err:     err,
	}
}

func (e *ClientError) Error() string {
	return e.Message
}

// InternalError wraps an error with file:line information for logging
type InternalError struct {
	Err  error
	File string
	Line int
	Msg  string
}

func (e *InternalError) Error() string {
	return e.Msg
}

func (e *InternalError) Unwrap() error {
	return e.Err
}

func (e *InternalError) Location() string {
	return fmt.Sprintf("%s:%d", e.File, e.Line)
}

// WrapErr wraps an error with file:line info captured at call site
func WrapErr(err error, msg string) error {
	if err == nil {
		return nil
	}
	_, file, line, _ := runtime.Caller(1)
	return &InternalError{
		Err:  err,
		File: file,
		Line: line,
		Msg:  fmt.Sprintf("%s: %v", msg, err),
	}
}
