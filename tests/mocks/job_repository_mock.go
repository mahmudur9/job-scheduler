package mocks

import (
	"job-scheduler/internal/domain"
	"sync"

	"github.com/google/uuid"
)

type MockJobRepository struct {
	mu sync.Mutex

	CreateJobFunc        func(job *domain.Job) error
	FetchAndMarkQueuedFn func(limit int) ([]domain.Job, error)
	MarkScheduledFn      func(jobID uuid.UUID) error

	FetchCalled int
	MarkCalled  int
}

func (m *MockJobRepository) CreateJob(job *domain.Job) error {
	return m.CreateJobFunc(job)
}

func (m *MockJobRepository) FetchAndMarkQueued(limit int) ([]domain.Job, error) {
	m.mu.Lock()
	m.FetchCalled++
	m.mu.Unlock()

	return m.FetchAndMarkQueuedFn(limit)
}

func (m *MockJobRepository) MarkScheduled(jobID uuid.UUID) error {
	m.mu.Lock()
	m.MarkCalled++
	m.mu.Unlock()

	return m.MarkScheduledFn(jobID)
}
