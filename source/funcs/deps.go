package funcs

import (
	"vextpss/source/secrets"
	"vextpss/source/shared/terminal"
)

type Deps struct {
	Repo     secrets.Repository
	Cryp     secrets.Encryptor
	Prompter *terminal.Prompter
}
