package helper

type AppError struct {
	Code    int
	Message string
	Err     error
}

func Error(code int, msg string, errs ...error) *AppError {
	var err error
	if len(errs) > 0 {
		err = errs[0]
	}
	return &AppError{
		Code:    code,
		Message: msg,
		Err:     err,
	}
}

func (e *AppError) Error() string {
	return e.Message
}
