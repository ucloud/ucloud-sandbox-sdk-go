package sandbox

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

type SandboxError struct {
	Message string
	Cause   error
}

func (e *SandboxError) Error() string { return e.Message }
func (e *SandboxError) Unwrap() error { return e.Cause }

var (
	ErrTimeout   = &TimeoutError{SandboxError: SandboxError{Message: "timeout"}}
	ErrNotFound  = &NotFoundError{SandboxError: SandboxError{Message: "not found"}}
	ErrAuth      = &AuthenticationError{SandboxError: SandboxError{Message: "authentication failed"}}
	ErrRateLimit = &RateLimitError{SandboxError: SandboxError{Message: "rate limit exceeded"}}
)

type TimeoutError struct{ SandboxError }

func (e *TimeoutError) Is(target error) bool { _, ok := target.(*TimeoutError); return ok }

type NotFoundError struct{ SandboxError }

func (e *NotFoundError) Is(target error) bool { _, ok := target.(*NotFoundError); return ok }

type AuthenticationError struct{ SandboxError }

func (e *AuthenticationError) Is(target error) bool {
	_, ok := target.(*AuthenticationError)
	return ok
}

type InvalidArgumentError struct{ SandboxError }

func (e *InvalidArgumentError) Is(target error) bool {
	_, ok := target.(*InvalidArgumentError)
	return ok
}

type NotEnoughSpaceError struct{ SandboxError }

func (e *NotEnoughSpaceError) Is(target error) bool {
	_, ok := target.(*NotEnoughSpaceError)
	return ok
}

type RateLimitError struct{ SandboxError }

func (e *RateLimitError) Is(target error) bool { _, ok := target.(*RateLimitError); return ok }

type ConflictError struct{ SandboxError }

func (e *ConflictError) Is(target error) bool { _, ok := target.(*ConflictError); return ok }

type CommandExitError struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Message  string
	Cause    error
}

func (e *CommandExitError) Error() string {
	return fmt.Sprintf("command exited with code %d: %s", e.ExitCode, e.Message)
}
func (e *CommandExitError) Unwrap() error { return e.Cause }

type BuildError struct {
	SandboxError
	BuildID    string
	TemplateID string
}

func (e *BuildError) Is(target error) bool { _, ok := target.(*BuildError); return ok }

type FileUploadError struct{ SandboxError }

func (e *FileUploadError) Is(target error) bool { _, ok := target.(*FileUploadError); return ok }

type TemplateError struct{ SandboxError }

func (e *TemplateError) Is(target error) bool { _, ok := target.(*TemplateError); return ok }

type ForbiddenError struct{ SandboxError }

func (e *ForbiddenError) Is(target error) bool { _, ok := target.(*ForbiddenError); return ok }

func isConflictError(err error) bool {
	var conflictErr *ConflictError
	return errors.As(err, &conflictErr)
}

func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	if ne, ok := err.(net.Error); ok && ne.Timeout() {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline exceeded")
}

func mapHTTPError(statusCode int, body string) error {
	switch statusCode {
	case 400:
		return &InvalidArgumentError{SandboxError{Message: body}}
	case 401:
		return &AuthenticationError{SandboxError{Message: body}}
	case 403:
		return &ForbiddenError{SandboxError{Message: body}}
	case 404:
		return &NotFoundError{SandboxError{Message: body}}
	case 409:
		return &ConflictError{SandboxError{Message: body}}
	case 429:
		return &RateLimitError{SandboxError{Message: body}}
	case 502:
		return &TimeoutError{SandboxError{Message: "sandbox is likely not running"}}
	case 507:
		return &NotEnoughSpaceError{SandboxError{Message: body}}
	default:
		return &SandboxError{Message: fmt.Sprintf("HTTP %d: %s", statusCode, body)}
	}
}
