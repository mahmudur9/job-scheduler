package main

import (
	"job-scheduler/internal/infrastructure/db"
	"job-scheduler/internal/infrastructure/rabbitmq"
	"job-scheduler/internal/usecase"
)

func main() {
	database, err := db.NewDb("sqlserver://sa:Akash@123@localhost:1433?database=JobScheduler")
	if err != nil {
		panic(err)
	}

	rmq, err := rabbitmq.New("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}

	jobExecutionRepository := db.NewJobExecutionRepository(database)

	workerUsecase := usecase.NewWorkerUsecase(jobExecutionRepository, rmq.Channel(), "worker1")

	workerUsecase.Start()
}
