package utils

type ClientError struct {
	Code    int
	Message string
	Err     error
}

func Error(code int, msg string, errs ...error) *ClientError {
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
