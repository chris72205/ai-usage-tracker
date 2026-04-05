# AI Usage Tracker — Service

A lightweight Go API that receives usage payloads from the browser extension, deduplicates them across pods via Redis, writes time-series metrics to InfluxDB, and publishes per-window messages to a RabbitMQ topic exchange.

Designed to run in k3s on Raspberry Pi 3 nodes (ARM64/ARM32).

## How it works

```
POST /usage
  │
  ├── Decode payload
  ├── Dedup check (Redis SET NX, keyed by platform, TTL = DEDUP_WINDOW_SECONDS)
  │     └── duplicate? → 204, discard
  │
  ├── Write to InfluxDB
  │     ├── measurement: claude_usage  (one point per window, tag: window=<name>)
  │     └── measurement: claude_extra_usage
  │
  └── Publish to RabbitMQ topic exchange
        └── one message per window, routing key: usage.<platform>.<window>
              e.g. usage.claude.five_hour
                   usage.claude.seven_day_sonnet
                   usage.claude.extra_usage
              (downstream consumers bind their own queues — dashboards, displays, alerts)
```

## RabbitMQ routing key scheme

Consumers bind to the `usage` topic exchange with patterns:

| Pattern | Receives |
|---|---|
| `usage.#` | Everything, all platforms |
| `usage.claude.#` | All Claude windows |
| `usage.*.five_hour` | Five-hour window across all platforms |
| `usage.claude.seven_day_sonnet` | Exact match |

## Project structure

```
service/
├── cmd/server/main.go              Entry point, wires dependencies, graceful shutdown
├── internal/
│   ├── config/config.go            Env-based configuration
│   ├── model/usage.go              Shared payload types
│   ├── handler/usage.go            POST /usage handler
│   ├── dedup/redis.go              Cross-pod deduplication via Redis SET NX
│   ├── messaging/rabbitmq.go       RabbitMQ topic exchange publisher
│   └── metrics/influxdb.go         InfluxDB time-series writer
├── Dockerfile
└── .env.example
```

## Configuration

Copy `.env.example` and fill in your values:

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP listen port |
| `SERVICE_NAME` | `ai-usage-svc` | Redis key prefix (avoids collisions on shared instances) |
| `PLATFORM` | `claude` | Source platform identifier |
| `REDIS_URL` | `redis://localhost:6379/0` | Redis connection URL |
| `DEDUP_WINDOW_SECONDS` | `30` | Drop captures arriving within this many seconds of the last accepted one |
| `RABBITMQ_URL` | `amqp://guest:guest@localhost:5672/` | RabbitMQ connection URL |
| `RABBITMQ_EXCHANGE` | `usage` | Topic exchange name |
| `INFLUX_URL` | `http://localhost:8086` | InfluxDB base URL |
| `INFLUX_TOKEN` | | InfluxDB auth token |
| `INFLUX_ORG` | | InfluxDB organisation |
| `INFLUX_BUCKET` | `ai-agent-usages` | InfluxDB bucket |

## Running locally

```sh
go mod tidy
go run ./cmd/server
```

## Building for Raspberry Pi

**64-bit OS (recommended):**
```sh
docker buildx build --platform linux/arm64 -t claude-usage-svc .
```

**32-bit OS:**
```sh
docker buildx build --platform linux/arm/v7 \
  --build-arg TARGETARCH=arm --build-arg GOARM=7 \
  -t claude-usage-svc .
```

## Endpoints

| Method | Path | Description |
|---|---|---|
| `POST` | `/usage` | Receive a usage payload from the extension |
| `GET` | `/healthz` | Liveness probe for k3s |
