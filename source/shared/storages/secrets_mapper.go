package storages

import "vextpss/source/secrets/core"

func toRecord(s core.Secret, spaceID int64) SecretRecord {
	return SecretRecord{
		SpaceID:    spaceID,
		Name:       s.Name,
		Type:       s.Type,
		Algorithm:  s.Encrypted.Algorithm,
		Ciphertext: s.Encrypted.Ciphertext,
		Metadata:   s.Encrypted.Metadata,
	}
}

func toSecret(r SecretRecord, spaceName string) core.Secret {
	return core.Secret{
		ID:    r.ID,
		Space: spaceName,
		Name:  r.Name,
		Type:  r.Type,
		Encrypted: core.Encrypted{
			Algorithm:  r.Algorithm,
			Ciphertext: r.Ciphertext,
			Metadata:   r.Metadata,
		},
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
