package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"job-scheduler/internal/domain"
	"time"
)

type SchedulerUsecase struct {
	jobRepository  JobRepository
	lockRepository LockRepository
	jobQueue       JobQueue
	nodeId         string
	fetchInterval  time.Duration
	lockInterval   time.Duration
	logger         Logger
}

func NewSchedulerUsecase(jobRepository JobRepository, lockRepository LockRepository, jobQueue JobQueue, nodeId string,
	fetchInterval time.Duration, lockInterval time.Duration, logger Logger) *SchedulerUsecase {
	return &SchedulerUsecase{
		jobRepository,
		lockRepository,
		jobQueue,
		nodeId,
		fetchInterval,
		lockInterval,
		logger,
	}
}

func (s *SchedulerUsecase) Run(ctx context.Context) {
	ticker := time.NewTicker(s.fetchInterval)
	defer ticker.Stop()

	lockTicker := time.NewTicker(s.lockInterval) // renew lock
	defer lockTicker.Stop()

	isLeader := false

	s.logger.Info("Scheduler started.")

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Scheduler stopping...")
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
				s.logger.Error(err, "fetch jobs error")
				continue
			}

			for _, j := range jobs {
				// Check shutdown signal between jobs too
				select {
				case <-ctx.Done():
					s.logger.Info("Scheduler interrupted during job processing.")
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
		s.logger.Error(err, "lock acquire error")
		return isLeader, err
	}

	if ok && !isLeader {
		s.logger.Info("Became leader:" + s.nodeId)
		isLeader = true
		return isLeader, nil
	}

	// Renew if already leader
	if isLeader {
		err := s.lockRepository.Renew(ctx, s.nodeId)
		if err != nil {
			s.logger.Error(err, "lock renew error")
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
		s.logger.Error(err, "marshal error")
		return
	}

	if pubErr := s.jobQueue.Publish(body, "jobs"); pubErr != nil {
		s.logger.Error(pubErr, "publish error")
		// Restore to the previous state like a rollback
		if restoreErr := s.jobRepository.RestoreToScheduled(j.Id); restoreErr != nil {
			s.logger.Error(restoreErr, "mark scheduled error")
			return
		}
	}

	s.logger.Info(fmt.Sprintf("job published successfully with jobId: %s", j.Id))
}
