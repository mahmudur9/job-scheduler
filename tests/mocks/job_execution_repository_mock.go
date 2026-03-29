package mocks

import (
	"sync"

	"github.com/google/uuid"
)

type MockJobExecutionRepository struct {
	mu sync.Mutex

	TryStartExecutionFn func(execKey string, jobID uuid.UUID, workerID string) (bool, error)

	Calls int
}

func (m *MockJobExecutionRepository) TryStartExecution(execKey string, jobID uuid.UUID, workerID string) (bool, error) {
	m.mu.Lock()
	m.Calls++
	m.mu.Unlock()

	if m.TryStartExecutionFn != nil {
		return m.TryStartExecutionFn(execKey, jobID, workerID)
	}
	return true, nil
}
