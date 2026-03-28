package repository

import "job-scheduler/internal/domain"

type JobRepository interface {
	CreateJob(job *domain.Job) error
}
