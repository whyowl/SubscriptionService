package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"go.uber.org/zap"
	apimw "subservice/internal/api/middleware"
	"subservice/internal/model"
	"subservice/internal/storage"
	"time"
)

type SubscriptionService struct {
	Repo storage.Facade
	l    *zap.Logger
}

func NewSubscriptionService(repo storage.Facade, l *zap.Logger) *SubscriptionService {
	return &SubscriptionService{
		Repo: repo,
		l:    l,
	}
}

func (ss *SubscriptionService) Subscribe(ctx context.Context, subUnit model.Subscription) error {
	l := apimw.FromContext(ctx).With(zap.String("user_id", subUnit.UserId.String()), zap.String("service_name", subUnit.ServiceName))
	if subUnit.EndDate != nil && subUnit.EndDate.Before(subUnit.StartDate) {
		l.Warn("End date is before start date", zap.Time("start_date", subUnit.StartDate), zap.Timep("end_date", subUnit.EndDate))
		return errors.New("end date cannot be before start date")
	}
	l.Info("Creating new subscription", zap.Any("subscription", subUnit))
	return ss.Repo.Insert(ctx, subUnit)
}

func (ss *SubscriptionService) GetSubscription(ctx context.Context, userId uuid.UUID, serviceName string) (*model.Subscription, error) {
	l := apimw.FromContext(ctx).With(zap.String("user_id", userId.String()), zap.String("service_name", serviceName))
	l.Info("Fetching subscription")
	return ss.Repo.Get(ctx, userId, serviceName)
}

func (ss *SubscriptionService) UpdateSubscription(ctx context.Context, subUnit model.Subscription) error {
	l := apimw.FromContext(ctx).With(zap.String("user_id", subUnit.UserId.String()), zap.String("service_name", subUnit.ServiceName))
	if subUnit.EndDate != nil && subUnit.EndDate.Before(subUnit.StartDate) {
		l.Warn("End date is before start date", zap.Time("start_date", subUnit.StartDate), zap.Timep("end_date", subUnit.EndDate))
		return errors.New("end date cannot be before start date")
	}
	l.Info("Updating subscription", zap.Any("subscription", subUnit))
	return ss.Repo.Update(ctx, subUnit)
}

func (ss *SubscriptionService) Unsubscribe(ctx context.Context, userId uuid.UUID, serviceName string) error {
	l := apimw.FromContext(ctx).With(zap.String("user_id", userId.String()), zap.String("service_name", serviceName))
	l.Info("Deleting subscription")
	return ss.Repo.Delete(ctx, userId, serviceName)
}

func (ss *SubscriptionService) ListSubscriptions(ctx context.Context, userId uuid.UUID) (*[]model.Subscription, error) {
	l := apimw.FromContext(ctx).With(zap.String("user_id", userId.String()))
	l.Info("Listing subscriptions")
	return ss.Repo.GetList(ctx, userId)
}

func (ss *SubscriptionService) GetSubscriptionSummary(ctx context.Context, from, to time.Time, userId *uuid.UUID, serviceName *string) (int, error) {
	l := apimw.FromContext(ctx)
	if userId != nil {
		l = l.With(zap.String("user_id", userId.String()))
	}
	if serviceName != nil {
		l = l.With(zap.String("service_name", *serviceName))
	}
	if from.After(to) {
		l.Warn("From date is after to date", zap.Time("from", from), zap.Time("to", to))
		return 0, errors.New("from date cannot be after to date")
	}
	l.Info("Getting subscription summary", zap.Time("from", from), zap.Time("to", to))
	return ss.Repo.GetSummary(ctx, from, to, userId, serviceName)
}
