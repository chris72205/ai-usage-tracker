// Runs in the extension's isolated world — has access to chrome.* APIs.
// Listens for payloads posted by intercept.js (MAIN world) and forwards
// them to the collection service using credentials from storage.

window.addEventListener('message', (event) => {
  if (event.source !== window || event.data?.type !== 'AI_USAGE_PAYLOAD') return;

  const usage = event.data.payload;

  chrome.storage.local.get(['serviceUrl', 'bearerToken'], ({ serviceUrl, bearerToken }) => {
    if (!serviceUrl) return console.warn('[AI Usage] No service URL configured — open extension options to set it.');
    if (!bearerToken) return console.warn('[AI Usage] No bearer token configured — open extension options to set it.');
    if (!serviceUrl.startsWith('https://')) return console.warn('[AI Usage] Service URL must start with https://', serviceUrl);

    console.log('[AI Usage] Posting to', serviceUrl);
    fetch(serviceUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${bearerToken}`,
      },
      body: JSON.stringify(usage),
    }).then(res => {
      if (!res.ok) console.warn('[AI Usage] Service responded with', res.status, res.statusText);
      else console.log('[AI Usage] Posted successfully');
    }).catch(err => console.error('[AI Usage] Failed to post:', err));
  });
});
