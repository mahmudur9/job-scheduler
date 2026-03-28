package domain

import "github.com/google/uuid"

// Message sent via RabbitMQ
type JobMessage struct {
	JobId        uuid.UUID `json:"job_id"`
	Payload      string    `json:"payload"`
	ScheduleTime int64     `json:"schedule_time"`
}
