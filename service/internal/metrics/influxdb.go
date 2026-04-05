package metrics

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"

	"github.com/chris72205/ai-usage-tracker/service/internal/config"
	"github.com/chris72205/ai-usage-tracker/service/internal/model"
)

type InfluxDB struct {
	client   influxdb2.Client
	writeAPI api.WriteAPIBlocking
}

func NewInfluxDB(cfg config.Config) *InfluxDB {
	client := influxdb2.NewClientWithOptions(
		cfg.InfluxURL,
		cfg.InfluxToken,
		influxdb2.DefaultOptions().SetHTTPRequestTimeout(5),
	)
	return &InfluxDB{
		client:   client,
		writeAPI: client.WriteAPIBlocking(cfg.InfluxOrg, cfg.InfluxBucket),
	}
}

func (i *InfluxDB) Write(ctx context.Context, payload *model.UsagePayload) error {
	t, err := time.Parse(time.RFC3339Nano, payload.CapturedAt)
	if err != nil {
		t = time.Now()
	}

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
		if w.usage == nil || w.usage.UtilizationPct == nil {
			continue
		}
		p := influxdb2.NewPoint(
			"claude_usage",
			map[string]string{"window": w.name},
			map[string]interface{}{"utilization_pct": *w.usage.UtilizationPct},
			t,
		)
		if err := i.writeAPI.WritePoint(ctx, p); err != nil {
			return fmt.Errorf("write %s: %w", w.name, err)
		}
	}

	if payload.ExtraUsage != nil {
		p := influxdb2.NewPoint(
			"claude_extra_usage",
			nil,
			map[string]interface{}{
				"used_credits":  payload.ExtraUsage.UsedCredits,
				"monthly_limit": payload.ExtraUsage.MonthlyLimit,
				"is_enabled":    payload.ExtraUsage.IsEnabled,
			},
			t,
		)
		if err := i.writeAPI.WritePoint(ctx, p); err != nil {
			return fmt.Errorf("write extra_usage: %w", err)
		}
	}

	return nil
}

func (i *InfluxDB) Close() {
	i.client.Close()
}
