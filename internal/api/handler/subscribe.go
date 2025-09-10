package handler

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	apimw "subservice/internal/api/middleware"
	"subservice/internal/model"
	"time"

	"github.com/google/uuid"
)

type SubscriptionRequest struct {
	ServiceName string `json:"service_name" example:"Yandex Plus"`
	Price       int64  `json:"price" example:"499"`
	UserId      string `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string `json:"start_date" example:"2023-10-01T00:00:00Z"`
	EndDate     string `json:"end_date,omitempty" example:"2025-10-01T00:00:00Z"`
}

type RequestError struct {
	Message    string
	StatusCode int
}

type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}

type SuccessResponse struct {
	Status string `json:"status" example:"success"`
}

type SummeryResponse struct {
	TotalPrice int `json:"total_price" example:"1497"`
}

// Subscribe godoc
// @Summary      Создать подписку
// @Description  Создает запись о подписке пользователя
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        body  body      SubscriptionRequest  true  "Данные подписки"
// @Success      201   {object}  SuccessResponse   "status: success"
// @Failure      400   {object}  ErrorResponse   "invalid json / validation error"
// @Failure      409   {object}  ErrorResponse   "subscription already exists"
// @Failure      500   {object}  ErrorResponse   "internal server error"
// @Router       /subscriptions [post]
func (h *RestHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	l := apimw.FromContext(r.Context())

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var parsedReq = model.Subscription{}

	var req SubscriptionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		l.Warn("Handler Subscribe: invalid json")
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if reqErr, sub := ValidateSubscriptionRequest(&req); reqErr != nil {
		l.Warn("Handler Subscribe: validation error", zap.String("error", reqErr.Message))
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

// GetSubscriptions godoc
// @Summary      Список подписок пользователя
// @Description  Возвращает все подписки для пользователя
// @Tags         subscriptions
// @Produce      json
// @Param        userId  path      string  true  "User ID (UUID)"
// @Success      200     {array}   model.Subscription
// @Failure      400     {object}  ErrorResponse "invalid userId parameter"
// @Failure      500     {object}  ErrorResponse "internal server error"
// @Router       /subscriptions/{userId} [get]
func (h *RestHandler) GetSubscriptions(w http.ResponseWriter, r *http.Request) {
	l := apimw.FromContext(r.Context())

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	userIdStr := chi.URLParam(r, "userId")
	userId, err := uuid.Parse(userIdStr)
	if err != nil || userId == uuid.Nil {
		l.Warn("Handler GetSubscriptions: invalid userId parameter")
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

// UpdateSubscription godoc
// @Summary      Обновить подписку
// @Description  Обновляет запись подписки (по user_id + service_name)
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        body  body      SubscriptionRequest  true  "Данные подписки"
// @Success      200   {object}  SuccessResponse "status: success"
// @Failure 400 {object} ErrorResponse "validation error"
// @Example {json} Ошибка валидации:
//
//	{
//	  "error": "invalid user_id format"
//	}
//
// @Failure      404   {object}  ErrorResponse   "subscription not found"
// @Example {json} Ошибка запроса:
//
//	{
//	  "error": "subscription not found"
//	}
//
// @Failure      500   {object}  ErrorResponse	 "internal server error"
// @Example {json} Ошибка сервера:
//
//	{
//	  "error": "internal server error"
//	}
//
// @Router       /subscriptions [put]
func (h *RestHandler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	l := apimw.FromContext(r.Context())

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var parsedReq = model.Subscription{}

	var req SubscriptionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		l.Warn("Handler UpdateSubscription: invalid json")
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if reqErr, sub := ValidateSubscriptionRequest(&req); reqErr != nil {
		l.Warn("Handler UpdateSubscription: validation error", zap.String("error", reqErr.Message))
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

// Unsubscribe godoc
// @Summary      Удалить подписку
// @Description  Удаляет запись о подписке по user_id и service_name
// @Tags         subscriptions
// @Produce      json
// @Param        user_id       query     string  true  "User ID (UUID)"
// @Param        service_name  query     string  true  "Название сервиса"
// @Success      200           {object}  SuccessResponse "status: success"
// @Failure      400           {object}  ErrorResponse "invalid user_id parameter / service_name is required"
// @Failure      404           {object}  ErrorResponse   "subscription not found"
// @Failure      500           {object}  ErrorResponse "internal server error"
// @Router       /subscriptions [delete]
func (h *RestHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	l := apimw.FromContext(r.Context())

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	userIdStr := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")

	userId, err := uuid.Parse(userIdStr)
	if err != nil || userId == uuid.Nil {
		l.Warn("Handler Unsubscribe: invalid user_id parameter", zap.String("user_id", userIdStr))
		respondError(w, http.StatusBadRequest, "invalid user_id parameter")
		return
	}

	if serviceName == "" {
		l.Warn("Handler Unsubscribe: service_name is empty")
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

// GetSubscription godoc
// @Summary      Получить подписку
// @Description  Возвращает одну подписку по user_id и service_name
// @Tags         subscriptions
// @Produce      json
// @Param        user_id       query     string  true  "User ID (UUID)"
// @Param        service_name  query     string  true  "Название сервиса"
// @Success      200           {object}  model.Subscription
// @Failure      400           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse   "subscription not found"
// @Failure      500           {object}  ErrorResponse
// @Router       /subscriptions [get]
func (h *RestHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	l := apimw.FromContext(r.Context())

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	userIdStr := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")

	userId, err := uuid.Parse(userIdStr)
	if err != nil || userId == uuid.Nil {
		l.Warn("Handler GetSubscription: invalid user_id parameter", zap.String("user_id", userIdStr))
		respondError(w, http.StatusBadRequest, "invalid user_id parameter")
		return
	}

	if serviceName == "" {
		l.Warn("Handler GetSubscription: service_name is empty")
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

// GetSubscriptionSummary godoc
// @Summary      Сумма подписок за период
// @Description  Считает суммарную стоимость активных подписок по месяцам за период, с фильтрами
// @Tags         subscriptions
// @Produce      json
// @Param        from          query     string  true  "Начало периода (RFC3339)"
// @Param        to            query     string  true  "Конец периода (RFC3339)"
// @Param        user_id       query     string  false "User ID (UUID)"
// @Param        service_name  query     string  false "Название сервиса"
// @Success      200           {object}  SummeryResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /subscriptions/summary [get]
func (h *RestHandler) GetSubscriptionSummary(w http.ResponseWriter, r *http.Request) {
	l := apimw.FromContext(r.Context())

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	userIdStr := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")

	if fromStr == "" || toStr == "" {
		l.Warn("Handler GetSubscriptionSummary: missing from or to parameters")
		respondError(w, http.StatusBadRequest, "from and to parameters are required")
		return
	}

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		l.Warn("Handler GetSubscriptionSummary: invalid from date format")
		respondError(w, http.StatusBadRequest, "invalid from date format")
		return
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		l.Warn("Handler GetSubscriptionSummary: invalid to date format")
		respondError(w, http.StatusBadRequest, "invalid to date format")
		return
	}

	var userId *uuid.UUID
	if userIdStr != "" {
		uid, err := uuid.Parse(userIdStr)
		if err != nil || uid == uuid.Nil {
			l.Warn("Handler GetSubscriptionSummary: invalid user_id parameter")
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
		l.Error("Handler GetSubscriptionSummary: internal error", zap.Error(err))
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]int{"total_price": summary})
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
		end, err := time.Parse(time.RFC3339, req.EndDate)
		if err != nil {
			return &RequestError{Message: "invalid end_date format", StatusCode: http.StatusBadRequest}, nil
		}
		parsedReq.EndDate = &end
	}
	return nil, &parsedReq
}
