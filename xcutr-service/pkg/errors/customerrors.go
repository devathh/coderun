package customerrors

import "errors"

var (
	// domain's
	ErrNoFiles         = errors.New("can't create a container without files.")
	ErrInvalidFilename = errors.New("invalid filename")
	ErrEmptyFile       = errors.New("can't create an empty file")
	ErrTooLargeFile    = errors.New("file is too large")

	// general's
	ErrNilArgs      = errors.New("some args are nil")
	ErrInvalidToken = errors.New("invalid token")

	// repository
	ErrNotFoundContainer = errors.New("container not found")

	// service's
	ErrTooLargeTimeout = errors.New("timeout is too large")
	ErrNoMain          = errors.New("main file doesn't exist")
	ErrInvalidLang     = errors.New("this language doesn't exist")
	ErrInternalServer  = errors.New("internal server error")
)
