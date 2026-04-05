package messaging

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/chris72205/ai-usage-tracker/service/internal/config"
	"github.com/chris72205/ai-usage-tracker/service/internal/model"
)

// windowMessage is the envelope published per routing key.
// Routing key format: usage.<platform>.<window>
// e.g. usage.claude.five_hour, usage.claude.seven_day_sonnet
type windowMessage struct {
	Platform       string   `json:"platform"`
	Window         string   `json:"window"`
	UtilizationPct *float64 `json:"utilizationPct"`
	ResetsAt       *string  `json:"resetsAt,omitempty"`
	CapturedAt     string   `json:"capturedAt"`
}

// extraMessage is the envelope published for extra/overflow usage.
// Routing key: usage.<platform>.extra_usage
type extraMessage struct {
	Platform     string   `json:"platform"`
	Window       string   `json:"window"`
	IsEnabled    bool     `json:"isEnabled"`
	MonthlyLimit int      `json:"monthlyLimit"`
	UsedCredits  float64  `json:"usedCredits"`
	Utilization  *float64 `json:"utilizationPct,omitempty"`
	CapturedAt   string   `json:"capturedAt"`
}

type RabbitMQ struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	exchange string
	platform string
}

func NewRabbitMQ(cfg config.Config) (*RabbitMQ, error) {
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("channel: %w", err)
	}

	// Topic exchange: consumers bind with patterns like:
	//   usage.#                    — all platforms, all windows
	//   usage.claude.#             — all Claude windows
	//   usage.*.seven_day_sonnet   — Sonnet usage across any platform
	//   usage.claude.five_hour     — exact match
	if err := ch.ExchangeDeclare(
		cfg.RabbitMQExchange,
		"topic",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,
	); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("exchange declare: %w", err)
	}

	return &RabbitMQ{
		conn:     conn,
		ch:       ch,
		exchange: cfg.RabbitMQExchange,
		platform: cfg.Platform,
	}, nil
}

// Publish fans out the payload as individual per-window messages, each with its
// own routing key so consumers can subscribe to exactly what they need.
func (r *RabbitMQ) Publish(payload *model.UsagePayload) error {
	windows := []struct {
		name  string
		usage *model.WindowUsage
	}{
		{"five_hour", payload.FiveHour},
		{"seven_day", payload.SevenDay},
		{"seven_day_oauth_apps", payload.SevenDayOauthApps},
		{"seven_day_opus", payload.SevenDayOpus},
		{"seven_day_sonnet", payload.SevenDaySonnet},
		{"seven_day_cowork", payload.SevenDayCowork},
		{"iguana_necktie", payload.IguanaNecktie},
	}

	for _, w := range windows {
		if w.usage == nil {
			continue
		}
		msg := windowMessage{
			Platform:       r.platform,
			Window:         w.name,
			UtilizationPct: w.usage.UtilizationPct,
			ResetsAt:       w.usage.ResetsAt,
			CapturedAt:     payload.CapturedAt,
		}
		if err := r.publish(fmt.Sprintf("usage.%s.%s", r.platform, w.name), msg); err != nil {
			return err
		}
	}

	if payload.ExtraUsage != nil {
		msg := extraMessage{
			Platform:     r.platform,
			Window:       "extra_usage",
			IsEnabled:    payload.ExtraUsage.IsEnabled,
			MonthlyLimit: payload.ExtraUsage.MonthlyLimit,
			UsedCredits:  payload.ExtraUsage.UsedCredits,
			Utilization:  payload.ExtraUsage.UtilizationPct,
			CapturedAt:   payload.CapturedAt,
		}
		if err := r.publish(fmt.Sprintf("usage.%s.extra_usage", r.platform), msg); err != nil {
			return err
		}
	}

	return nil
}

func (r *RabbitMQ) publish(routingKey string, v any) error {
	body, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", routingKey, err)
	}
	return r.ch.Publish(
		r.exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
}

func (r *RabbitMQ) Close() {
	if r.ch != nil {
		if err := r.ch.Close(); err != nil {
			log.Printf("rabbitmq channel close: %v", err)
		}
	}
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			log.Printf("rabbitmq connection close: %v", err)
		}
	}
}
