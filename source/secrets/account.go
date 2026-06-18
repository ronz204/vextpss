package secrets

type AccountSecret struct {
	Username string `json:"username"`
	Password []byte `json:"password"`
}
