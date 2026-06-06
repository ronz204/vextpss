package mocks

import (
	"context"

	"vextpss/source/core"
	"vextpss/source/dal"
)

// MockRepository is a configurable test double for dal.SecretRepository.
type MockRepository struct {
	SaveFn      func(ctx context.Context, secret *core.Secret, encrypted []byte) error
	GetByNameFn func(ctx context.Context, name string) (*core.Secret, []byte, error)
	ListAllFn   func(ctx context.Context) ([]core.Secret, error)
	GetAllFn    func(ctx context.Context) ([]dal.FullRecord, error)
	DeleteFn    func(ctx context.Context, name string) error
	UpdateFn    func(ctx context.Context, secret *core.Secret, encrypted []byte) error
}

func (m *MockRepository) Save(ctx context.Context, secret *core.Secret, encrypted []byte) error {
	if m.SaveFn != nil {
		return m.SaveFn(ctx, secret, encrypted)
	}
	return nil
}

func (m *MockRepository) GetByName(ctx context.Context, name string) (*core.Secret, []byte, error) {
	if m.GetByNameFn != nil {
		return m.GetByNameFn(ctx, name)
	}
	return nil, nil, nil
}

func (m *MockRepository) ListAll(ctx context.Context) ([]core.Secret, error) {
	if m.ListAllFn != nil {
		return m.ListAllFn(ctx)
	}
	return nil, nil
}

func (m *MockRepository) GetAll(ctx context.Context) ([]dal.FullRecord, error) {
	if m.GetAllFn != nil {
		return m.GetAllFn(ctx)
	}
	return nil, nil
}

func (m *MockRepository) Delete(ctx context.Context, name string) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, name)
	}
	return nil
}

func (m *MockRepository) Update(ctx context.Context, secret *core.Secret, encrypted []byte) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, secret, encrypted)
	}
	return nil
}
