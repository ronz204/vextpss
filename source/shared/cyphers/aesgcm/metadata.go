package aesgcm

import (
	"fmt"

	"vextpss/source/secrets/core"
)

func encodeMetadata(salt, nonce []byte) []byte {
	buf := make([]byte, 0, 2+len(salt)+len(nonce))
	buf = append(buf, byte(len(salt)))
	buf = append(buf, salt...)
	buf = append(buf, byte(len(nonce)))
	buf = append(buf, nonce...)
	return buf
}

func decodeMetadata(b []byte) (salt, nonce []byte, err error) {
	if len(b) < 1 {
		return nil, nil, fmt.Errorf("aesgcm: metadata too short: %w", core.ErrCorruptPayload)
	}
	saltLen := int(b[0])
	if len(b) < 1+saltLen+1 {
		return nil, nil, fmt.Errorf("aesgcm: malformed metadata (salt): %w", core.ErrCorruptPayload)
	}
	salt = b[1 : 1+saltLen]

	rest := b[1+saltLen:]
	nonceLen := int(rest[0])
	if len(rest) < 1+nonceLen {
		return nil, nil, fmt.Errorf("aesgcm: malformed metadata (nonce): %w", core.ErrCorruptPayload)
	}
	nonce = rest[1 : 1+nonceLen]

	return salt, nonce, nil
}
