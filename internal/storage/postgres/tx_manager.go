package postgres

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type txManagerKey struct{}

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

func (tm *TxManager) RunSerializable(ctx context.Context, fn func(ctxTx context.Context) error) error {
	options := pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	}
	return tm.beginWithRetry(ctx, options, fn)
}

func (tm *TxManager) RunReadUncommitted(ctx context.Context, fn func(ctxTx context.Context) error) error {
	options := pgx.TxOptions{
		IsoLevel:   pgx.ReadUncommitted,
		AccessMode: pgx.ReadOnly,
	}
	return tm.beginFunc(ctx, options, fn)
}

func (tm *TxManager) beginFunc(ctx context.Context, txOptions pgx.TxOptions, fn func(ctxTx context.Context) error) error {
	tx, err := tm.pool.BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	ctx = context.WithValue(ctx, txManagerKey{}, tx)
	if err := fn(ctx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (tm *TxManager) beginWithRetry(ctx context.Context, txOptions pgx.TxOptions, fn func(ctxTx context.Context) error) error {
	const maxRetries = 5
	baseDelay := 10 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		err := tm.beginFunc(ctx, txOptions, fn)
		if err == nil {
			return nil
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "40001" || pgErr.Code == "40P01" {
				jitter := time.Duration(rand.Intn(10)) * time.Millisecond
				time.Sleep(time.Duration(1<<attempt)*baseDelay + jitter)
				continue
			}
		}
		return err
	}
	return fmt.Errorf("transaction failed after retries")
}

func (tm *TxManager) GetQueryEngine(ctx context.Context) QueryEngine {
	tx, ok := ctx.Value(txManagerKey{}).(QueryEngine)
	if ok && tx != nil {
		return tx
	}
	return tm.pool
}
