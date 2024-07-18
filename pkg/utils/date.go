package utils

import "time"

// Date returns a time.Time object with the given year, month, and day.
func Date(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
