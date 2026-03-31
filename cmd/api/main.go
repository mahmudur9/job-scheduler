package main

import (
	"context"
	"errors"
	"job-scheduler/config"
	deliveryHttp "job-scheduler/internal/delivery/http"
	"job-scheduler/internal/infrastructure/db"
	"job-scheduler/internal/infrastructure/repository"
	"job-scheduler/internal/logger"
	"job-scheduler/internal/middleware"
	"job-scheduler/internal/migrations"
	"job-scheduler/internal/usecase"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	migrations.RunMigration(cf.DBConn)
	database, _ := db.NewDb(cf.DBConn)

	jobRepository := repository.NewJobRepository(database)

	jobUsecase := usecase.NewJobUsecase(jobRepository, logr)

	jobHandler := deliveryHttp.NewJobHandler(jobUsecase)

	mux := http.NewServeMux()

	mux.HandleFunc("/jobs", jobHandler.Create)

	handlerWithMiddleware :=
		middleware.CORSMiddleware(middleware.LoggingMiddleware(logr)(
			middleware.RecoveryMiddleware(mux),
		))

	server := &http.Server{
		Addr:    ":8080",
		Handler: handlerWithMiddleware,
	}

	go func() {
		log.Println("Server started on :8080")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Listen error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
