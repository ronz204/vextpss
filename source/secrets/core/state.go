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

type StateRepository interface {
	Load(ctx context.Context) (State, error)
	Save(ctx context.Context, state State) error
}
