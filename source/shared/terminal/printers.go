package terminal

import (
	"fmt"

	"vextpss/source/shared/payloads/accounts"
	"vextpss/source/shared/payloads/finances"
)

func PrintAccount(a accounts.Account) {
	field("Username", a.Username)
	field("Password", string(a.Password))
}

func PrintFinance(f finances.Finance) {
	fmt.Println("  Card")
	field("  Number", f.Card.Number)
	field("  PIN", string(f.Card.Pin))
	field("  Security Code", string(f.Card.SecurityCode))
	field("  Expiration", fmt.Sprintf("%02d/%d", f.Card.ExpirationMonth, f.Card.ExpirationYear))
	fmt.Println()
	fmt.Println("  Mobile Banking")
	field("  Username", f.Mobile.Username)
	field("  Password", string(f.Mobile.Password))
	field("  Virtual Key", string(f.Mobile.VirtualKey))
	field("  Cellphone", f.Mobile.Cellphone)
}

func field(label, value string) {
	fmt.Printf("  %-16s %s\n", label, value)
}
