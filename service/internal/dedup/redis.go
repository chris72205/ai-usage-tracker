package dedup

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client  *redis.Client
	window  time.Duration
	prefix  string
}

func NewRedis(url, servicePrefix string, window time.Duration) (*Redis, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}
	return &Redis{client: redis.NewClient(opts), window: window, prefix: servicePrefix}, nil
}

// Allow returns true if this capture should be processed, false if it should
// be discarded. It uses SET NX with a TTL so the first caller in any window
// wins atomically, across all pods.
//
// Key format: <service>:dedup:usage:<platform>
// e.g. claude-usage-svc:dedup:usage:claude
func (r *Redis) Allow(ctx context.Context, platform string) (bool, error) {
	key := fmt.Sprintf("%s:dedup:usage:%s", r.prefix, platform)
	set, err := r.client.SetNX(ctx, key, 1, r.window).Result()
	if err != nil {
		return false, fmt.Errorf("setnx: %w", err)
	}
	return set, nil
}

func (r *Redis) Close() {
	if err := r.client.Close(); err != nil {
		log.Printf("redis close: %v", err)
	}
}
