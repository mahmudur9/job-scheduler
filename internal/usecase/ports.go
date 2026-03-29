package usecase

import (
	"job-scheduler/internal/domain"

	"github.com/google/uuid"
)

type JobRepository interface {
	CreateJob(job *domain.Job) error
	FetchDueJobs(limit int) ([]domain.Job, error)
	MarkQueued(jobID uuid.UUID) error
}

type JobExecutionRepository interface {
	TryStartExecution(execKey string, jobID uuid.UUID, workerID string) (bool, error)
}

type JobQueue interface {
	Publish(body []byte) error
}

type JobConsumer interface {
	Consume(queueName string) (<-chan Message, error)
}

type Message interface {
	Body() []byte
	Ack() error
	Nack(requeue bool) error
}
