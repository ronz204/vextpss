package app

import (
	"encoding/json"
	"fmt"

	"vextpss/source/core"
	"vextpss/source/core/secrets"
)

// payloadDecoder is a function that unmarshals JSON bytes into a concrete SecretPayload.
type payloadDecoder func([]byte) (core.SecretPayload, error)

// decoderRegistry maps each secret type to its JSON decoder.
// Adding a new secret type requires one new entry here.
var decoderRegistry = map[string]payloadDecoder{
	secrets.TypeAccount: func(b []byte) (core.SecretPayload, error) {
		var p secrets.AccountSecret
		return &p, json.Unmarshal(b, &p)
	},
	secrets.TypeCredit: func(b []byte) (core.SecretPayload, error) {
		var p secrets.CreditSecret
		return &p, json.Unmarshal(b, &p)
	},
}

// marshalPayload serialises any SecretPayload to JSON bytes.
func marshalPayload(payload core.SecretPayload) ([]byte, error) {
	return json.Marshal(payload)
}

// unmarshalPayload deserialises JSON bytes into the correct SecretPayload type.
func unmarshalPayload(secretType string, data []byte) (core.SecretPayload, error) {
	dec, ok := decoderRegistry[secretType]
	if !ok {
		return nil, fmt.Errorf("unknown secret type: %s", secretType)
	}
	p, err := dec(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s secret: %w", secretType, err)
	}
	return p, nil
}
