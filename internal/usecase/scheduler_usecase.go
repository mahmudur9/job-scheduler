package usecase

import (
	"context"
	"encoding/json"
	"job-scheduler/internal/domain"
	"log"
	"time"
)

type SchedulerUsecase struct {
	jobRepository JobRepository
	jobQueue      JobQueue
	nodeId        string
}

func NewSchedulerUsecase(jobRepository JobRepository, jobQueue JobQueue, nodeId string) *SchedulerUsecase {
	return &SchedulerUsecase{
		jobRepository,
		jobQueue,
		nodeId,
	}
}

func (s *SchedulerUsecase) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	log.Println("Scheduler started.")

	for {
		select {
		case <-ctx.Done():
			log.Println("Scheduler stopping gracefully...")
			return

		case <-ticker.C:
			jobs, err := s.jobRepository.FetchDueJobs(10)
			if err != nil {
				log.Println("fetch jobs error:", err)
				continue
			}

			for _, j := range jobs {
				// Check shutdown signal between jobs too
				select {
				case <-ctx.Done():
					log.Println("Scheduler interrupted during job processing.")
					return
				default:
				}

				msg := domain.JobMessage{
					JobId:        j.Id,
					Payload:      j.Payload,
					ScheduleTime: j.ScheduleTime.Unix(),
				}

				body, err := json.Marshal(msg)
				if err != nil {
					log.Println("marshal error:", err)
					continue
				}

				err = s.jobQueue.Publish(body)
				if err != nil {
					log.Println("publish error:", err)
					continue
				}

				err = s.jobRepository.MarkQueued(j.Id)
				if err != nil {
					log.Println("mark queued error:", err)
				}
			}
		}
	}
}
