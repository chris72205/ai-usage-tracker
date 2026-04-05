const serviceUrlInput = document.getElementById('serviceUrl');
const bearerTokenInput = document.getElementById('bearerToken');
const saveButton = document.getElementById('save');
const statusEl = document.getElementById('status');

chrome.storage.local.get(['serviceUrl', 'bearerToken'], (data) => {
  if (data.serviceUrl) serviceUrlInput.value = data.serviceUrl;
  if (data.bearerToken) bearerTokenInput.value = data.bearerToken;
});

saveButton.addEventListener('click', () => {
  const serviceUrl = serviceUrlInput.value.trim();
  const bearerToken = bearerTokenInput.value.trim();
  chrome.storage.local.set({ serviceUrl, bearerToken }, () => {
    statusEl.textContent = 'Saved.';
    setTimeout(() => { statusEl.textContent = ''; }, 2000);
  });
});
