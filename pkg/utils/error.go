package utils

import (
	"fmt"
	"runtime"
)

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
func WrapErr(err error, msg ...string) error {
	if err == nil {
		return nil
	}

	var fullMsg string
	fullMsg = err.Error()
	if len(msg) > 0 {
		fullMsg = fmt.Sprintf("%s: %v", msg[0], err)
	}

	_, file, line, _ := runtime.Caller(1)
	return &InternalError{
		Err:  err,
		File: file,
		Line: line,
		Msg:  fullMsg,
	}
}
