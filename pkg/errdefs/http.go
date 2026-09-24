package errdefs

import "fmt"

// FromHTTP turns a non-2xx response into the matching error type. body is the
// raw response body, which BodyMessage reduces to its human-readable part: the
// control plane and the sandbox gateway both answer with a JSON object whose
// message field explains the failure, and burying that inside the raw JSON
// makes callers parse it themselves. A body in any other shape is passed
// through unchanged.
//
// Statuses with no specific type map to *SandboxError carrying the status code,
// so an unrecognised failure is still distinguishable from success.
func FromHTTP(statusCode int, body string) error {
	message := BodyMessage(body)

	switch statusCode {
	case 400:
		return &InvalidArgumentError{SandboxError{Message: message}}
	case 401:
		return &AuthenticationError{SandboxError{Message: message}}
	case 403:
		return &ForbiddenError{SandboxError{Message: message}}
	case 404:
		return &NotFoundError{SandboxError{Message: message}}
	case 409:
		return &ConflictError{SandboxError{Message: message}}
	case 429:
		return &RateLimitError{SandboxError{Message: message}}
	case 502:
		// The proxy answers 502 when no sandbox is listening behind it, which
		// in practice means the sandbox is not running rather than a gateway
		// fault. Reported as a timeout to match the platform's own wording.
		return &TimeoutError{SandboxError{Message: "sandbox is likely not running"}}
	case 507:
		return &NotEnoughSpaceError{SandboxError{Message: message}}
	default:
		return &SandboxError{Message: fmt.Sprintf("HTTP %d: %s", statusCode, message)}
	}
}
