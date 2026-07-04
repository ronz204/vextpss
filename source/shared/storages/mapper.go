package storages

import "vextpss/source/secrets/core"

func toRecord(s core.Secret) SecretRecord {
	return SecretRecord{
		Name:       s.Name,
		Type:       s.Type,
		Algorithm:  s.Encrypted.Algorithm,
		Ciphertext: s.Encrypted.Ciphertext,
		Metadata:   s.Encrypted.Metadata,
	}
}

func toSecret(r SecretRecord) core.Secret {
	return core.Secret{
		ID:   r.ID,
		Name: r.Name,
		Type: r.Type,
		Encrypted: core.Encrypted{
			Algorithm:  r.Algorithm,
			Ciphertext: r.Ciphertext,
			Metadata:   r.Metadata,
		},
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
