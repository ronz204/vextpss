package funcs

import (
	"vextpss/source/secrets/core"
	"vextpss/source/shared/terminal"
)

type Deps struct {
	Repo     core.Repository
	Cryp     core.Encryptor
	Prompter *terminal.Prompter
}
