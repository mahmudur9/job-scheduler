package usecase

import (
	"job-scheduler/internal/domain"
	"job-scheduler/internal/requests"
	"time"

	"github.com/google/uuid"
)

type JobUsecase struct {
	jobRepository JobRepository
}

func NewJobUsecase(jobRepository JobRepository) *JobUsecase {
	return &JobUsecase{
		jobRepository,
	}
}

func (j *JobUsecase) Create(jobRequest *requests.JobRequest) error {
	var job domain.Job
	job.Id = uuid.New()
	job.Payload = jobRequest.Payload
	job.ScheduleTime = jobRequest.ScheduleTime
	job.Status = "SCHEDULED"
	job.CreatedAt = time.Now()
	return j.jobRepository.CreateJob(&job)
}
