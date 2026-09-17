package utils

import "time"

// Now returns the current time in UTC. Store and compare times in UTC; convert to a local
// time zone only for display.
func Now() time.Time {
	return time.Now().UTC()
}
