package requests

import "time"

type JobRequest struct {
	Payload      string
	ScheduleTime time.Time
}
