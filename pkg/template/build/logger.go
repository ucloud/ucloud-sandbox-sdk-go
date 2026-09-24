package build

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
)

// DefaultLogger returns a log callback that prints entries of level info and
// above to stderr.
func DefaultLogger() func(api.BuildLogEntry) {
	return DefaultLoggerWithLevel(api.LogLevelInfo)
}

// DefaultLoggerWithLevel returns a log callback that prints entries at or above
// minLevel to stderr, prefixed with the elapsed time.
func DefaultLoggerWithLevel(minLevel api.LogLevel) func(api.BuildLogEntry) {
	start := time.Now()
	minOrd := levelOrder(minLevel)

	return func(entry api.BuildLogEntry) {
		if levelOrder(entry.Level) < minOrd {
			return
		}
		// A build log entry can arrive without a timestamp; showing the
		// current time is more use than showing the zero time.
		stamp := entry.Timestamp
		if stamp.IsZero() {
			stamp = time.Now()
		}
		fmt.Fprintf(os.Stderr, "%5.1fs | %s %-5s %s\n",
			time.Since(start).Seconds(),
			stamp.Format("15:04:05"),
			strings.ToUpper(string(entry.Level)),
			entry.Message)
	}
}

// levelOrder ranks a severity, treating anything unrecognised as info.
func levelOrder(level api.LogLevel) int {
	switch level {
	case api.LogLevelDebug:
		return 0
	case api.LogLevelWarn:
		return 2
	case api.LogLevelError:
		return 3
	default:
		return 1
	}
}
