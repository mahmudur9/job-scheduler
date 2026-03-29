package db

import (
	"strings"

	"github.com/google/uuid"
)

type JobExecutionRepository struct {
	db *DB
}

func NewJobExecutionRepository(db *DB) *JobExecutionRepository {
	return &JobExecutionRepository{db}
}

func (j *JobExecutionRepository) TryStartExecution(execKey string, jobID uuid.UUID, workerId string) (bool, error) {
	tx, err := j.db.Conn.Begin()
	if err != nil {
		return false, err
	}

	_, err = tx.Exec(`
		INSERT INTO JobExecutions (Id, JobId, ExecutionKey, Status, WorkerId, StartedAt)
		VALUES (NEWID(), @p1, @p2, 'STARTED', @p3, SYSDATETIME())
	`, jobID[:], execKey, workerId)

	if err != nil {
		_ = tx.Rollback()

		// detect duplicate key
		if isDuplicateKeyError(err) {
			return false, nil // already being processed
		}

		return false, err
	}

	return true, tx.Commit()
}

func isDuplicateKeyError(err error) bool {
	return strings.Contains(err.Error(), "2627") ||
		strings.Contains(err.Error(), "2601")
}
