package usecase

import (
	"context"
	"job-scheduler/internal/domain"

	"github.com/google/uuid"
)

type JobRepository interface {
	CreateJob(job *domain.Job) error
	FetchAndMarkQueued(limit int) ([]domain.Job, error)
	RestoreToScheduled(jobID uuid.UUID) error
}

type JobExecutionRepository interface {
	TryStartExecution(execKey string, jobID uuid.UUID, workerID string) (bool, error)
}

type LockRepository interface {
	TryAcquire(ctx context.Context, nodeId string) (bool, error)
	Renew(ctx context.Context, nodeId string) error
	Release(ctx context.Context, nodeId string) error
}

type JobQueue interface {
	Publish(body []byte, routingKey string) error
}

type JobConsumer interface {
	Consume(routingKey string, queueName string) (<-chan Message, error)
}

type Message interface {
	Body() []byte
	Ack() error
	Nack(requeue bool) error
}

type Logger interface {
	Info(msg string)
	Error(err error, msg string)
}
