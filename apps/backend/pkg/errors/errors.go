package errors

import "fmt"

const (
	ErrNotFound    = 404
	ErrBadRequest  = 400
	ErrUnauthorized = 401
	ErrForbidden   = 403
	ErrInternal    = 500
	ErrRateLimit   = 429
	ErrTimeout     = 408
	ErrConflict    = 409
)

// AppError is the application-level error type.
type AppError struct {
	Code     int
	Message  string
	Internal error
}

func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Internal)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Internal
}

// IsAppError returns true if err is an *AppError.
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

func NewNotFoundError(msg string, internal error) *AppError {
	return &AppError{Code: ErrNotFound, Message: msg, Internal: internal}
}

func NewBadRequestError(msg string, internal error) *AppError {
	return &AppError{Code: ErrBadRequest, Message: msg, Internal: internal}
}

func NewUnauthorizedError(msg string, internal error) *AppError {
	return &AppError{Code: ErrUnauthorized, Message: msg, Internal: internal}
}

func NewInternalError(msg string, internal error) *AppError {
	return &AppError{Code: ErrInternal, Message: msg, Internal: internal}
}

func NewRateLimitError(msg string, internal error) *AppError {
	return &AppError{Code: ErrRateLimit, Message: msg, Internal: internal}
}

func NewConflictError(msg string, internal error) *AppError {
	return &AppError{Code: ErrConflict, Message: msg, Internal: internal}
}
