// Package shared provides common types and utilities for the GRC domain.
package shared

import (
	"fmt"
	"strings"
)

// Result represents either a success value or an error.
// Go doesn't have generics-based sum types like Rust's Result,
// so we use a struct with both fields and an ok flag.
type Result[T any] struct {
	value T
	err   error
	ok    bool
}

// Ok creates a successful Result.
func Ok[T any](value T) Result[T] {
	return Result[T]{value: value, ok: true}
}

// Err creates a failed Result.
func Err[T any](err error) Result[T] {
	return Result[T]{err: err, ok: false}
}

// IsOk returns true if the Result is successful.
func (r Result[T]) IsOk() bool {
	return r.ok
}

// IsErr returns true if the Result is an error.
func (r Result[T]) IsErr() bool {
	return !r.ok
}

// Unwrap returns the value if Ok, panics if Err.
func (r Result[T]) Unwrap() T {
	if !r.ok {
		panic(fmt.Sprintf("called Unwrap on an Err: %v", r.err))
	}
	return r.value
}

// UnwrapOr returns the value if Ok, or the default value if Err.
func (r Result[T]) UnwrapOr(defaultValue T) T {
	if r.ok {
		return r.value
	}
	return defaultValue
}

// Error returns the error if Err, nil if Ok.
func (r Result[T]) Error() error {
	if r.ok {
		return nil
	}
	return r.err
}

// Match applies the appropriate function based on the Result state.
func Match[T any, U any](r Result[T], onOk func(T) U, onErr func(error) U) U {
	if r.ok {
		return onOk(r.value)
	}
	return onErr(r.err)
}

// ValidationError represents a domain validation error.
type ValidationError struct {
	Field   string
	Message string
	Code    string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Field, e.Message)
}

// NewValidationError creates a new ValidationError.
func NewValidationError(field, message, code string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	}
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// Add appends a validation error to the collection.
func (e *ValidationErrors) Add(field, message, code string) {
	*e = append(*e, NewValidationError(field, message, code))
}

// HasErrors returns true if there are any validation errors.
func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

// ToError converts to error interface, returns nil if no errors.
func (e ValidationErrors) ToError() error {
	if len(e) == 0 {
		return nil
	}
	return e
}
