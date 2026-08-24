package customerrors

import "errors"

var (
	// ErrNotFound indicates that the requested resource could not be found.
	ErrNotFound = errors.New("resource not found")

	// ErrDuplicate indicates that a resource with the same unique identifiers already exists.
	ErrDuplicate = errors.New("resource already exists")

	// ErrInvalidInput indicates that the provided input is invalid or malformed.
	ErrInvalidInput = errors.New("invalid input provided")

	// ErrUnauthorized indicates that the user is not authenticated or lacks required permissions.
	ErrUnauthorized = errors.New("unauthorized access")
)
