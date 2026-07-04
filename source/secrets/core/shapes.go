package core

const (
	TypeAccount = "account"
	TypeFinance = "finance"
)

var knownTypes = map[string]bool{
	TypeAccount: true,
	TypeFinance: true,
}

func IsKnownType(t string) bool {
	return knownTypes[t]
}
