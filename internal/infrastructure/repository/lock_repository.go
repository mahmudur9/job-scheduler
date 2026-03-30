package repository

import (
	"context"
	"database/sql"
	"job-scheduler/internal/infrastructure/db"
)

type LockRepository struct {
	db *db.DB
}

func NewLockRepository(db *db.DB) *LockRepository {
	return &LockRepository{db}
}

func (r *LockRepository) TryAcquire(ctx context.Context, nodeId string) (bool, error) {
	res, err := r.db.Conn.ExecContext(ctx, `
		UPDATE SchedulerLock
		SET LockedAt = SYSUTCDATETIME(),
		    LockedBy = @p1
		WHERE Id = 1
		  AND (
		        LockedAt IS NULL
		        OR LockedAt < DATEADD(SECOND, -10, SYSUTCDATETIME())
		      )
	`, nodeId)

	if err != nil {
		return false, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows == 1, nil
}

func (r *LockRepository) Renew(ctx context.Context, nodeId string) error {
	res, err := r.db.Conn.ExecContext(ctx, `
		UPDATE SchedulerLock
		SET LockedAt = SYSUTCDATETIME()
		WHERE Id = 1 AND LockedBy = @p1
	`, nodeId)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	// If we lost ownership → signal caller
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *LockRepository) Release(ctx context.Context, nodeId string) error {
	_, err := r.db.Conn.ExecContext(ctx, `
		UPDATE SchedulerLock
		SET LockedAt = NULL,
		    LockedBy = NULL
		WHERE Id = 1 AND LockedBy = @p1
	`, nodeId)

	return err
}
