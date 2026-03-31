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
	fetchInterval  time.Duration
	lockInterval   time.Duration
}

func NewSchedulerUsecase(jobRepository JobRepository, lockRepository LockRepository, jobQueue JobQueue, nodeId string,
	fetchInterval time.Duration, lockInterval time.Duration) *SchedulerUsecase {
	return &SchedulerUsecase{
		jobRepository,
		lockRepository,
		jobQueue,
		nodeId,
		fetchInterval,
		lockInterval,
	}
}

func (s *SchedulerUsecase) Run(ctx context.Context) {
	ticker := time.NewTicker(s.fetchInterval)
	defer ticker.Stop()

	lockTicker := time.NewTicker(s.lockInterval) // renew lock
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
			isL, err := s.acquireLock(ctx, isLeader)
			if err != nil {
				continue
			}
			isLeader = isL

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

				// Publish job to the queue
				s.publishJob(j)
			}
		}
	}
}

func (s *SchedulerUsecase) acquireLock(ctx context.Context, isLeader bool) (bool, error) {
	ok, err := s.lockRepository.TryAcquire(ctx, s.nodeId)
	if err != nil {
		log.Println("lock acquire error:", err)
		return isLeader, err
	}

	if ok && !isLeader {
		log.Println("Became leader:", s.nodeId)
		isLeader = true
		return isLeader, nil
	}

	// Renew if already leader
	if isLeader {
		err := s.lockRepository.Renew(ctx, s.nodeId)
		if err != nil {
			log.Println("lock renew error:", err)
			isLeader = false
		}
	}
	return isLeader, nil
}

func (s *SchedulerUsecase) publishJob(j domain.Job) {
	msg := domain.JobMessage{
		JobId:        j.Id,
		Payload:      j.Payload,
		ScheduleTime: j.ScheduleTime.Unix(),
	}

	body, err := json.Marshal(msg)
	if err != nil {
		log.Println("marshal error:", err)
		return
	}

	err = s.jobQueue.Publish(body)
	if err != nil {
		err = s.jobRepository.MarkScheduled(j.Id)
		if err != nil {
			log.Println("mark scheduled error:", err)
		}
	}
	log.Println("publish error:", err)
}
