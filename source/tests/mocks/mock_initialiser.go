package mocks

import "context"

// MockInitialiser is a configurable test double for apps.StorageInitialiser.
type MockInitialiser struct {
	InitFn    func(ctx context.Context) error
	DBPathVal string
}

func (m *MockInitialiser) Init(ctx context.Context) error {
	if m.InitFn != nil {
		return m.InitFn(ctx)
	}
	return nil
}

func (m *MockInitialiser) DBPath() string {
	return m.DBPathVal
}
