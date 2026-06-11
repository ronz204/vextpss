package password

import (
	"crypto/rand"
	"math/big"
)

const (
	charsetLower   = "abcdefghijklmnopqrstuvwxyz"
	charsetUpper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charsetDigits  = "0123456789"
	charsetSymbols = "!@#$%^&*()_+-=[]{}|;:,.<>?"
)

// GeneratePassword returns a cryptographically secure random password drawn from
// lowercase + uppercase + digits + (optionally) symbols.
func Generate(length int, useSymbols bool) (string, error) {
	charset := charsetLower + charsetUpper + charsetDigits
	if useSymbols {
		charset += charsetSymbols
	}
	result := make([]byte, length)
	max := big.NewInt(int64(len(charset)))
	for i := range result {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		result[i] = charset[idx.Int64()]
	}
	return string(result), nil
}
