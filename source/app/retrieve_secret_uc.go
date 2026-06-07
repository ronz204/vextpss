package app

import (
	"context"

	"vextpss/source/core"
	"vextpss/source/shared"
)

// RetrieveSecretUC is the use case for fetching and decrypting a stored secret.
type RetrieveSecretUC struct {
	repo      core.SecretRepository
	encryptor core.Encryptor
}

func NewRetrieveSecretUC(repo core.SecretRepository, enc core.Encryptor) *RetrieveSecretUC {
	return &RetrieveSecretUC{repo: repo, encryptor: enc}
}

// RetrieveSecretRequest carries the lookup key and the master password for decryption.
type RetrieveSecretRequest struct {
	Name           string
	MasterPassword []byte
}

// RetrieveSecretResponse contains the decrypted payload ready for display.
type RetrieveSecretResponse struct {
	Name    string
	Type    string
	Payload core.SecretPayload
}

// Execute fetches the encrypted record, decrypts it, and deserialises the payload.
func (uc *RetrieveSecretUC) Execute(ctx context.Context, req RetrieveSecretRequest) (*RetrieveSecretResponse, error) {
	secret, ciphertext, err := uc.repo.GetByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	plaintext, err := uc.encryptor.Decrypt(ctx, req.MasterPassword, secret.Salt, secret.Nonce, ciphertext)
	if err != nil {
		return nil, core.ErrDecryptionFailed
	}
	defer shared.Zero(plaintext)

	payload, err := unmarshalPayload(secret.Type, plaintext)
	if err != nil {
		return nil, err
	}

	return &RetrieveSecretResponse{
		Name:    secret.Name,
		Type:    secret.Type,
		Payload: payload,
	}, nil
}
