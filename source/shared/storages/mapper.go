package storages

import "vextpss/source/secrets"

func toRecord(s secrets.Secret) SecretRecord {
	return SecretRecord{
		Name:       s.Name,
		Type:       s.Type,
		Algorithm:  s.Encrypted.Algorithm,
		Ciphertext: s.Encrypted.Ciphertext,
		Metadata:   s.Encrypted.Metadata,
	}
}

func toSecret(r SecretRecord) secrets.Secret {
	return secrets.Secret{
		ID:   r.ID,
		Name: r.Name,
		Type: r.Type,
		Encrypted: secrets.Encrypted{
			Algorithm:  r.Algorithm,
			Ciphertext: r.Ciphertext,
			Metadata:   r.Metadata,
		},
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
