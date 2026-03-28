package db

import (
	"job-scheduler/internal/domain"
)

type JobRepository struct {
	db *DB
}

func NewJobRepository(db *DB) *JobRepository {
	return &JobRepository{db}
}

func (j *JobRepository) CreateJob(job *domain.Job) error {
	_, err := j.db.Conn.Exec(`
		INSERT INTO Jobs (Id, Payload, ScheduleTime, Status, CreatedAt)
		VALUES (NEWID(), @p1, @p2, @p3, @p4)
	`, job.Payload, job.ScheduleTime, job.Status, job.CreatedAt)

	return err
}
