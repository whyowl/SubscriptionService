package handler

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
	"subservice/internal/model"
	"time"

	"github.com/google/uuid"
)

type SubscriptionRequest struct {
	ServiceName string `json:"service_name"`
	Price       int64  `json:"price"`
	UserId      string `json:"user_id"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date,omitempty"`
}

type RequestError struct {
	Message    string
	StatusCode int
}

func (h *RestHandler) Subscribe(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var parsedReq = model.Subscription{}

	var req SubscriptionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if reqErr, sub := ValidateSubscriptionRequest(&req); reqErr != nil {
		respondError(w, reqErr.StatusCode, reqErr.Message)
		return
	} else {
		parsedReq = *sub
	}

	if err := h.s.Subscribe(ctx, parsedReq); err != nil {
		if err.Error() == "subscription already exists" {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"status": "success"})
}

func (h *RestHandler) GetSubscriptions(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	userIdStr := chi.URLParam(r, "userId")
	userId, err := uuid.Parse(userIdStr)
	if err != nil || userId == uuid.Nil {
		respondError(w, http.StatusBadRequest, "invalid userId parameter")
		return
	}

	subs, err := h.s.ListSubscriptions(ctx, userId)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, subs)
}

func (h *RestHandler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var parsedReq = model.Subscription{}

	var req SubscriptionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if reqErr, sub := ValidateSubscriptionRequest(&req); reqErr != nil {
		respondError(w, reqErr.StatusCode, reqErr.Message)
		return
	} else {
		parsedReq = *sub
	}

	if err := h.s.UpdateSubscription(ctx, parsedReq); err != nil {
		if err.Error() == "subscription not found" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *RestHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	userIdStr := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")

	userId, err := uuid.Parse(userIdStr)
	if err != nil || userId == uuid.Nil {
		respondError(w, http.StatusBadRequest, "invalid user_id parameter")
		return
	}

	if serviceName == "" {
		respondError(w, http.StatusBadRequest, "service_name is required")
		return
	}

	if err := h.s.Unsubscribe(ctx, userId, serviceName); err != nil {
		if err.Error() == "subscription not found" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *RestHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	userIdStr := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")

	userId, err := uuid.Parse(userIdStr)
	if err != nil || userId == uuid.Nil {
		respondError(w, http.StatusBadRequest, "invalid user_id parameter")
		return
	}

	if serviceName == "" {
		respondError(w, http.StatusBadRequest, "service_name is required")
		return
	}

	sub, err := h.s.GetSubscription(ctx, userId, serviceName)
	if err != nil {
		if err.Error() == "subscription not found" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, sub)
}

func (h *RestHandler) GetSubscriptionSummary(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	userIdStr := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")

	if fromStr == "" || toStr == "" {
		respondError(w, http.StatusBadRequest, "from and to parameters are required")
		return
	}

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid from date format")
		return
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid to date format")
		return
	}

	var userId *uuid.UUID
	if userIdStr != "" {
		uid, err := uuid.Parse(userIdStr)
		if err != nil || uid == uuid.Nil {
			respondError(w, http.StatusBadRequest, "invalid user_id parameter")
			return
		}
		userId = &uid
	}

	var svcName *string
	if serviceName != "" {
		svcName = &serviceName
	}

	summary, err := h.s.GetSubscriptionSummary(ctx, from, to, userId, svcName)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]int{"total_subscriptions": summary})
}

func ValidateSubscriptionRequest(req *SubscriptionRequest) (*RequestError, *model.Subscription) {
	var parsedReq = model.Subscription{}
	var err error

	parsedReq.UserId, err = uuid.Parse(req.UserId)
	if err != nil || parsedReq.UserId == uuid.Nil {
		return &RequestError{Message: "invalid user_id parameter", StatusCode: http.StatusBadRequest}, nil
	}

	if req.ServiceName == "" {
		return &RequestError{Message: "service_name is required", StatusCode: http.StatusBadRequest}, nil
	}
	parsedReq.ServiceName = req.ServiceName

	if req.Price < 0 {
		return &RequestError{Message: "price cannot be negative", StatusCode: http.StatusBadRequest}, nil
	}
	parsedReq.Price = req.Price

	if parsedReq.StartDate, err = time.Parse(time.RFC3339, req.StartDate); err != nil {
		return &RequestError{Message: "invalid start_date format", StatusCode: http.StatusBadRequest}, nil
	}

	if req.EndDate != "" {
		if *parsedReq.EndDate, err = time.Parse(time.RFC3339, req.EndDate); err != nil {
			return &RequestError{Message: "invalid end_date format", StatusCode: http.StatusBadRequest}, nil
		}
	}
	return nil, &parsedReq
}
