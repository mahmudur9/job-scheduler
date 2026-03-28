package domain

import (
	"time"

	"github.com/google/uuid"
)

type Job struct {
	Id           uuid.UUID
	Payload      string
	ScheduleTime time.Time
	Status       string
	CreatedAt    time.Time
}
