package tests

import (
	"errors"
	"job-scheduler/internal/domain"
	"job-scheduler/internal/requests"
	"job-scheduler/internal/usecase"
	"job-scheduler/tests/mocks"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJobUsecase_Create_Success(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockJobRepository{}

	var capturedJob *domain.Job

	mockRepo.CreateJobFunc = func(job *domain.Job) error {
		capturedJob = job
		return nil
	}

	mockLogger := &mocks.MockLogger{
		InfoFn:  func(msg string) {},
		ErrorFn: func(err error, msg string) {},
	}

	jobUsecase := usecase.NewJobUsecase(mockRepo, mockLogger)

	req := &requests.JobRequest{
		Payload:      "test payload",
		ScheduleTime: time.Now().Add(1 * time.Hour),
	}

	// Act
	err := jobUsecase.Create(req)

	// Assert
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if capturedJob == nil {
		t.Fatal("expected job to be passed to repository")
	}

	if capturedJob.Id == uuid.Nil {
		t.Error("expected job ID to be set")
	}

	if capturedJob.Payload != req.Payload {
		t.Errorf("expected payload %v, got %v", req.Payload, capturedJob.Payload)
	}

	if !capturedJob.ScheduleTime.Equal(req.ScheduleTime) {
		t.Errorf("expected schedule time %v, got %v", req.ScheduleTime, capturedJob.ScheduleTime)
	}

	if capturedJob.Status != "SCHEDULED" {
		t.Errorf("expected status SCHEDULED, got %v", capturedJob.Status)
	}

	if capturedJob.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestJobUsecase_Create_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockJobRepository{}

	expectedErr := errors.New("db error")

	mockRepo.CreateJobFunc = func(job *domain.Job) error {
		return expectedErr
	}

	mockLogger := &mocks.MockLogger{
		InfoFn:  func(msg string) {},
		ErrorFn: func(err error, msg string) {},
	}

	jobUsecase := usecase.NewJobUsecase(mockRepo, mockLogger)

	req := &requests.JobRequest{
		Payload:      "test payload",
		ScheduleTime: time.Now(),
	}

	// Act
	err := jobUsecase.Create(req)

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != expectedErr {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}
