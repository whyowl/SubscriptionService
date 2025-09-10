package storage

import (
	"context"
	"github.com/google/uuid"
	"subservice/internal/model"
	"subservice/internal/storage/postgres"
	"time"
)

type Facade interface {
	Insert(ctx context.Context, subUnit model.Subscription) error
	Get(ctx context.Context, userId uuid.UUID, serviceId string) (*model.Subscription, error)
	Update(ctx context.Context, subUnit model.Subscription) error
	Delete(ctx context.Context, userId uuid.UUID, serviceId string) error
	GetList(ctx context.Context, userId uuid.UUID) (*[]model.Subscription, error)
	GetSummary(ctx context.Context, from time.Time, to time.Time, userId *uuid.UUID, serviceId *string) (int, error)
}

type StorageFacade struct {
	txManager    postgres.TransactionManager
	pgRepository postgres.ServiceRepository
}

func NewStorageFacade(txManager postgres.TransactionManager, pgRepository postgres.ServiceRepository) Facade {
	return &StorageFacade{
		txManager:    txManager,
		pgRepository: pgRepository,
	}
}

func (f *StorageFacade) Insert(ctx context.Context, subUnit model.Subscription) error {
	return f.pgRepository.InsertSubscription(ctx, subUnit)
}

func (f *StorageFacade) Get(ctx context.Context, userId uuid.UUID, serviceId string) (*model.Subscription, error) {
	return f.pgRepository.GetSubscription(ctx, userId, serviceId)
}

func (f *StorageFacade) Update(ctx context.Context, subUnit model.Subscription) error {
	return f.pgRepository.UpdateSubscription(ctx, subUnit)
}

func (f *StorageFacade) Delete(ctx context.Context, userId uuid.UUID, serviceId string) error {
	return f.pgRepository.DeleteSubscription(ctx, userId, serviceId)
}

func (f *StorageFacade) GetList(ctx context.Context, userId uuid.UUID) (*[]model.Subscription, error) {
	return f.pgRepository.GetSubscriptionsList(ctx, &userId, nil)
}

func (f *StorageFacade) GetSummary(ctx context.Context, from time.Time, to time.Time, userId *uuid.UUID, serviceId *string) (int, error) {
	return f.pgRepository.GetSubscriptionsSummary(ctx, from, to, userId, serviceId)
}
