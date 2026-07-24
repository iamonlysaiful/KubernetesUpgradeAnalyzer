package aks

import "time"

// fixedTime returns a fixed time for deterministic tests.
func fixedTime() time.Time {
	return time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
}
