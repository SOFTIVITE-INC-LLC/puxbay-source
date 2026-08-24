package customerrors

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// AppError represents a safe error that can be sent to the client
type AppError struct {
	Code    int
	Message string
	Err     error // The original, underlying error for logging
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// WrapDBError inspects the database error and returns a generic AppError
// to prevent leaking schema details or stack traces to the client.
func WrapDBError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &AppError{Code: 404, Message: "Resource not found", Err: err}
	}

	// For Postgres unique constraint violations (e.g. 23505)
	if strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "SQLSTATE 23505") {
		return &AppError{Code: 409, Message: "Resource already exists with those details", Err: err}
	}

	// For foreign key violations
	if strings.Contains(err.Error(), "violates foreign key constraint") {
		return &AppError{Code: 400, Message: "Referenced resource does not exist", Err: err}
	}

	// Generic fallback for all other DB errors
	return &AppError{Code: 500, Message: "An internal database error occurred", Err: err}
}
