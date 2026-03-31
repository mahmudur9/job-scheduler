package usecase

import (
	"job-scheduler/internal/domain"
	"job-scheduler/internal/requests"
	"time"

	"github.com/google/uuid"
)

type JobUsecase struct {
	jobRepository JobRepository
	logger        Logger
}

func NewJobUsecase(jobRepository JobRepository, logger Logger) *JobUsecase {
	return &JobUsecase{
		jobRepository,
		logger,
	}
}

func (j *JobUsecase) Create(jobRequest *requests.JobRequest) error {
	var job domain.Job
	job.Id = uuid.New()
	job.Payload = jobRequest.Payload
	job.ScheduleTime = jobRequest.ScheduleTime
	job.Status = "SCHEDULED"
	job.CreatedAt = time.Now()

	err := j.jobRepository.CreateJob(&job)
	if err != nil {
		j.logger.Error(err, "Failed to create job")
		return err
	}
	return nil
}
