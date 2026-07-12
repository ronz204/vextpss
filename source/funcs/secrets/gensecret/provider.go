package gensecret

import (
	"fmt"

	"vextpss/source/shared/memory"
	"vextpss/source/shared/passgen"
)

func run(length int, symbols bool) error {
	b, err := passgen.Generate(length, symbols)
	if err != nil {
		return err
	}
	defer memory.Cleaner(b)
	fmt.Println(string(b))
	return nil
}
