package postgres

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"subservice/internal/model"
	"time"
)

type ServiceRepository interface {
	InsertSubscription(ctx context.Context, subUnit model.Subscription) error
	GetSubscription(ctx context.Context, userId uuid.UUID, serviceName string) (*model.Subscription, error)
	UpdateSubscription(ctx context.Context, subUnit model.Subscription) error
	DeleteSubscription(ctx context.Context, userId uuid.UUID, serviceName string) error
	GetSubscriptionsList(ctx context.Context, userId *uuid.UUID, serviceName *string) (*[]model.Subscription, error)
	GetSubscriptionsSummary(ctx context.Context, from time.Time, to time.Time, userId *uuid.UUID, serviceName *string) (int, error)
}

type QueryEngine interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)

	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)

	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

type TransactionManager interface {
	GetQueryEngine(ctx context.Context) QueryEngine
	RunReadUncommitted(ctx context.Context, fn func(ctxTx context.Context) error) error
	RunSerializable(ctx context.Context, fn func(ctxTx context.Context) error) error
}
