// Package errors provides error handling primitives for the multigit application.
package errors

import (
	"errors"
	"fmt"
)

// ErrorType represents the type of error.
type ErrorType string

// List of error types
const (
	// ErrorTypeConfig represents configuration related errors
	ErrorTypeConfig ErrorType = "config"
	// ErrorTypeSSH represents SSH related errors
	ErrorTypeSSH ErrorType = "ssh"
	// ErrorTypeGit represents Git related errors
	ErrorTypeGit ErrorType = "git"
	// ErrorTypeIO represents I/O related errors
	ErrorTypeIO ErrorType = "io"
	// ErrorTypeValidation represents validation errors
	ErrorTypeValidation ErrorType = "validation"
)

// Error represents an application error
type Error struct {
	// The error type
	Type ErrorType
	// The underlying error that triggered this one, if any
	Err error
	// Contextual information about the error
	Context map[string]interface{}
}

// Error returns the error message
func (e *Error) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return "unknown error"
}

// Unwrap returns the underlying error
func (e *Error) Unwrap() error {
	return e.Err
}

// WithContext adds context to the error
func (e *Error) WithContext(key string, value interface{}) *Error {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// New creates a new error with the given type and message
func New(errType ErrorType, message string) *Error {
	return &Error{
		Type:    errType,
		Err:     errors.New(message),
		Context: make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with a new error type
func Wrap(err error, errType ErrorType, message string) *Error {
	return &Error{
		Type:    errType,
		Err:     fmt.Errorf("%s: %w", message, err),
		Context: make(map[string]interface{}),
	}
}

// Errorf creates a new formatted error with the given type
func Errorf(errType ErrorType, format string, args ...interface{}) *Error {
	return &Error{
		Type:    errType,
		Err:     fmt.Errorf(format, args...),
		Context: make(map[string]interface{}),
	}
}

// IsType checks if the error is of the given type
func IsType(err error, errType ErrorType) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Type == errType
	}
	return false
}

// GetContext returns the context of the error if it's of type *Error
func GetContext(err error) map[string]interface{} {
	var e *Error
	if errors.As(err, &e) {
		return e.Context
	}
	return nil
}
