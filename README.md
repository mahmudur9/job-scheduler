# Job Scheduler

A distributed job scheduling system built in Go.

## Overview

This project provides a robust, distributed job scheduling architecture designed to handle job execution across multiple workers. It leverages RabbitMQ for message queuing and a database for persistent job and lock management.

## Architecture

The system is composed of several key components:

- **API (`cmd/api`)**: Exposes HTTP endpoints for managing jobs.
- **Scheduler (`cmd/scheduler`)**: Periodically schedules jobs.
- **Worker (`cmd/worker`)**: Consumes job tasks from the queue and executes them.
- **Infrastructure**: Uses RabbitMQ for communication and a SQL-based database for persistence.

## Project Structure

- `cmd/`: Entry points for the application (API, Scheduler, Worker).
- `internal/`: Core business logic, domain models, and infrastructure implementations.
- `config/`: Configuration handling.
- `tests/`: Unit and integration tests.

## Getting Started

### Prerequisites

- Go 1.26 or higher
- Docker and Docker Compose
- RabbitMQ

### Installation & Running

You can use the provided `docker-compose.yml` to spin up the required infrastructure and services.

```bash
docker-compose up --build
```

### Configuration

Configuration is managed via environment variables. Refer to the `config/` package for details on required settings.

## Development

### Running Tests

To run the project tests:

```bash
go test ./tests/...
```
