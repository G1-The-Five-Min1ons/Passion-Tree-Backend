package apperror

import (
    "fmt"
    "github.com/gofiber/fiber/v2"
)

type AppError struct {
    Code    int    json:"code"
    Message string json:"message"
    Log     error  json:"-"
}

func (e AppError) Error() string {
    return e.Message
}

func (eAppError) Unwrap() error {
    return e.Log
}

// 404
func NewNotFound(format string, args ...interface{}) AppError {
    return &AppError{
        Code:    fiber.StatusNotFound,
        Message: fmt.Sprintf(format, args...),
    }
}

// 400
func NewBadRequest(format string, args ...interface{})AppError {
    return &AppError{
        Code:    fiber.StatusBadRequest,
        Message: fmt.Sprintf(format, args...),
    }
}

// 409
func NewConflict(format string, args ...interface{}) AppError {
    return &AppError{
        Code:    fiber.StatusConflict,
        Message: fmt.Sprintf(format, args...),
    }
}

// 500
func NewInternal(err error)AppError {
    return &AppError{
        Code:    fiber.StatusInternalServerError,
        Message: "internal server error",
        Log:     err,
    }
}

// 401
func NewUnauthorized(format string, args ...interface{}) AppError {
    return &AppError{
        Code:    fiber.StatusUnauthorized,
        Message: fmt.Sprintf(format, args...),
    }
}

// 403
func NewForbidden(format string, args ...interface{}) AppError {
    return &AppError{
        Code:    fiber.StatusForbidden,
        Message: fmt.Sprintf(format, args...),
    }
}

// HandleError sends error response with appropriate status code
func HandleError(c *fiber.Ctx, err error) error {
    if appErr, ok := err.(AppError); ok {
        return c.Status(appErr.Code).JSON(fiber.Map{
            "error":   appErr.Message,
            "message": appErr.Message,
        })
    }
    
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
        "error":   "internal_server_error",
        "message": "An unexpected error occurred",
    })
}