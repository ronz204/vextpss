package secrets

const (
	TypeAccount = "account"
	TypeFinance = "finance"
)

// IsKnownType reports whether t is a registered secret type.
// Add new constants here as new types are introduced.
func IsKnownType(t string) bool {
	switch t {
	case TypeAccount, TypeFinance:
		return true
	}
	return false
}
