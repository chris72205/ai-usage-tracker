const serviceUrlInput = document.getElementById('serviceUrl');
const bearerTokenInput = document.getElementById('bearerToken');
const debugLoggingInput = document.getElementById('debugLogging');
const saveButton = document.getElementById('save');
const statusEl = document.getElementById('status');

chrome.storage.local.get(['serviceUrl', 'bearerToken', 'debugLogging'], (data) => {
  if (data.serviceUrl) serviceUrlInput.value = data.serviceUrl;
  if (data.bearerToken) bearerTokenInput.value = data.bearerToken;
  debugLoggingInput.checked = data.debugLogging === true;
});

saveButton.addEventListener('click', () => {
  const serviceUrl = serviceUrlInput.value.trim();
  const bearerToken = bearerTokenInput.value.trim();
  const debugLogging = debugLoggingInput.checked;
  chrome.storage.local.set({ serviceUrl, bearerToken, debugLogging }, () => {
    statusEl.textContent = 'Saved.';
    setTimeout(() => { statusEl.textContent = ''; }, 2000);
  });
});
