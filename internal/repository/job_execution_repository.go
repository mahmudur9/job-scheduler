package repository

import "github.com/google/uuid"

type JobExecutionRepository interface {
	TryStartExecution(execKey string, jobID uuid.UUID, workerID string) (bool, error)
}
