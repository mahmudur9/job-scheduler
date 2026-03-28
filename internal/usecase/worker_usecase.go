package usecase

import (
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

func (w *WorkerUsecase) Start() {
	msgs, _ := w.rmqChannel.Consume("jobs", "", false, false, false, false, nil)

	for msg := range msgs {
		err := w.handle(msg.Body)

		if err != nil {
			msg.Nack(false, true)
		} else {
			msg.Ack(false)
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
	execKey := domain.GenerateExecutionKey(msg.JobID, msg.ScheduleTime)

	// EXACTLY-ONCE CHECK
	ok, err := w.jobExecutionRepository.TryStartExecution(execKey, msg.JobID, w.workerId)
	if err != nil {
		return err
	}

	if !ok {
		// Already processed OR another worker is processing
		log.Println("Skipping duplicate job:", msg.JobID)
		return nil
	}

	return nil
}
