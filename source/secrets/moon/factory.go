package moon

import (
	"encoding/json"
	"fmt"

	"vextpss/source/secrets/core"
	"vextpss/source/secrets/moon/accounts"
	"vextpss/source/secrets/moon/finances"
)

var registry = map[string]func() core.Payload{
	core.TypeAccount: func() core.Payload { return &accounts.Account{} },
	core.TypeFinance: func() core.Payload { return &finances.Finance{} },
}

func IsKnownType(t string) bool {
	_, ok := registry[t]
	return ok
}

func Decode(secretType string, data []byte) (core.Payload, error) {
	factory, ok := registry[secretType]
	if !ok {
		return nil, fmt.Errorf("unknown secret type %q", secretType)
	}
	p := factory()
	if err := json.Unmarshal(data, p); err != nil {
		return nil, core.ErrCorruptPayload
	}
	return p, nil
}
