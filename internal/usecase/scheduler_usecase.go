package usecase

import (
	"context"
	"encoding/json"
	"job-scheduler/internal/domain"
	"log"
	"time"
)

type SchedulerUsecase struct {
	jobRepository  JobRepository
	lockRepository LockRepository
	jobQueue       JobQueue
	nodeId         string
}

func NewSchedulerUsecase(jobRepository JobRepository, lockRepository LockRepository, jobQueue JobQueue, nodeId string) *SchedulerUsecase {
	return &SchedulerUsecase{
		jobRepository,
		lockRepository,
		jobQueue,
		nodeId,
	}
}

func (s *SchedulerUsecase) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	lockTicker := time.NewTicker(5 * time.Second) // renew lock
	defer lockTicker.Stop()

	isLeader := false

	log.Println("Scheduler started.")

	for {
		select {
		case <-ctx.Done():
			log.Println("Scheduler stopping...")
			if isLeader {
				_ = s.lockRepository.Release(ctx, s.nodeId)
			}
			return

		// Try acquiring lock periodically
		case <-lockTicker.C:
			ok, err := s.lockRepository.TryAcquire(ctx, s.nodeId)
			if err != nil {
				log.Println("lock acquire error:", err)
				continue
			}

			if ok && !isLeader {
				log.Println("Became leader:", s.nodeId)
				isLeader = true
			}

			// Renew if already leader
			if isLeader {
				err := s.lockRepository.Renew(ctx, s.nodeId)
				if err != nil {
					log.Println("lock renew error:", err)
					isLeader = false
				}
			}

		case <-ticker.C:
			if !isLeader {
				continue
			}

			jobs, err := s.jobRepository.FetchAndMarkQueued(10)
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
				if err == nil {
					continue
				}
				log.Println("publish error:", err)

				err = s.jobRepository.MarkScheduled(j.Id)
				if err != nil {
					log.Println("mark scheduled error:", err)
				}
			}
		}
	}
}
