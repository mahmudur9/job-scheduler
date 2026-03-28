package main

import (
	"context"
	"job-scheduler/internal/infrastructure/db"
	"job-scheduler/internal/infrastructure/rabbitmq"
	"job-scheduler/internal/usecase"
	"log"
	"os"
	"os/signal"
	"syscall"
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

	jobRepository := db.NewJobRepository(database)

	s := usecase.NewSchedulerUsecase(jobRepository, rmq, "node1")

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-quit
		log.Printf("Received signal: %v. Shutting down...", sig)
		cancel()
	}()

	s.Run(ctx)

	log.Println("Scheduler shut down")
}
