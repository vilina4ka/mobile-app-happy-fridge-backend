package domain

import "fmt"

// AppError provides a standard way to handle application errors
type AppError struct {
	OriginalError error
	Message       string
	HTTPStatus    int
}

func (e *AppError) Error() string {
	if e.OriginalError != nil {
		return fmt.Sprintf("Status %d: %s (Caused by: %v)", e.HTTPStatus, e.Message, e.OriginalError)
	}
	return fmt.Sprintf("Status %d: %s", e.HTTPStatus, e.Message)
}

// NewAppError creates a new application error
func NewAppError(err error, message string, status int) *AppError {
	return &AppError{
		OriginalError: err,
		Message:       message,
		HTTPStatus:    status,
	}
}

var (
	ErrNotFound = func(entity string, id interface{}) *AppError {
		return NewAppError(nil, fmt.Sprintf("%s with ID %v not found", entity, id), 404)
	}
	ErrUserNotFoundByLogin = func(login string) *AppError {
		return NewAppError(nil, fmt.Sprintf("User with login '%s' not found", login), 404)
	}
	ErrFridgeNotFoundForUser = func(userId int64) *AppError {
		return NewAppError(nil, fmt.Sprintf("Fridge not found for user ID %d", userId), 404)
	}
	ErrBadRequest = func(message string) *AppError {
		return NewAppError(nil, message, 400)
	}
	ErrValidationFailed = func(err error) *AppError {
		return NewAppError(err, fmt.Sprintf("Input validation failed: %v", err), 400)
	}
	ErrUnauthorized = func() *AppError {
		return NewAppError(nil, "Unauthorized", 401)
	}
	ErrForbidden = func() *AppError {
		return NewAppError(nil, "Forbidden", 403)
	}
	ErrInternalServer = func(err error) *AppError {
		return NewAppError(err, "Internal server error", 500)
	}
	ErrDatabase = func(err error) *AppError {
		return NewAppError(err, "Database operation failed", 500)
	}
	ErrDuplicateLogin = func(login string) *AppError {
		return NewAppError(nil, fmt.Sprintf("Login '%s' is already taken", login), 409)
	}
)
