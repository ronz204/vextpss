package memory

// ======================================
// Cleaner overwrites a byte slice with zeros to clear sensitive data from memory.
// Call immediately after the slice is no longer needed.
// ======================================
func Cleaner(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
