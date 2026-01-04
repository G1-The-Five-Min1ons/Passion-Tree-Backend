package apperror

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Log     error  `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Log
}

// 404
func NewNotFound(format string, args ...interface{}) *AppError {
	return &AppError{
		Code:    fiber.StatusNotFound,
		Message: fmt.Sprintf(format, args...),
	}
}

// 400
func NewBadRequest(format string, args ...interface{}) *AppError {
	return &AppError{
		Code:    fiber.StatusBadRequest,
		Message: fmt.Sprintf(format, args...),
	}
}

// 409
func NewConflict(format string, args ...interface{}) *AppError {
	return &AppError{
		Code:    fiber.StatusConflict,
		Message: fmt.Sprintf(format, args...),
	}
}

// 500
func NewInternal(err error) *AppError {
	return &AppError{
		Code:    fiber.StatusInternalServerError,
		Message: "internal server error",
		Log:     err,
	}
}

// 401
func NewUnauthorized(format string, args ...interface{}) *AppError {
	return &AppError{
		Code:    fiber.StatusUnauthorized,
		Message: fmt.Sprintf(format, args...),
	}
}

// 403
func NewForbidden(format string, args ...interface{}) *AppError {
	return &AppError{
		Code:    fiber.StatusForbidden,
		Message: fmt.Sprintf(format, args...),
	}
}