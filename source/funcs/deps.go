package funcs

import (
	"vextpss/source/secrets/core"
	"vextpss/source/shared/terminal"
)

type Deps struct {
	Prompter    *terminal.Prompter
	SecretRepo  core.SecretRepository
	SpaceRepo   core.SpaceRepository
	StateRepo   core.StateRepository
	Cypher      core.Encryptor
	ActiveSpace string
}
