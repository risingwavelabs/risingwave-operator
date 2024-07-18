package utils

import (
	"strings"
	"time"
)

// IsNightlyVersionAfter checks if the version is a nightly version and is after the given date.
func IsNightlyVersionAfter(version string, format string, date time.Time) bool {
	if strings.HasPrefix(version, "nightly-") {
		version = strings.TrimPrefix(version, "nightly-")
		t, err := time.Parse(format, version)
		if err != nil {
			return false
		}
		return t.After(date)
	}
	return false
}

// IsRisingWaveNightlyVersionAfter checks if the version is a nightly version and is after the given date.
func IsRisingWaveNightlyVersionAfter(version string, date time.Time) bool {
	return IsNightlyVersionAfter(version, "2006-01-02", date)
}
