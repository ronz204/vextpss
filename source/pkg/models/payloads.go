package models

// AccountPayload is the typed payload for secrets with type = "account".
// This struct is serialized to JSON and then encrypted before storage.
type AccountPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// CardPayload is the typed payload for secrets with type = "card". (Phase 2)
type CardPayload struct {
	CardNumber string `json:"card_number"`
	CVV        string `json:"cvv"`
	Expiration string `json:"expiration"`
	PIN        string `json:"pin"`
}

// NotePayload is the typed payload for secrets with type = "note". (Phase 2)
type NotePayload struct {
	Content string `json:"content"`
}
