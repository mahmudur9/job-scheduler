package main

import (
	"context"
	"errors"
	deliveryHttp "job-scheduler/internal/delivery/http"
	"job-scheduler/internal/infrastructure/db"
	"job-scheduler/internal/usecase"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	database, _ := db.NewDb("sqlserver://sa:Akash@123@localhost:1433?database=JobScheduler")

	jobRepository := db.NewJobRepository(database)

	jobUsecase := usecase.NewJobUsecase(jobRepository)

	jobHandler := deliveryHttp.NewJobHandler(jobUsecase)

	mux := http.NewServeMux()

	mux.HandleFunc("/jobs", jobHandler.Create)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
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

	// Create timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
