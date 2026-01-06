// Microsoft Graph Token Extractor - Popup Script

document.addEventListener('DOMContentLoaded', () => {
  loadTokens();
  
  // Set up copy buttons
  document.getElementById('copy-access').addEventListener('click', () => {
    copyToClipboard('access-token');
  });
  
  document.getElementById('copy-refresh').addEventListener('click', () => {
    copyToClipboard('refresh-token');
  });
  
  document.getElementById('copy-env-vars').addEventListener('click', () => {
    copyEnvironmentVariables();
  });
});

function loadTokens() {
  chrome.storage.local.get(['tokens'], (result) => {
    if (result.tokens) {
      displayTokens(result.tokens);
    } else {
      showEmptyState();
    }
  });
}

function displayTokens(tokens) {
  document.getElementById('empty-state').style.display = 'none';
  document.getElementById('tokens-section').style.display = 'block';
  
  // Display access token
  const accessTokenEl = document.getElementById('access-token');
  if (tokens.accessToken) {
    accessTokenEl.textContent = tokens.accessToken;
    document.getElementById('copy-access').disabled = false;
  } else {
    accessTokenEl.textContent = 'Not available';
    document.getElementById('copy-access').disabled = true;
  }
  
  // Display refresh token
  const refreshTokenEl = document.getElementById('refresh-token');
  if (tokens.refreshToken) {
    refreshTokenEl.textContent = tokens.refreshToken;
    document.getElementById('copy-refresh').disabled = false;
  } else {
    refreshTokenEl.textContent = 'Not available';
    document.getElementById('copy-refresh').disabled = true;
  }
  
  // Display token info
  const infoEl = document.getElementById('token-info');
  let info = [];
  if (tokens.expiresIn) {
    const expiresInMinutes = Math.floor(tokens.expiresIn / 60);
    info.push(`Expires in: ${expiresInMinutes} minutes`);
  }
  if (tokens.timestamp) {
    const date = new Date(tokens.timestamp);
    info.push(`Extracted: ${date.toLocaleString()}`);
  }
  infoEl.textContent = info.join(' • ');
  infoEl.className = 'status success';
}

function showEmptyState() {
  document.getElementById('empty-state').style.display = 'block';
  document.getElementById('tokens-section').style.display = 'none';
}

function copyToClipboard(elementId) {
  const element = document.getElementById(elementId);
  const text = element.textContent;
  
  if (text === 'Not available') {
    showStatus('No token available to copy', 'error');
    return;
  }
  
  navigator.clipboard.writeText(text).then(() => {
    showStatus('Copied to clipboard!', 'success');
  }).catch(err => {
    showStatus('Failed to copy: ' + err.message, 'error');
  });
}

function copyEnvironmentVariables() {
  chrome.storage.local.get(['tokens'], (result) => {
    if (!result.tokens) {
      showStatus('No tokens available', 'error');
      return;
    }
    
    const tokens = result.tokens;
    let envVars = '';
    
    if (tokens.accessToken) {
      envVars += `export MS_GRAPH_ACCESS_TOKEN="${tokens.accessToken}"\n`;
    }
    
    if (tokens.refreshToken) {
      envVars += `export MS_GRAPH_REFRESH_TOKEN="${tokens.refreshToken}"\n`;
    }
    
    if (envVars) {
      navigator.clipboard.writeText(envVars).then(() => {
        showStatus('Environment variables copied to clipboard!', 'success');
      }).catch(err => {
        showStatus('Failed to copy: ' + err.message, 'error');
      });
    } else {
      showStatus('No tokens available to copy', 'error');
    }
  });
}

function showStatus(message, type) {
  const statusEl = document.getElementById('status');
  statusEl.textContent = message;
  statusEl.className = `status ${type || ''}`;
  
  setTimeout(() => {
    statusEl.textContent = '';
    statusEl.className = 'status';
  }, 3000);
}

