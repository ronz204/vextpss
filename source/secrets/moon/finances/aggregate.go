package finances

import "vextpss/source/secrets/core"

type Finance struct {
	Card   Card   `json:"card"`
	Mobile Mobile `json:"mobile"`
}

func NewFinance(card Card, mobile Mobile) Finance {
	return Finance{Card: card, Mobile: mobile}
}

func (f Finance) Display() string {
	return core.TypeFinance
}

func (f Finance) Validate() error {
	if err := f.Card.Validate(); err != nil {
		return err
	}
	return f.Mobile.Validate()
}
