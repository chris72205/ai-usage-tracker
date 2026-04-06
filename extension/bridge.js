// Runs in the extension's isolated world — has access to chrome.* APIs.
// Listens for payloads posted by intercept.js (MAIN world) and forwards
// them to the collection service using credentials from storage.

// Broadcast the current debugLogging setting to intercept.js (MAIN world).
const broadcastDebugLogging = () => {
  chrome.storage.local.get(['debugLogging'], ({ debugLogging }) => {
    window.postMessage({ type: 'AI_USAGE_DEBUG', debugLogging: debugLogging === true }, '*');
  });
};
broadcastDebugLogging();
chrome.storage.onChanged.addListener((changes) => {
  if ('debugLogging' in changes) broadcastDebugLogging();
});

window.addEventListener('message', (event) => {
  if (event.source !== window || event.data?.type !== 'AI_USAGE_PAYLOAD') return;

  const usage = event.data.payload;

  chrome.storage.local.get(['serviceUrl', 'bearerToken', 'debugLogging'], ({ serviceUrl, bearerToken, debugLogging }) => {
    const debug = (...args) => { if (debugLogging) console.log('[AI Usage]', ...args); };

    if (!serviceUrl) return console.warn('[AI Usage] No service URL configured — open extension options to set it.');
    if (!bearerToken) return console.warn('[AI Usage] No bearer token configured — open extension options to set it.');
    if (!serviceUrl.startsWith('https://')) return console.warn('[AI Usage] Service URL must start with https://', serviceUrl);

    debug('Posting to', serviceUrl);
    fetch(serviceUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${bearerToken}`,
      },
      body: JSON.stringify(usage),
    }).then(res => {
      if (!res.ok) console.warn('[AI Usage] Service responded with', res.status, res.statusText);
      else debug('Posted successfully');
    }).catch(err => console.error('[AI Usage] Failed to post:', err));
  });
});
