package customerrors

import "errors"

var (
	// Domain's
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidUsername = errors.New("invalid username")
	ErrInvalidPassword = errors.New("invalid password")

	// General's
	ErrNilArgs               = errors.New("some args are nil")
	ErrUserAlreadyRegistered = errors.New("user is already registered")
	ErrUserDoesntExist       = errors.New("user doesn't exist")
	ErrInvalidToken          = errors.New("invalid token")
	ErrNoSessions            = errors.New("there aren't any sessions")

	// Service's
	ErrNilRequest         = errors.New("request cannot be nil")
	ErrInternalServer     = errors.New("internal server error")
	ErrInvalidRequest     = errors.New("invalid request")
	ErrInvalidUserID      = errors.New("invalid user's id")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
