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
    ExecutionKey NVARCHAR(64) NOT NULL UNIQUE,
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