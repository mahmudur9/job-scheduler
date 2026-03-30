package repository

import (
	"job-scheduler/internal/domain"
	db2 "job-scheduler/internal/infrastructure/db"

	"github.com/google/uuid"
)

type JobRepository struct {
	db *db2.DB
}

func NewJobRepository(db *db2.DB) *JobRepository {
	return &JobRepository{db}
}

func (j *JobRepository) CreateJob(job *domain.Job) error {
	_, err := j.db.Conn.Exec(`
		INSERT INTO Jobs (Id, Payload, ScheduleTime, Status, CreatedAt)
		VALUES (NEWID(), @p1, @p2, @p3, @p4)
	`, job.Payload, job.ScheduleTime, job.Status, job.CreatedAt)

	return err
}

func (j *JobRepository) FetchAndMarkQueued(limit int) ([]domain.Job, error) {
	rows, err := j.db.Conn.Query(`
		UPDATE TOP (@p1) Jobs
		SET Status = 'QUEUED'
		OUTPUT INSERTED.Id, INSERTED.Payload, INSERTED.ScheduleTime
		WHERE Status = 'SCHEDULED'
		  AND ScheduleTime <= SYSUTCDATETIME()
	`, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []domain.Job

	for rows.Next() {
		var j domain.Job
		err := rows.Scan(&j.Id, &j.Payload, &j.ScheduleTime)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}

	return jobs, nil
}

func (j *JobRepository) MarkScheduled(jobId uuid.UUID) error {
	_, err := j.db.Conn.Exec(`
		UPDATE Jobs
		SET Status = 'SCHEDULED'
		WHERE Id = @p1
		  AND Status = 'QUEUED'
	`, jobId[:])

	return err
}
