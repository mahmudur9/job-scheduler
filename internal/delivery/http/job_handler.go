package http

import (
	"encoding/json"
	"job-scheduler/internal/requests"
	"job-scheduler/internal/usecase"
	"net/http"
)

type JobHandler struct {
	jobUsecase *usecase.JobUsecase
}

func NewJobHandler(jobUsecase *usecase.JobUsecase) *JobHandler {
	return &JobHandler{
		jobUsecase,
	}
}

func (h *JobHandler) Create(w http.ResponseWriter, r *http.Request) {
	var jobRequest requests.JobRequest
	err := json.NewDecoder(r.Body).Decode(&jobRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.jobUsecase.Create(&jobRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "job created successfully",
	})
}
