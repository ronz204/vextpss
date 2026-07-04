package finances

import (
	"fmt"

	"vextpss/source/secrets/core"
)

var _ core.Payload = Finance{}

type Finance struct {
	Card   Card   `json:"card"`
	Mobile Mobile `json:"mobile"`
}

func NewFinance(card Card, mobile Mobile) Finance {
	return Finance{Card: card, Mobile: mobile}
}

func (f Finance) Type() string {
	return core.TypeFinance
}

func (f Finance) Display() {
	fmt.Println("  Card")
	fmt.Printf("  %-16s %s\n", "  Number", f.Card.Number)
	fmt.Printf("  %-16s %s\n", "  PIN", string(f.Card.Pin))
	fmt.Printf("  %-16s %s\n", "  Security Code", string(f.Card.SecurityCode))
	fmt.Printf("  %-16s %s\n", "  Expiration", fmt.Sprintf("%02d/%d", f.Card.ExpirationMonth, f.Card.ExpirationYear))
	fmt.Println()
	fmt.Println("  Banking")
	fmt.Printf("  %-16s %s\n", "  Username", f.Mobile.Username)
	fmt.Printf("  %-16s %s\n", "  Password", string(f.Mobile.Password))
	fmt.Printf("  %-16s %s\n", "  Virtual Key", string(f.Mobile.VirtualKey))
	fmt.Printf("  %-16s %s\n", "  Cellphone", f.Mobile.Cellphone)
}

func (f Finance) Validate() error {
	if err := f.Card.Validate(); err != nil {
		return err
	}
	return f.Mobile.Validate()
}
