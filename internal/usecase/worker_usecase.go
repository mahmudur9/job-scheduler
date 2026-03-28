package usecase

import (
	"context"
	"encoding/json"
	"job-scheduler/internal/domain"
	"job-scheduler/internal/repository"
	"log"

	"github.com/rabbitmq/amqp091-go"
)

type WorkerUsecase struct {
	jobExecutionRepository repository.JobExecutionRepository
	rmqChannel             *amqp091.Channel
	workerId               string
}

func NewWorkerUsecase(jobExecutionRepository repository.JobExecutionRepository, rmqChannel *amqp091.Channel, workerId string) *WorkerUsecase {
	return &WorkerUsecase{
		jobExecutionRepository,
		rmqChannel,
		workerId,
	}
}

func (w *WorkerUsecase) Start(ctx context.Context) {
	messages, err := w.rmqChannel.Consume("jobs", "", false, false, false, false, nil)
	if err != nil {
		log.Println("failed to start consuming:", err)
		return
	}

	log.Println("Worker started.")

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker stopping gracefully...")
			return

		case msg, ok := <-messages:
			if !ok {
				log.Println("RabbitMQ channel closed. Worker stopping...")
				return
			}

			err := w.handle(msg.Body)
			if err != nil {
				log.Println("handle error:", err)
				msg.Nack(false, true) // requeue on failure
			} else {
				msg.Ack(false)
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
