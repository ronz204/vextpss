package password

import "time"

// Zero overwrites a byte slice with zeros to clear sensitive data from memory.
// Call immediately after the slice is no longer needed.
func Zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

// Now returns the current UTC time. Centralised so tests can swap it out easily.
func Now() time.Time {
	return time.Now().UTC()
}
