package aesgcm

const algorithmID = "aes-gcm-argon2id"

type Argon2Config struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
}

type Config struct {
	Argon    Argon2Config
	SaltLen  int
	NonceLen int
}

func DefaultConfig() Config {
	return Config{
		Argon: Argon2Config{
			Time:    3,
			Memory:  64 * 1024,
			Threads: 2,
			KeyLen:  32,
		},
		SaltLen:  16,
		NonceLen: 12,
	}
}
