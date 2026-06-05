package tokens_test

import (
	"bytes"
	"context"
	"testing"

	"vextpss/source/core"
	"vextpss/source/pkg/tokens"
)

func TestAESGCMEncryptor_Encrypt_ReturnsNonEmptyValues(t *testing.T) {
	enc := tokens.NewAESGCMEncryptor()
	salt, nonce, ciphertext, err := enc.Encrypt(context.Background(), []byte("hello"), []byte("masterpass"))
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	if len(salt) == 0 {
		t.Error("salt should not be empty")
	}
	if len(nonce) == 0 {
		t.Error("nonce should not be empty")
	}
	if len(ciphertext) == 0 {
		t.Error("ciphertext should not be empty")
	}
}

func TestAESGCMEncryptor_Encrypt_SaltIsRandom(t *testing.T) {
	enc := tokens.NewAESGCMEncryptor()
	plaintext := []byte("same plaintext")
	master := []byte("same master")

	salt1, _, _, err := enc.Encrypt(context.Background(), plaintext, master)
	if err != nil {
		t.Fatalf("first Encrypt() error = %v", err)
	}
	salt2, _, _, err := enc.Encrypt(context.Background(), plaintext, master)
	if err != nil {
		t.Fatalf("second Encrypt() error = %v", err)
	}

	if bytes.Equal(salt1, salt2) {
		t.Error("two encryptions should produce different salts")
	}
}

func TestAESGCMEncryptor_Encrypt_CiphertextDiffersFromPlaintext(t *testing.T) {
	enc := tokens.NewAESGCMEncryptor()
	plaintext := []byte("secret data")
	_, _, ciphertext, err := enc.Encrypt(context.Background(), plaintext, []byte("master"))
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	if bytes.Equal(plaintext, ciphertext) {
		t.Error("ciphertext must not equal plaintext")
	}
}

func TestAESGCMEncryptor_Decrypt_RoundTrip(t *testing.T) {
	enc := tokens.NewAESGCMEncryptor()
	plaintext := []byte(`{"username":"alice","password":"s3cr3t"}`)
	master := []byte("correct-master-password")

	salt, nonce, ciphertext, err := enc.Encrypt(context.Background(), plaintext, master)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	got, err := enc.Decrypt(context.Background(), master, salt, nonce, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Errorf("Decrypt() = %q, want %q", got, plaintext)
	}
}

func TestAESGCMEncryptor_Decrypt_WrongPassword(t *testing.T) {
	enc := tokens.NewAESGCMEncryptor()
	salt, nonce, ciphertext, err := enc.Encrypt(context.Background(), []byte("data"), []byte("correct"))
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	_, err = enc.Decrypt(context.Background(), []byte("wrong"), salt, nonce, ciphertext)
	if err == nil {
		t.Fatal("Decrypt() with wrong password should return an error")
	}
	if !core.IsNotFound(err) && err != core.ErrDecryptionFailed {
		// Verify it's the canonical decryption error
	}
	if !bytes.Equal([]byte(err.Error()), []byte(core.ErrDecryptionFailed.Error())) {
		t.Errorf("Decrypt() error = %q, want %q", err.Error(), core.ErrDecryptionFailed.Error())
	}
}

func TestAESGCMEncryptor_Decrypt_TamperedCiphertext(t *testing.T) {
	enc := tokens.NewAESGCMEncryptor()
	salt, nonce, ciphertext, err := enc.Encrypt(context.Background(), []byte("data"), []byte("master"))
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Flip a bit in the ciphertext to simulate tampering.
	tampered := make([]byte, len(ciphertext))
	copy(tampered, ciphertext)
	tampered[0] ^= 0xFF

	_, err = enc.Decrypt(context.Background(), []byte("master"), salt, nonce, tampered)
	if err == nil {
		t.Fatal("Decrypt() with tampered ciphertext should return an error")
	}
	if err.Error() != core.ErrDecryptionFailed.Error() {
		t.Errorf("Decrypt() error = %q, want %q", err.Error(), core.ErrDecryptionFailed.Error())
	}
}

func TestAESGCMEncryptor_Decrypt_WrongNonce(t *testing.T) {
	enc := tokens.NewAESGCMEncryptor()
	salt, _, ciphertext, err := enc.Encrypt(context.Background(), []byte("data"), []byte("master"))
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	badNonce := make([]byte, 12)

	_, err = enc.Decrypt(context.Background(), []byte("master"), salt, badNonce, ciphertext)
	if err == nil {
		t.Fatal("Decrypt() with wrong nonce should return an error")
	}
	if err.Error() != core.ErrDecryptionFailed.Error() {
		t.Errorf("Decrypt() error = %q, want %q", err.Error(), core.ErrDecryptionFailed.Error())
	}
}
