package usecase

import (
	"context"
	"encoding/json"
	"job-scheduler/internal/domain"
	"log"
)

type WorkerUsecase struct {
	jobExecutionRepository JobExecutionRepository
	jobConsumer            JobConsumer
	workerId               string
}

func NewWorkerUsecase(jobExecutionRepository JobExecutionRepository, jobConsumer JobConsumer, workerId string) *WorkerUsecase {
	return &WorkerUsecase{
		jobExecutionRepository,
		jobConsumer,
		workerId,
	}
}

func (w *WorkerUsecase) Start(ctx context.Context) {
	messages, err := w.jobConsumer.Consume("jobs")
	if err != nil {
		log.Println("failed to start consuming:", err)
		return
	}

	log.Println("Worker started.")

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker stopping...")
			return

		case msg, ok := <-messages:
			if !ok {
				log.Println("RabbitMQ channel closed. Worker stopping...")
				return
			}

			err := w.handle(msg.Body())
			if err != nil {
				log.Println("handle error:", err)
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
		log.Println("Skipping duplicate job:", msg.JobId)
		return nil
	}

	return nil
}
