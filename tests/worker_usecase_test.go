package tests

import (
	"context"
	"encoding/json"
	"errors"
	"job-scheduler/internal/usecase"
	"job-scheduler/tests/mocks"
	"testing"
	"time"

	"job-scheduler/internal/domain"

	"github.com/google/uuid"
)

//
// ---- MOCKS ----
//

type MockMessage struct {
	body []byte

	acked  bool
	nacked bool
}

func (m *MockMessage) Body() []byte {
	return m.body
}

func (m *MockMessage) Ack() error {
	m.acked = true
	return nil
}

func (m *MockMessage) Nack(requeue bool) error {
	m.nacked = true
	return nil
}

// Mock Consumer
type MockConsumer struct {
	Messages chan usecase.Message
	Err      error
}

func (m *MockConsumer) Consume(queueName string) (<-chan usecase.Message, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Messages, nil
}

//
// ---- TESTS ----
//

func TestWorker_Start_Success(t *testing.T) {
	jobID := uuid.New()

	msgBody, _ := json.Marshal(domain.JobMessage{
		JobId:        jobID,
		Payload:      "test",
		ScheduleTime: time.Now().Unix(),
	})

	mockMsg := &MockMessage{body: msgBody}

	msgChan := make(chan usecase.Message, 1)
	msgChan <- mockMsg
	close(msgChan)

	mockConsumer := &MockConsumer{
		Messages: msgChan,
	}

	mockRepo := &mocks.MockJobExecutionRepository{
		TryStartExecutionFn: func(execKey string, jobID uuid.UUID, workerID string) (bool, error) {
			return true, nil
		},
	}

	worker := usecase.NewWorkerUsecase(mockRepo, mockConsumer, "worker-1")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	worker.Start(ctx)

	if !mockMsg.acked {
		t.Fatal("expected message to be acknowledged")
	}

	if mockRepo.Calls == 0 {
		t.Fatal("expected TryStartExecution to be called")
	}
}

func TestWorker_Start_HandleError_ShouldNack(t *testing.T) {
	// invalid JSON → handle() error
	mockMsg := &MockMessage{
		body: []byte("invalid-json"),
	}

	msgChan := make(chan usecase.Message, 1)
	msgChan <- mockMsg
	close(msgChan)

	mockConsumer := &MockConsumer{
		Messages: msgChan,
	}

	mockRepo := &mocks.MockJobExecutionRepository{}

	worker := usecase.NewWorkerUsecase(mockRepo, mockConsumer, "worker-1")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	worker.Start(ctx)

	if !mockMsg.nacked {
		t.Fatal("expected message to be nacked on error")
	}
}

func TestWorker_Start_ExecutionError_ShouldNack(t *testing.T) {
	jobID := uuid.New()

	msgBody, _ := json.Marshal(domain.JobMessage{
		JobId:        jobID,
		Payload:      "test",
		ScheduleTime: time.Now().Unix(),
	})

	mockMsg := &MockMessage{body: msgBody}

	msgChan := make(chan usecase.Message, 1)
	msgChan <- mockMsg
	close(msgChan)

	mockConsumer := &MockConsumer{
		Messages: msgChan,
	}

	mockRepo := &mocks.MockJobExecutionRepository{
		TryStartExecutionFn: func(execKey string, jobID uuid.UUID, workerID string) (bool, error) {
			return false, errors.New("db error")
		},
	}

	worker := usecase.NewWorkerUsecase(mockRepo, mockConsumer, "worker-1")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	worker.Start(ctx)

	if !mockMsg.nacked {
		t.Fatal("expected message to be nacked on execution error")
	}
}

func TestWorker_Start_DuplicateJob_ShouldAck(t *testing.T) {
	jobID := uuid.New()

	msgBody, _ := json.Marshal(domain.JobMessage{
		JobId:        jobID,
		Payload:      "test",
		ScheduleTime: time.Now().Unix(),
	})

	mockMsg := &MockMessage{body: msgBody}

	msgChan := make(chan usecase.Message, 1)
	msgChan <- mockMsg
	close(msgChan)

	mockConsumer := &MockConsumer{
		Messages: msgChan,
	}

	mockRepo := &mocks.MockJobExecutionRepository{
		TryStartExecutionFn: func(execKey string, jobID uuid.UUID, workerID string) (bool, error) {
			return false, nil // duplicate
		},
	}

	worker := usecase.NewWorkerUsecase(mockRepo, mockConsumer, "worker-1")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	worker.Start(ctx)

	if !mockMsg.acked {
		t.Fatal("expected duplicate job to be acknowledged")
	}
}

func TestWorker_Start_ConsumeError(t *testing.T) {
	mockConsumer := &MockConsumer{
		Err: errors.New("consume failed"),
	}

	mockRepo := &mocks.MockJobExecutionRepository{}

	worker := usecase.NewWorkerUsecase(mockRepo, mockConsumer, "worker-1")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Should not panic
	worker.Start(ctx)
}

func TestWorker_Start_GracefulShutdown(t *testing.T) {
	msgChan := make(chan usecase.Message)

	mockConsumer := &MockConsumer{
		Messages: msgChan,
	}

	mockRepo := &mocks.MockJobExecutionRepository{}

	worker := usecase.NewWorkerUsecase(mockRepo, mockConsumer, "worker-1")

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	worker.Start(ctx)
	elapsed := time.Since(start)

	if elapsed > time.Second {
		t.Fatal("worker did not shutdown gracefully")
	}
}
