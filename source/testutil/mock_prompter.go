package testutil

// MockPrompter is a test double for ui.Prompter that returns pre-set values without touching stdin.
type MockPrompter struct {
	LineResponse    string
	PasswordBytes   []byte
	ConfirmResponse bool
	Err             error
}

func (m *MockPrompter) ReadLine(_ string) (string, error)     { return m.LineResponse, m.Err }
func (m *MockPrompter) ReadPassword(_ string) ([]byte, error) { return m.PasswordBytes, m.Err }
func (m *MockPrompter) Confirm(_ string) (bool, error)        { return m.ConfirmResponse, m.Err }
func (m *MockPrompter) Zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
