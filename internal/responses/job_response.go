package responses

import (
	"time"

	"github.com/google/uuid"
)

type JobResponse struct {
	Id           uuid.UUID
	Payload      string
	ScheduleTime time.Time
	Status       string
	CreatedAt    time.Time
}
