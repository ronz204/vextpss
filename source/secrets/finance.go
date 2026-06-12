package secrets

type FinanceSecret struct {
	CardPin         []byte `json:"card_pin"`
	CardNumber      string `json:"card_number"`
	SecurityCode    []byte `json:"security_code"`
	ExpirationMonth int    `json:"expiration_month"`
	ExpirationYear  int    `json:"expiration_year"`
	BankUsername    string `json:"bank_username"`
	BankPassword    []byte `json:"bank_password"`
	BankVirtualKey  []byte `json:"bank_virtual_key"`
	BankCellphone   string `json:"bank_cellphone"`
}
