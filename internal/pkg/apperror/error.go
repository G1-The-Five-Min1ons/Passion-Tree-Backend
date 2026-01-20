package apperror

import (
	"fmt"
	"strings"

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

// Helper functions to check database error types
func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	// MSSQL duplicate key error messages
	return strings.Contains(errMsg, "duplicate key") ||
		strings.Contains(errMsg, "unique constraint") ||
		strings.Contains(errMsg, "violation of unique key") ||
		strings.Contains(errMsg, "cannot insert duplicate")
}

func IsForeignKeyError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	// MSSQL foreign key error messages
	return strings.Contains(errMsg, "foreign key constraint") ||
		strings.Contains(errMsg, "the delete statement conflicted with the reference constraint") ||
		strings.Contains(errMsg, "reference constraint")
}
