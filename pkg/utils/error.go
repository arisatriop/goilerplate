package utils

import "goilerplate/pkg/constants"

var (
	ErrEmailNotRegistered         = Error(constants.StatusEmailNotRegistered, constants.MsgEmailNotRegistered)
	ErrInvalidCredentials         = Error(constants.StatusInvalidCredentials, constants.MsgInvalidCredentials)
	ErrAccountLocked              = Error(constants.StatusAccountLocked, constants.MsgAccountLocked)
	ErrAccountDisabled            = Error(constants.StatusAccountDisabled, constants.MsgAccountDisabled)
	ErrEmailNotVerified           = Error(constants.StatusEmailNotVerified, constants.MsgEmailNotVerified)
	ErrEmailAlreadyExists         = Error(constants.StatusEmailAlreadyExists, constants.MsgEmailAlreadyExists)
	ErrUserNotFound               = Error(constants.StatusUserNotFound, constants.MsgUserNotFound)
	ErrSessionNotFound            = Error(constants.StatusSessionNotFound, constants.MsgSessionNotFound)
	ErrSessionExpired             = Error(constants.StatusSessionExpired, constants.MsgSessionExpired)
	ErrSessionInvalid             = Error(constants.StatusSessionInvalid, constants.MsgSessionInvalid)
	ErrTokenNotFound              = Error(constants.StatusTokenNotFound, constants.MsgTokenNotFound)
	ErrTokenExpired               = Error(constants.StatusTokenExpired, constants.MsgTokenExpired)
	ErrTokenAlreadyUsed           = Error(constants.StatusTokenAlreadyUsed, constants.MsgTokenAlreadyUsed)
	ErrTokenInvalid               = Error(constants.StatusTokenInvalid, constants.MsgTokenInvalid)
	ErrPasswordTooWeak            = Error(constants.StatusPasswordTooWeak, constants.MsgPasswordTooWeak)
	ErrCurrentPasswordWrong       = Error(constants.StatusCurrentPasswordWrong, constants.MsgCurrentPasswordWrong)
	ErrUnauthorized               = Error(constants.StatusUnauthorized, constants.MsgUnauthorized)
	ErrPermissionDenied           = Error(constants.StatusPermissionDenied, constants.MsgPermissionDenied)
	ErrNotImplemented             = Error(constants.StatusNotImplemented, constants.MsgNotImplemented)
	ErrAuthorizationHeaderMissing = Error(constants.StatusAuthorizationHeaderMissing, constants.MsgAuthorizationHeaderMissing)
	ErrInvalidAuthorizationFormat = Error(constants.StatusInvalidAuthorizationFormat, constants.MsgInvalidAuthorizationFormat)
	ErrTokenEmpty                 = Error(constants.StatusTokenEmpty, constants.MsgTokenEmpty)
)

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
