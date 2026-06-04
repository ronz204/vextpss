package crypto_test

import (
	"context"
	"bytes"
	"testing"
	"vextpss/source/pkg/domain"
	infracrypto "vextpss/source/pkg/infrastructure/crypto"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	enc := infracrypto.NewAESGCMEncryptor()
	ctx := context.Background()

	plaintext := []byte("super secret data")
	masterPwd := []byte("correct-password")

	salt, nonce, ciphertext, err := enc.Encrypt(ctx, plaintext, masterPwd)
	if err != nil {
		t.Fatalf("Encrypt() failed: %v", err)
	}

	recovered, err := enc.Decrypt(ctx, masterPwd, salt, nonce, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() failed: %v", err)
	}

	if !bytes.Equal(recovered, plaintext) {
		t.Errorf("recovered %q, want %q", recovered, plaintext)
	}
}

func TestDecryptWrongPasswordReturnsErrDecryptionFailed(t *testing.T) {
	enc := infracrypto.NewAESGCMEncryptor()
	ctx := context.Background()

	salt, nonce, ciphertext, err := enc.Encrypt(ctx, []byte("data"), []byte("right-password"))
	if err != nil {
		t.Fatalf("Encrypt() failed: %v", err)
	}

	_, err = enc.Decrypt(ctx, []byte("wrong-password"), salt, nonce, ciphertext)
	if err != domain.ErrDecryptionFailed {
		t.Errorf("expected ErrDecryptionFailed, got %v", err)
	}
}

func TestEncryptProducesDifferentCiphertextsForSamePlaintext(t *testing.T) {
	enc := infracrypto.NewAESGCMEncryptor()
	ctx := context.Background()

	plaintext := []byte("same data")
	masterPwd := []byte("password")

	_, _, ct1, _ := enc.Encrypt(ctx, plaintext, masterPwd)
	_, _, ct2, _ := enc.Encrypt(ctx, plaintext, masterPwd)

	// Fresh salt each time → different ciphertext even for identical input.
	if bytes.Equal(ct1, ct2) {
		t.Error("two encryptions of same plaintext should produce different ciphertexts")
	}
}
