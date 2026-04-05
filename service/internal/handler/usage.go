package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/chris72205/ai-usage-tracker/service/internal/dedup"
	"github.com/chris72205/ai-usage-tracker/service/internal/messaging"
	"github.com/chris72205/ai-usage-tracker/service/internal/metrics"
	"github.com/chris72205/ai-usage-tracker/service/internal/model"
)

const maxBodyBytes = 32 * 1024 // 32KB — well above any realistic payload size

type UsageHandler struct {
	dedup  *dedup.Redis
	rmq    *messaging.RabbitMQ
	influx *metrics.InfluxDB
}

func NewUsageHandler(d *dedup.Redis, rmq *messaging.RabbitMQ, influx *metrics.InfluxDB) *UsageHandler {
	return &UsageHandler{dedup: d, rmq: rmq, influx: influx}
}

func (h *UsageHandler) Handle(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var payload model.UsagePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	allowed, err := h.dedup.Allow(r.Context(), payload.Platform)
	if err != nil {
		// Redis is unavailable — log and allow through rather than silently dropping data.
		log.Printf("dedup check failed: %v — allowing payload", err)
		allowed = true
	}
	if !allowed {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := h.influx.Write(r.Context(), &payload); err != nil {
		log.Printf("influxdb write: %v", err)
	}

	if err := h.rmq.Publish(&payload); err != nil {
		log.Printf("rabbitmq publish: %v", err)
	}

	w.WriteHeader(http.StatusNoContent)
}
