package errdefs

import (
	"encoding/json"
	"strings"
)

// maxMessageLength caps an extracted message. A server that answers an error
// with a whole HTML page should not turn that page into an error string.
const maxMessageLength = 2048

// BodyMessage pulls the human-readable part out of an error response body.
//
// The platforms in front of a sandbox each answer with their own JSON shape —
// {"code":401,"message":"..."} from the gateway, {"message":"..."} from envd,
// {"error":"..."} elsewhere — but all of them carry the explanation under one
// of a few well-known keys. When none of them is there, or the body is not
// JSON at all, the body is returned trimmed: whatever the server sent is still
// more useful than nothing.
func BodyMessage(body string) string {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return ""
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(trimmed), &fields); err != nil {
		return truncate(trimmed)
	}

	for _, key := range []string{"message", "error", "msg", "detail", "error_description"} {
		if message := stringField(fields[key]); message != "" {
			return truncate(message)
		}
	}
	return truncate(trimmed)
}

// stringField decodes a field that is expected to hold a string, and reports an
// empty string for anything else. {"error":{...}} is a nested object rather
// than a message, so it is left for the next candidate key.
func stringField(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return ""
	}
	return strings.TrimSpace(value)
}

// truncate bounds a message, marking where it was cut.
func truncate(s string) string {
	if len(s) <= maxMessageLength {
		return s
	}
	return s[:maxMessageLength] + "…"
}
