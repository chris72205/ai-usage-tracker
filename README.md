# AI Usage Tracker

A self-hosted system for tracking AI platform usage limits in real time — captured from the browser, stored in InfluxDB, and distributed via RabbitMQ to downstream consumers (dashboards, hardware displays, alerts).

## How it works

```
Browser (claude.ai)
  └── extension/          intercepts the usage API response
        │
        └── POST /usage ──► service/      deduplicates, writes to InfluxDB,
                                           publishes to RabbitMQ topic exchange
                                                │
                                    ┌───────────┴───────────┐
                                 InfluxDB              RabbitMQ
                                 (metrics)         (downstream consumers)
```

1. **[extension/](./extension/README.md)** — A Chrome/Firefox content script that wraps `window.fetch` on supported platforms. When it intercepts a usage API response, it normalises the payload and POSTs it to the collection service.

2. **[service/](./service/README.md)** — A lightweight Go API running in k3s. It deduplicates incoming captures across pods via Redis, writes time-series data to InfluxDB, and publishes per-window messages to a RabbitMQ topic exchange so any number of downstream consumers can subscribe to exactly what they need.

## Repository structure

```
├── extension/     Browser extension (Chrome/Firefox)
└── service/       Go collection service (k3s / Docker)
```
