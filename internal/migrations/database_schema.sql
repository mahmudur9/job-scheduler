IF NOT EXISTS (SELECT name FROM sys.databases WHERE name = N'JobScheduler')
BEGIN
    CREATE DATABASE [JobScheduler];
END;