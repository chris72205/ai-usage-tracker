// --- Helpers ---

const formatWindow = (w) => w === null ? null : {
  utilizationPct: w.utilization,
  resetsAt: w.resets_at,
};

// --- Per-platform parsers ---
// Each entry defines:
//   matches  — RegExp tested against the intercepted URL
//   parse    — transforms the raw API response into a normalised payload

const PLATFORMS = {
  claude: {
    matches: /\/api\/organizations\/[^/]+\/usage/,
    parse(raw) {
      return {
        platform: 'claude',
        fiveHour:          formatWindow(raw.five_hour),
        sevenDay:          formatWindow(raw.seven_day),
        sevenDayOauthApps: formatWindow(raw.seven_day_oauth_apps),
        sevenDayOpus:      formatWindow(raw.seven_day_opus),
        sevenDaySonnet:    formatWindow(raw.seven_day_sonnet),
        sevenDayCowork:    formatWindow(raw.seven_day_cowork),
        iguanaNecktie:     formatWindow(raw.iguana_necktie),
        extraUsage: raw.extra_usage === null ? null : {
          isEnabled:      raw.extra_usage.is_enabled,
          monthlyLimit:   raw.extra_usage.monthly_limit,
          usedCredits:    raw.extra_usage.used_credits,
          utilizationPct: raw.extra_usage.utilization,
        },
        capturedAt: new Date().toISOString(),
      };
    },
  },

  // Add further platforms here, e.g.:
  // copilot: {
  //   matches: /\/api\/usage/,
  //   parse(raw) { ... },
  // },
};

// --- Hostname → platform key ---

const PLATFORM_MAP = {
  'claude.ai': 'claude',
};

// --- Interceptor ---

const platform = PLATFORMS[PLATFORM_MAP[location.hostname]];

// debugLogging is synced from bridge.js (isolated world) via postMessage.
let _debugLogging = false;
window.addEventListener('message', (event) => {
  if (event.source === window && event.data?.type === 'AI_USAGE_DEBUG') {
    _debugLogging = event.data.debugLogging;
  }
});

const _fetch = window.fetch;
window.fetch = async (...args) => {
  const res = await _fetch(...args);
  if (platform) {
    const url = typeof args[0] === 'string' ? args[0] : args[0]?.url ?? '';
    if (platform.matches.test(url)) {
      res.clone().json().then(raw => {
        const usage = platform.parse(raw);
        if (_debugLogging) console.log('[AI Usage]', usage);

        // chrome.* APIs are unavailable in MAIN world — delegate to bridge.js
        // via window messaging so it can access storage and make the fetch.
        window.postMessage({ type: 'AI_USAGE_PAYLOAD', payload: usage }, '*');
      }).catch(() => {});
    }
  }
  return res;
};
