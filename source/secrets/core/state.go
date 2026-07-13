package core

import "context"

const (
	ActiveSpaceKey     = "active:space"
	ActiveAlgorithmKey = "active:algorithm"

	DefaultActiveSpace = "default"
	DefaultAlgorithm   = "aes-gcm"
)

type State struct {
	ActiveSpace     string
	ActiveAlgorithm string
}

func DefaultState() State {
	return State{
		ActiveSpace:     DefaultActiveSpace,
		ActiveAlgorithm: DefaultAlgorithm,
	}
}

func (s State) ToMap() map[string]string {
	return map[string]string{
		ActiveSpaceKey:     s.ActiveSpace,
		ActiveAlgorithmKey: s.ActiveAlgorithm,
	}
}

func StateFromMap(m map[string]string) State {
	s := DefaultState()
	if v := m[ActiveSpaceKey]; v != "" {
		s.ActiveSpace = v
	}
	if v := m[ActiveAlgorithmKey]; v != "" {
		s.ActiveAlgorithm = v
	}
	return s
}

type StateRepository interface {
	Load(ctx context.Context) (State, error)
	Save(ctx context.Context, state State) error
}
