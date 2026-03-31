package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"job-scheduler/internal/domain"
)

type WorkerUsecase struct {
	jobExecutionRepository JobExecutionRepository
	jobConsumer            JobConsumer
	workerId               string
	logger                 Logger
}

func NewWorkerUsecase(jobExecutionRepository JobExecutionRepository, jobConsumer JobConsumer, workerId string, logger Logger) *WorkerUsecase {
	return &WorkerUsecase{
		jobExecutionRepository,
		jobConsumer,
		workerId,
		logger,
	}
}

func (w *WorkerUsecase) Start(ctx context.Context) {
	messages, err := w.jobConsumer.Consume("jobs")
	if err != nil {
		w.logger.Error(err, "failed to start consuming")
		return
	}

	w.logger.Info("Worker started.")

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Worker stopping...")
			return

		case msg, ok := <-messages:
			if !ok {
				w.logger.Info("RabbitMQ channel closed. Worker stopping...")
				return
			}

			err := w.handle(msg.Body())
			if err != nil {
				w.logger.Error(err, "handle error")
				_ = msg.Nack(true) // requeue on failure
			} else {
				_ = msg.Ack()
			}
		}
	}
}

func (w *WorkerUsecase) handle(body []byte) error {
	var msg domain.JobMessage
	err := json.Unmarshal(body, &msg)
	if err != nil {
		return err
	}

	// Generate execution key
	execKey := domain.GenerateExecutionKey(msg.JobId, msg.ScheduleTime)

	// EXACTLY-ONCE CHECK
	ok, err := w.jobExecutionRepository.TryStartExecution(execKey, msg.JobId, w.workerId)
	if err != nil {
		return err
	}

	if !ok {
		// Already processed OR another worker is processing
		w.logger.Info(fmt.Sprintf("Skipping duplicate job: %s", msg.JobId))
		return nil
	}

	return nil
}
