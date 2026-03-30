IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'Jobs')
BEGIN
CREATE TABLE Jobs
(
    Id           UNIQUEIDENTIFIER PRIMARY KEY,
    Payload      NVARCHAR(MAX),
    ScheduleTime DATETIME2,
    Status       NVARCHAR(50),
    CreatedAt    DATETIME2
);
END;

-- Create JobExecutions table if not exists
IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'JobExecutions')
BEGIN
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
END;

-- Create SchedulerLock table if not exists
IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'SchedulerLock')
BEGIN
CREATE TABLE SchedulerLock
(
    Id       INT PRIMARY KEY,
    LockedAt DATETIME2,
    LockedBy NVARCHAR(255)
);
END;

IF NOT EXISTS (SELECT 1 FROM SchedulerLock WHERE Id = 1)
BEGIN
INSERT INTO SchedulerLock (Id, LockedAt, LockedBy)
VALUES (1, NULL, NULL)
END