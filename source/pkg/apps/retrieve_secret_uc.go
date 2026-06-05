package apps

import (
	"context"
	"encoding/json"
	"fmt"

	"vextpss/source/core"
	"vextpss/source/core/secrets"
	"vextpss/source/dal"
	"vextpss/source/pkg/shared"
	"vextpss/source/pkg/tokens"
)

// RetrieveSecretUC is the use case for fetching and decrypting a stored secret.
type RetrieveSecretUC struct {
	repo      dal.SecretRepository
	encryptor tokens.Encryptor
}

func NewRetrieveSecretUC(repo dal.SecretRepository, enc tokens.Encryptor) *RetrieveSecretUC {
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

	payload, err := deserialisePayload(secret.Type, plaintext)
	if err != nil {
		return nil, err
	}

	return &RetrieveSecretResponse{
		Name:    secret.Name,
		Type:    secret.Type,
		Payload: payload,
	}, nil
}

func deserialisePayload(secretType string, plaintext []byte) (core.SecretPayload, error) {
	switch secretType {
	case "account":
		var ap secrets.AccountSecret
		if err := json.Unmarshal(plaintext, &ap); err != nil {
			return nil, fmt.Errorf("failed to parse account secret: %w", err)
		}
		return &ap, nil
	default:
		return nil, fmt.Errorf("unknown secret type: %s", secretType)
	}
}
