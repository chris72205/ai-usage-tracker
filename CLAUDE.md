# AI Usage Tracker — Session Context

## What this is

A self-hosted system for tracking AI platform usage limits across multiple browsers and platforms. It consists of a browser extension that intercepts usage API responses and a Go collection service that deduplicates, stores, and distributes the data.

## Why it exists

- Claude Code usage limits are only shown on `https://claude.ai/settings/usage`
- The underlying API endpoint (`/api/organizations/:org_id/usage`) is blocked by Cloudflare when called directly
- The extension intercepts the request from within the browser, where session cookies are already valid
- Multiple browsers may be running simultaneously — the service deduplicates captures across them

## Repository structure

```
├── extension/     Chrome/Firefox browser extension
├── service/       Go collection service (k3s / Docker)
└── .github/
    └── workflows/
        └── ai-usage-extension-release.yml   Packages and releases the extension on push to main
```

## Extension (`extension/`)

- `manifest.json` — MV3, compatible with Chrome and Firefox 109+ (no permissions required)
- `intercept.js` — Wraps `window.fetch` in the `MAIN` world to intercept usage API responses
  - Platform inferred from `location.hostname` via `PLATFORM_MAP`
  - Normalises raw API fields to camelCase, converts `utilization` to `utilizationPct`
  - Logs payload as `[AI Usage]` in the console
  - POSTs to the collection service (fetch call is currently commented out pending service deployment)
- Currently matches `https://claude.ai/*` only — add entries to `PLATFORM_MAP` and `manifest.json` matches for additional platforms

## Service (`service/`)

A lightweight Go API designed for k3s on Raspberry Pi 3 nodes (ARM64/ARM32).

- **`POST /usage`** — receives payloads from the extension
- **`GET /healthz`** — liveness probe
- **Deduplication** — Redis `SET NX` with TTL (default 30s), keyed as `<SERVICE_NAME>:dedup:usage:<platform>`. If Redis is unreachable, the payload is allowed through rather than dropped.
- **InfluxDB** — writes one point per usage window (`claude_usage` measurement, `window` tag) and one for extra usage (`claude_extra_usage`)
- **RabbitMQ** — publishes one message per window to a topic exchange (`usage`) with routing key `usage.<platform>.<window>` (e.g. `usage.claude.five_hour`)

### Key config defaults

| Variable | Default |
|---|---|
| `SERVICE_NAME` | `ai-usage-svc` |
| `RABBITMQ_EXCHANGE` | `usage` |
| `INFLUX_BUCKET` | `ai-agent-usages` |
| `DEDUP_WINDOW_SECONDS` | `30` |

## Current state

- Extension intercepts and normalises the Claude usage API response correctly
- Service is scaffolded and ready — needs `go mod tidy` to resolve dependencies before building
- The `fetch(...)` call in `intercept.js` is commented out — uncomment and set the service URL once deployed
- No popup UI

## Technical notes

- Extension uses `"world": "MAIN"` so `window.fetch` can be wrapped directly without `<script>` tag injection
- The regex `/\/api\/organizations\/[^/]+\/usage/` matches regardless of org ID
- `platform` field on the payload is set by the extension from `location.hostname`, not by the service config — the service's `PLATFORM` env var is only used as a fallback default for the Redis dedup key
- RabbitMQ fanout is per-window (not one message per full payload) so downstream consumers can bind to exactly the windows they care about
