package tests

import (
	"context"
	"encoding/json"
	"errors"
	"job-scheduler/internal/domain"
	"job-scheduler/internal/usecase"
	"job-scheduler/tests/mocks"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestScheduler_Run_Success(t *testing.T) {
	jobID := uuid.New()

	mockRepo := &mocks.MockJobRepository{
		FetchAndMarkQueuedFn: func(limit int) ([]domain.Job, error) {
			return []domain.Job{
				{
					Id:           jobID,
					Payload:      "test",
					ScheduleTime: time.Now(),
				},
			}, nil
		},
		MarkScheduledFn: func(jobID uuid.UUID) error {
			return nil
		},
	}

	mockQueue := &mocks.MockJobQueue{
		PublishFn: func(body []byte) error {
			return nil
		},
	}

	s := usecase.NewSchedulerUsecase(mockRepo, mockQueue, "node-1")

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(1500 * time.Millisecond)
		cancel()
	}()

	s.Run(ctx)

	if mockRepo.FetchCalled == 0 {
		t.Fatal("expected FetchDueJobs to be called")
	}

	if mockQueue.PublishCalled == 0 {
		t.Fatal("expected Publish to be called")
	}

	// Validate message format
	if len(mockQueue.Messages) > 0 {
		var msg domain.JobMessage
		err := json.Unmarshal(mockQueue.Messages[0], &msg)
		if err != nil {
			t.Fatal("invalid message format")
		}

		if msg.JobId != jobID {
			t.Fatal("job ID mismatch")
		}
	}
}

func TestScheduler_Run_FetchError(t *testing.T) {
	mockRepo := &mocks.MockJobRepository{
		FetchAndMarkQueuedFn: func(limit int) ([]domain.Job, error) {
			return nil, errors.New("db error")
		},
		MarkScheduledFn: func(jobID uuid.UUID) error {
			return nil
		},
	}

	mockQueue := &mocks.MockJobQueue{
		PublishFn: func(body []byte) error {
			return nil
		},
	}

	s := usecase.NewSchedulerUsecase(mockRepo, mockQueue, "node-1")

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(1200 * time.Millisecond)
		cancel()
	}()

	s.Run(ctx)

	if mockRepo.FetchCalled == 0 {
		t.Fatal("expected FetchDueJobs to be called")
	}

	if mockQueue.PublishCalled != 0 {
		t.Fatal("publish should not be called on fetch error")
	}
}

func TestScheduler_Run_PublishError(t *testing.T) {
	jobID := uuid.New()

	mockRepo := &mocks.MockJobRepository{
		FetchAndMarkQueuedFn: func(limit int) ([]domain.Job, error) {
			return []domain.Job{
				{
					Id:           jobID,
					Payload:      "test",
					ScheduleTime: time.Now(),
				},
			}, nil
		},
		MarkScheduledFn: func(jobID uuid.UUID) error {
			return nil
		},
	}

	mockQueue := &mocks.MockJobQueue{
		PublishFn: func(body []byte) error {
			return errors.New("queue down")
		},
	}

	s := usecase.NewSchedulerUsecase(mockRepo, mockQueue, "node-1")

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(1200 * time.Millisecond)
		cancel()
	}()

	s.Run(ctx)

	if mockQueue.PublishCalled == 0 {
		t.Fatal("expected publish to be attempted")
	}

	if mockRepo.MarkCalled == 0 {
		t.Fatal("MarkQueued should be called when publish fails")
	}
}

func TestScheduler_Run_GracefulShutdown(t *testing.T) {
	mockRepo := &mocks.MockJobRepository{
		FetchAndMarkQueuedFn: func(limit int) ([]domain.Job, error) {
			time.Sleep(2 * time.Second) // simulate slow DB
			return nil, nil
		},
		MarkScheduledFn: func(jobID uuid.UUID) error {
			return nil
		},
	}

	mockQueue := &mocks.MockJobQueue{
		PublishFn: func(body []byte) error {
			return nil
		},
	}

	s := usecase.NewSchedulerUsecase(mockRepo, mockQueue, "node-1")

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(500 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	s.Run(ctx)
	elapsed := time.Since(start)

	if elapsed > 3*time.Second {
		t.Fatal("scheduler did not shutdown gracefully")
	}
}
