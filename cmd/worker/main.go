package main

import (
	"context"
	"job-scheduler/config"
	"job-scheduler/internal/infrastructure/db"
	"job-scheduler/internal/infrastructure/rabbitmq"
	"job-scheduler/internal/infrastructure/repository"
	"job-scheduler/internal/logger"
	"job-scheduler/internal/usecase"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}
}

func main() {
	cf := config.Load()

	logr := logger.NewLogger()

	database, err := db.NewDb(cf.DBConn)
	if err != nil {
		panic(err)
	}

	rmq, err := rabbitmq.New(cf.RabbitURL)
	if err != nil {
		panic(err)
	}
	defer rmq.Close()

	jobExecutionRepository := repository.NewJobExecutionRepository(database)

	workerUsecase := usecase.NewWorkerUsecase(jobExecutionRepository, rmq, cf.NodeID, logr)

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

	workerUsecase.Start(ctx)

	log.Println("Scheduler shut down")
}
