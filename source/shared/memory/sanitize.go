package memory

func Cleaner(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
