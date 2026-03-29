package db

import (
	"job-scheduler/internal/domain"

	"github.com/google/uuid"
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

func (j *JobRepository) FetchDueJobs(limit int) ([]domain.Job, error) {
	rows, err := j.db.Conn.Query(`
		SELECT TOP (@p1) Id, Payload, ScheduleTime
		FROM Jobs
		WHERE Status = 'SCHEDULED'
		  AND ScheduleTime <= SYSDATETIME()
		ORDER BY ScheduleTime ASC
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

func (j *JobRepository) MarkQueued(jobId uuid.UUID) error {
	_, err := j.db.Conn.Exec(`
		UPDATE Jobs
		SET Status = 'QUEUED'
		WHERE Id = @p1
		  AND Status = 'SCHEDULED'
	`, jobId)

	return err
}
