# AI Usage Tracker — Extension

A Chrome/Firefox browser extension that intercepts AI platform usage API responses and forwards them to the collection service.

## How it works

The content script (`intercept.js`) runs at `document_start` in the `MAIN` world — the same JS context as the page — so it can wrap `window.fetch` directly. When a request matching `/api/organizations/<org_id>/usage` completes, the response is cloned, parsed into a normalised object, and POSTed to the collection service.

The platform is inferred from `location.hostname` at script load, so the same extension handles multiple platforms without any per-site configuration.

## Files

- `manifest.json` — MV3 manifest, compatible with Chrome and Firefox 109+
- `intercept.js` — Content script: intercepts fetch, normalises payload, forwards to service

## Supported platforms

| Hostname | Platform value |
|---|---|
| `claude.ai` | `claude` |

Add entries to `PLATFORM_MAP` in `intercept.js` and the corresponding `matches` pattern in `manifest.json` to support additional platforms.

## Installation (without publishing to a store)

### Chrome

1. Go to `chrome://extensions`
2. Enable **Developer mode** (toggle, top-right)
3. Click **Load unpacked** and select this `extension/` directory
4. The extension activates immediately

To reload after changes: click the refresh icon on the extension card.

### Firefox

Temporary add-ons are cleared on browser restart — fine for development.

1. Go to `about:debugging#/runtime/this-firefox`
2. Click **Load Temporary Add-on...** and select `manifest.json`

For a persistent install without publishing, use Firefox Developer Edition or Nightly:

1. Set `xpinstall.signatures.required` to `false` in `about:config`
2. Zip the `extension/` directory and rename it to `.xpi`
3. Install via `about:addons` → gear icon → **Install Add-on From File...**

## Configuring the service endpoint

In `intercept.js`, uncomment the `fetch(...)` call and replace the URL with your service address:

```js
fetch('https://your-service.example.com/usage', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(usage),
}).catch(() => {});
```

## Verifying it works

Visit the usage page for a supported platform (e.g. `https://claude.ai/settings/usage`). Open DevTools → Console and look for a `[AI Usage]` entry with the formatted payload.
