package addsecret

import (
	"vextpss/source/secrets/core"
	"vextpss/source/shared/terminal"
)

type payloadCollector func(*terminal.Prompter) (core.Payload, error)

var collectors = map[string]payloadCollector{
	core.TypeAccount: collectAccount,
	core.TypeFinance: collectFinance,
}
