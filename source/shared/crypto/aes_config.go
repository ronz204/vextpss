package crypto

// Argon2Config holds parameters for key derivation.
type Argon2Config struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
}

// CryptoConfig bundles all cryptographic parameters.
type AESGCMConfig struct {
	Argon    Argon2Config
	SaltLen  int
	NonceLen int
}

// DefaultConfig returns production-grade parameters.
func DefaultConfig() AESGCMConfig {
	return AESGCMConfig{
		Argon: Argon2Config{
			Time:    3,
			Memory:  64 * 1024, // 64 MB
			Threads: 2,
			KeyLen:  32,
		},
		SaltLen:  16,
		NonceLen: 12,
	}
}
