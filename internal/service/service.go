package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"subservice/internal/model"
	"subservice/internal/storage"
	"time"
)

type SubscriptionService struct {
	Repo storage.Facade
}

func NewSubscriptionService(repo storage.Facade) *SubscriptionService {
	return &SubscriptionService{
		Repo: repo,
	}
}

func (ss *SubscriptionService) Subscribe(ctx context.Context, subUnit model.Subscription) error {
	if subUnit.EndDate != nil && subUnit.EndDate.Before(subUnit.StartDate) {
		return errors.New("end date cannot be before start date")
	}
	return ss.Repo.Insert(ctx, subUnit)
}

func (ss *SubscriptionService) GetSubscription(ctx context.Context, userId uuid.UUID, serviceName string) (*model.Subscription, error) {
	return ss.Repo.Get(ctx, userId, serviceName)
}

func (ss *SubscriptionService) UpdateSubscription(ctx context.Context, subUnit model.Subscription) error {
	if subUnit.EndDate != nil && subUnit.EndDate.Before(subUnit.StartDate) {
		return errors.New("end date cannot be before start date")
	}
	return ss.Repo.Update(ctx, subUnit)
}

func (ss *SubscriptionService) Unsubscribe(ctx context.Context, userId uuid.UUID, serviceName string) error {
	return ss.Repo.Delete(ctx, userId, serviceName)
}

func (ss *SubscriptionService) ListSubscriptions(ctx context.Context, userId uuid.UUID) (*[]model.Subscription, error) {
	return ss.Repo.GetList(ctx, userId)
}

func (ss *SubscriptionService) GetSubscriptionSummary(ctx context.Context, from, to time.Time, userId *uuid.UUID, serviceName *string) (int, error) {
	if from.After(to) {
		return 0, errors.New("from date cannot be after to date")
	}
	return ss.Repo.GetSummary(ctx, from, to, userId, serviceName)
}
