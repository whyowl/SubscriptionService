package api

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"net/http"
	"subservice/internal/api/handler"
	apimw "subservice/internal/api/middleware"
	"subservice/internal/service"
)

type Router struct {
	r *chi.Mux
	s *http.Server
}

func SetupRouter(s *service.SubscriptionService, l *zap.Logger) *Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(apimw.WithLogger(l))
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	h := handler.NewHandler(s)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/subscriptions", h.Subscribe)
		r.Get("/subscriptions/{userId}", h.GetSubscriptions)
		r.Put("/subscriptions", h.UpdateSubscription)
		r.Delete("/subscriptions", h.Unsubscribe)
		r.Get("/subscriptions", h.GetSubscription)
		r.Get("/subscriptions/summary", h.GetSubscriptionSummary)
	})

	return &Router{r: r}
}

func (router *Router) Run(addr string) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: router.r,
	}
	router.s = srv

	return router.s.ListenAndServe()
}

func (router *Router) Stop(ctx context.Context) error {
	err := router.s.Shutdown(ctx)
	if err != nil {
		return err
	}
	return nil
}
