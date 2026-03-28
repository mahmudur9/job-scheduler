package repository

import (
	"job-scheduler/internal/domain"

	"github.com/google/uuid"
)

type JobRepository interface {
	CreateJob(job *domain.Job) error
	FetchDueJobs(limit int) ([]domain.Job, error)
	MarkQueued(jobID uuid.UUID) error
}
