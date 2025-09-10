package handler

import (
	"encoding/json"
	"net/http"
	"subservice/internal/service"
)

type RestHandler struct {
	s *service.SubscriptionService
}

func NewHandler(svc *service.SubscriptionService) *RestHandler {
	return &RestHandler{
		s: svc,
	}
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
