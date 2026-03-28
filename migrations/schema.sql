CREATE TABLE Jobs
(
    Id           UNIQUEIDENTIFIER PRIMARY KEY,
    Payload      NVARCHAR(MAX),
    ScheduleTime DATETIME2,
    Status       NVARCHAR(50),
    CreatedAt    DATETIME2
);

CREATE TABLE JobExecutions
(
    Id           UNIQUEIDENTIFIER PRIMARY KEY,
    JobId        UNIQUEIDENTIFIER,
    ExecutionKey UNIQUEIDENTIFIER UNIQUE,
    Status       NVARCHAR(50),
    WorkerId     NVARCHAR(255),
    StartedAt    DATETIME2,
    FinishedAt   DATETIME2
);

CREATE TABLE SchedulerLock
(
    Id       INT PRIMARY KEY,
    LockedAt DATETIME2,
    LockedBy NVARCHAR(255)
);

-- Update 1

ALTER TABLE JobExecutions
DROP CONSTRAINT UQ__JobExecu__3B93F7FDE65D4F23;

ALTER TABLE JobExecutions
ALTER COLUMN ExecutionKey NVARCHAR(64) NOT NULL;

ALTER TABLE JobExecutions
    ADD CONSTRAINT UQ_JobExecutions_ExecutionKey UNIQUE (ExecutionKey);