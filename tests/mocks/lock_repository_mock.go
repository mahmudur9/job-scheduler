package mocks

import (
	"context"
	"sync"
)

type MockLockRepository struct {
	mu sync.Mutex

	TryAcquireFn func(ctx context.Context, nodeId string) (bool, error)
	RenewFn      func(ctx context.Context, nodeId string) error
	ReleaseFn    func(ctx context.Context, nodeId string) error
}

func (mock *MockLockRepository) TryAcquire(ctx context.Context, nodeId string) (bool, error) {
	return mock.TryAcquireFn(ctx, nodeId)
}

func (mock *MockLockRepository) Renew(ctx context.Context, nodeId string) error {
	return mock.RenewFn(ctx, nodeId)
}

func (mock *MockLockRepository) Release(ctx context.Context, nodeId string) error {
	return mock.ReleaseFn(ctx, nodeId)
}
