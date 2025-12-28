package customerrors

import "errors"

var (
	ErrNilArgs = errors.New("some args are nil")

	ErrInternalServer = errors.New("internal server error")

	ErrUserNotFound = errors.New("user not found")
)
