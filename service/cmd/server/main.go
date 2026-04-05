package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/chris72205/ai-usage-tracker/service/internal/config"
	"github.com/chris72205/ai-usage-tracker/service/internal/dedup"
	"github.com/chris72205/ai-usage-tracker/service/internal/handler"
	"github.com/chris72205/ai-usage-tracker/service/internal/messaging"
	"github.com/chris72205/ai-usage-tracker/service/internal/metrics"
)

func main() {
	cfg := config.Load()

	d, err := dedup.NewRedis(cfg.RedisURL, cfg.ServiceName, cfg.DedupWindow)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer d.Close()

	rmq, err := messaging.NewRabbitMQ(cfg)
	if err != nil {
		log.Fatalf("rabbitmq: %v", err)
	}
	defer rmq.Close()

	influx := metrics.NewInfluxDB(cfg)
	defer influx.Close()

	h := handler.NewUsageHandler(d, rmq, influx)

	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
		MaxAge:         300,
	}))
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.With(handler.BearerAuth(cfg.APIBearerToken)).Post("/usage", h.Handle)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
	log.Println("stopped")
}
