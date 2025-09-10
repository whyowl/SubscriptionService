package main

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"os"
	"os/signal"
	_ "subservice/docs"
	"subservice/internal/api"
	"subservice/internal/config"
	"subservice/internal/logger"
	"subservice/internal/service"
	"subservice/internal/storage"
	"subservice/internal/storage/postgres"
	"syscall"
	"time"
)

// @title           SubService API
// @version         1.0
// @description     REST API для управления онлайн-подписками и агрегации стоимости.
// @BasePath        /api/v1
func main() {
	cfg := config.Load()

	l, cleanup := logger.New(cfg)
	defer cleanup()
	zap.ReplaceGlobals(l)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sig
		l.Info("signal received, shutting down...")
		cancel()
	}()

	pool, err := pgxpool.Connect(ctx, cfg.PostgresURL)
	if err != nil {
		l.Fatal("failed to connect to database:", zap.Error(err))
	}
	defer pool.Close()

	SubscriptionService := service.NewSubscriptionService(InitStorage(pool), l)

	router := api.SetupRouter(SubscriptionService, l)

	go func() {
		err := router.Run(cfg.ApiAddress)
		if err != nil {
			l.Fatal("failed to start server:", zap.Error(err))
		}
	}()

	<-ctx.Done()
	l.Info("shutting down server...")
	ctxSvr, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := router.Stop(ctxSvr); err != nil {
		l.Error("failed to gracefully shutdown server:", zap.Error(err))
	}
	time.Sleep(7 * time.Second)
}

func InitStorage(pool *pgxpool.Pool) storage.Facade {
	txMngr := postgres.NewTxManager(pool)
	pgRepo := postgres.NewPgRepository(txMngr)

	return storage.NewStorageFacade(txMngr, pgRepo)
}
