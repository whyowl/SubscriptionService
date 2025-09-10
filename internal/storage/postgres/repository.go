package postgres

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	apimw "subservice/internal/api/middleware"
	"subservice/internal/model"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type PgRepository struct {
	txManager TransactionManager
}

func NewPgRepository(txManager TransactionManager) *PgRepository {
	return &PgRepository{txManager: txManager}
}

func (r *PgRepository) InsertSubscription(ctx context.Context, subUnit model.Subscription) error {
	l := apimw.FromContext(ctx)
	subUnit.StartDate = firstOfMonth(subUnit.StartDate)
	if subUnit.EndDate != nil {
		end := firstOfMonth(*subUnit.EndDate)
		subUnit.EndDate = &end
	}
	l.Info("Updated dates for subscription", zap.Time("start_date", subUnit.StartDate), zap.Timep("end_date", subUnit.EndDate))

	tx := r.txManager.GetQueryEngine(ctx)

	query := "INSERT INTO subscriptions (user_id, service_name, price, start_date, end_date) VALUES ($1, $2, $3, $4, $5)"

	_, err := tx.Exec(ctx, query, subUnit.UserId, subUnit.ServiceName, subUnit.Price, subUnit.StartDate, subUnit.EndDate)
	if err != nil {

		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			l.Warn("Subscription already exists", zap.String("user_id", subUnit.UserId.String()), zap.String("service_name", subUnit.ServiceName))
			return errors.New("subscription already exists")
		}
		l.Error("Failed to insert subscription", zap.Error(err))
		return err
	}
	l.Info("Subscription inserted successfully", zap.String("user_id", subUnit.UserId.String()), zap.String("service_name", subUnit.ServiceName))
	return nil
}

func (r *PgRepository) GetSubscription(ctx context.Context, userId uuid.UUID, serviceName string) (*model.Subscription, error) {
	l := apimw.FromContext(ctx)

	tx := r.txManager.GetQueryEngine(ctx)

	query := `
		SELECT user_id, service_name, price, start_date, end_date
		FROM subscriptions
		WHERE user_id = $1 AND service_name = $2
	`

	row := tx.QueryRow(ctx, query, userId, serviceName)

	var sub model.Subscription
	err := row.Scan(
		&sub.UserId,
		&sub.ServiceName,
		&sub.Price,
		&sub.StartDate,
		&sub.EndDate,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			l.Warn("Subscription not found", zap.String("user_id", userId.String()), zap.String("service_name", serviceName))
			return nil, errors.New("subscription not found")
		}
		l.Error("Failed to get subscription", zap.Error(err))
		return nil, err
	}
	l.Info("Subscription fetched successfully", zap.String("user_id", userId.String()), zap.String("service_name", serviceName))
	return &sub, nil
}

func (r *PgRepository) UpdateSubscription(ctx context.Context, subUnit model.Subscription) error {
	l := apimw.FromContext(ctx)

	subUnit.StartDate = firstOfMonth(subUnit.StartDate)
	if subUnit.EndDate != nil {
		end := firstOfMonth(*subUnit.EndDate)
		subUnit.EndDate = &end
	}
	l.Info("Updated dates for subscription", zap.Time("start_date", subUnit.StartDate), zap.Timep("end_date", subUnit.EndDate))

	tx := r.txManager.GetQueryEngine(ctx)

	query := `
		UPDATE subscriptions
		SET price = $1,
		    start_date = $2,
		    end_date = $3
		WHERE user_id = $4 AND service_name = $5
	`

	cmdTag, err := tx.Exec(ctx, query,
		subUnit.Price,
		subUnit.StartDate,
		subUnit.EndDate,
		subUnit.UserId,
		subUnit.ServiceName,
	)
	if err != nil {
		l.Error("Failed to update subscription", zap.Error(err))
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		l.Warn("Subscription not found for update", zap.String("user_id", subUnit.UserId.String()), zap.String("service_name", subUnit.ServiceName))
		return errors.New("subscription not found")
	}
	l.Info("Subscription updated successfully", zap.String("user_id", subUnit.UserId.String()), zap.String("service_name", subUnit.ServiceName))
	return nil
}

func (r *PgRepository) DeleteSubscription(ctx context.Context, userId uuid.UUID, serviceName string) error {
	l := apimw.FromContext(ctx)

	tx := r.txManager.GetQueryEngine(ctx)

	query := "DELETE FROM subscriptions WHERE user_id = $1 AND service_name = $2"

	cmdTag, err := tx.Exec(ctx, query, userId, serviceName)
	if err != nil {
		l.Error("Failed to delete subscription", zap.Error(err))
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		l.Warn("Subscription not found for deletion", zap.String("user_id", userId.String()), zap.String("service_name", serviceName))
		return errors.New("subscription not found")
	}
	l.Info("Subscription deleted successfully", zap.String("user_id", userId.String()), zap.String("service_name", serviceName))
	return nil
}

func (r *PgRepository) GetSubscriptionsList(ctx context.Context, userId *uuid.UUID, serviceName *string) (*[]model.Subscription, error) {
	l := apimw.FromContext(ctx)

	tx := r.txManager.GetQueryEngine(ctx)

	query := `
		SELECT user_id, service_name, price, start_date, end_date
		FROM subscriptions
		WHERE 1=1
	`

	args := []interface{}{}
	argIdx := 1

	if userId != nil {
		l.Info("Filtering subscriptions by user_id", zap.String("user_id", userId.String()))
		query += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, *userId)
		argIdx++
	}

	if serviceName != nil {
		l.Info("Filtering subscriptions by service_name", zap.String("service_name", *serviceName))
		query += fmt.Sprintf(" AND service_name = $%d", argIdx)
		args = append(args, *serviceName)
		argIdx++
	}

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		l.Error("Failed to query subscriptions", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var subs []model.Subscription

	for rows.Next() {
		var s model.Subscription
		err := rows.Scan(&s.UserId, &s.ServiceName, &s.Price, &s.StartDate, &s.EndDate)
		if err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	l.Info("Fetched subscriptions successfully", zap.Int("count", len(subs)))
	return &subs, rows.Err()
}

func (r *PgRepository) GetSubscriptionsSummary(ctx context.Context, from time.Time, to time.Time, userId *uuid.UUID, serviceName *string) (int, error) {
	l := apimw.FromContext(ctx)

	tx := r.txManager.GetQueryEngine(ctx)

	query := `
		SELECT COALESCE(SUM(s.price), 0) AS total_price
		FROM subscriptions s
		JOIN generate_series($1::date, $2::date, interval '1 month') m
			ON m >= s.start_date
		   AND (s.end_date IS NULL OR m <= s.end_date)
		WHERE ($3::uuid IS NULL OR s.user_id = $3)
		  AND ($4::text IS NULL OR s.service_name = $4)
	`

	var total int
	err := tx.QueryRow(ctx, query, from, to, userId, serviceName).Scan(&total)
	if err != nil {
		l.Error("Failed to get subscriptions summary", zap.Error(err))
		return 0, err
	}
	l.Info("Fetched subscriptions summary successfully", zap.Int("total_price", total))
	return total, nil
}

func firstOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}
