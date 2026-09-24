package envd

import (
	"strconv"
	"strings"
)

// envd versions at which a capability became available. A sandbox built from an
// older template reports an older version, and the SDK degrades rather than
// calling something that is not there.
var (
	VersionMinimum        = Version{0, 1, 0}
	VersionRecursiveWatch = Version{0, 1, 4}
	VersionMetrics        = Version{0, 1, 5}
	VersionDiskMetrics    = Version{0, 2, 4}
	VersionStdin          = Version{0, 3, 0}
	VersionDefaultUser    = Version{0, 4, 0}
)

// Version is a major.minor.patch triple.
type Version [3]int

// ParseVersion reads a version string, treating anything unparsable as 0.
// An empty or malformed version therefore sorts below every gate, which fails
// closed: the capability is assumed absent.
func ParseVersion(raw string) Version {
	var parsed Version
	parts := strings.Split(strings.TrimPrefix(raw, "v"), ".")
	for i := 0; i < 3 && i < len(parts); i++ {
		if n, err := strconv.Atoi(parts[i]); err == nil {
			parsed[i] = n
		}
	}
	return parsed
}

// LessThan reports whether v is older than other.
func (v Version) LessThan(other Version) bool {
	for i := range v {
		if v[i] != other[i] {
			return v[i] < other[i]
		}
	}
	return false
}

// supports reports whether this envd is new enough for a capability.
func (v Version) Supports(gate Version) bool {
	return !v.LessThan(gate)
}
