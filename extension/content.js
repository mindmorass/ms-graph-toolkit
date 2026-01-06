// Microsoft Graph Token Extractor - Content Script
// Monitors network requests to extract tokens from Graph Explorer

(function() {
  'use strict';

  // Intercept fetch requests
  const originalFetch = window.fetch;
  window.fetch = function(...args) {
    return originalFetch.apply(this, args)
      .then(response => {
        // Clone response to read body without consuming it
        const clonedResponse = response.clone();
        
        // Check if this is a token endpoint request
        if (args[0] && typeof args[0] === 'string' && args[0].includes('login.microsoftonline.com')) {
          clonedResponse.json().then(data => {
            if (data.access_token || data.refresh_token) {
              handleTokenResponse(data);
            }
          }).catch(() => {
            // Not JSON, ignore
          });
        }
        
        return response;
      });
  };

  // Intercept XMLHttpRequest
  const originalOpen = XMLHttpRequest.prototype.open;
  const originalSend = XMLHttpRequest.prototype.send;

  XMLHttpRequest.prototype.open = function(method, url, ...rest) {
    this._url = url;
    return originalOpen.apply(this, [method, url, ...rest]);
  };

  XMLHttpRequest.prototype.send = function(...args) {
    this.addEventListener('load', function() {
      if (this._url && this._url.includes('login.microsoftonline.com')) {
        try {
          const response = JSON.parse(this.responseText);
          if (response.access_token || response.refresh_token) {
            handleTokenResponse(response);
          }
        } catch (e) {
          // Not JSON, ignore
        }
      }
    });
    return originalSend.apply(this, args);
  };

  // Handle token response
  function handleTokenResponse(data) {
    const tokens = {
      accessToken: data.access_token,
      refreshToken: data.refresh_token,
      expiresIn: data.expires_in,
      timestamp: new Date().toISOString()
    };

    // Store tokens in extension storage
    chrome.storage.local.set({ tokens: tokens }, () => {
      console.log('Tokens extracted and stored:', {
        hasAccessToken: !!tokens.accessToken,
        hasRefreshToken: !!tokens.refreshToken,
        expiresIn: tokens.expiresIn
      });
    });

    // Show notification
    showNotification('Tokens extracted! Click the extension icon to view.');
  }

  // Show notification
  function showNotification(message) {
    // Create a simple notification element
    const notification = document.createElement('div');
    notification.style.cssText = `
      position: fixed;
      top: 20px;
      right: 20px;
      background: #0078d4;
      color: white;
      padding: 12px 20px;
      border-radius: 4px;
      box-shadow: 0 2px 8px rgba(0,0,0,0.2);
      z-index: 10000;
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      font-size: 14px;
    `;
    notification.textContent = message;
    document.body.appendChild(notification);

    setTimeout(() => {
      notification.remove();
    }, 3000);
  }

  // Listen for messages from popup
  chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
    if (request.action === 'getTokens') {
      chrome.storage.local.get(['tokens'], (result) => {
        sendResponse({ tokens: result.tokens });
      });
      return true; // Keep channel open for async response
    }
  });

  console.log('Microsoft Graph Token Extractor: Content script loaded');
})();

