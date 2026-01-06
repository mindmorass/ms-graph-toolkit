// Microsoft Graph Token Extractor - Content Script
// Monitors network requests to extract tokens from Graph Explorer

(function() {
  'use strict';

  console.log('Microsoft Graph Token Extractor: Content script loaded');

  // Intercept fetch requests
  const originalFetch = window.fetch;
  window.fetch = function(...args) {
    const url = args[0];
    const urlString = typeof url === 'string' ? url : (url instanceof Request ? url.url : '');
    
    console.log('Fetch intercepted:', urlString);
    
    return originalFetch.apply(this, args)
      .then(response => {
        // Clone response to read body without consuming it
        const clonedResponse = response.clone();
        
        // Check if this is a token endpoint request
        if (urlString && urlString.includes('login.microsoftonline.com') && urlString.includes('/token')) {
          console.log('Token endpoint detected, reading response...');
          clonedResponse.json().then(data => {
            console.log('Token response data:', { 
              hasAccessToken: !!data.access_token, 
              hasRefreshToken: !!data.refresh_token 
            });
            if (data.access_token || data.refresh_token) {
              handleTokenResponse(data);
            }
          }).catch(err => {
            console.log('Failed to parse token response as JSON:', err);
          });
        }
        
        return response;
      })
      .catch(err => {
        console.log('Fetch error:', err);
        throw err;
      });
  };

  // Intercept XMLHttpRequest
  const originalOpen = XMLHttpRequest.prototype.open;
  const originalSend = XMLHttpRequest.prototype.send;

  XMLHttpRequest.prototype.open = function(method, url, ...rest) {
    this._url = url;
    this._method = method;
    console.log('XHR open:', method, url);
    return originalOpen.apply(this, [method, url, ...rest]);
  };

  XMLHttpRequest.prototype.send = function(...args) {
    const xhr = this;
    xhr.addEventListener('load', function() {
      if (xhr._url && xhr._url.includes('login.microsoftonline.com') && xhr._url.includes('/token')) {
        console.log('XHR token endpoint response detected');
        try {
          const response = JSON.parse(xhr.responseText);
          console.log('XHR token response:', { 
            hasAccessToken: !!response.access_token, 
            hasRefreshToken: !!response.refresh_token 
          });
          if (response.access_token || response.refresh_token) {
            handleTokenResponse(response);
          }
        } catch (e) {
          console.log('Failed to parse XHR response as JSON:', e);
        }
      }
    });
    return originalSend.apply(this, args);
  };

  // Handle token response
  function handleTokenResponse(data) {
    console.log('handleTokenResponse called with:', {
      hasAccessToken: !!data.access_token,
      hasRefreshToken: !!data.refresh_token,
      expiresIn: data.expires_in
    });

    const tokens = {
      accessToken: data.access_token,
      refreshToken: data.refresh_token,
      expiresIn: data.expires_in,
      timestamp: new Date().toISOString()
    };

    // Store tokens in extension storage
    chrome.storage.local.set({ tokens: tokens }, () => {
      if (chrome.runtime.lastError) {
        console.error('Error storing tokens:', chrome.runtime.lastError);
      } else {
        console.log('Tokens extracted and stored:', {
          hasAccessToken: !!tokens.accessToken,
          hasRefreshToken: !!tokens.refreshToken,
          expiresIn: tokens.expiresIn
        });
        
        // Notify background script
        chrome.runtime.sendMessage({ action: 'tokensExtracted' }, (response) => {
          if (chrome.runtime.lastError) {
            console.log('Background script not available:', chrome.runtime.lastError);
          }
        });
      }
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
  
  // Also try to monitor for token in localStorage/sessionStorage
  // Graph Explorer might store tokens there
  function checkStorageForTokens() {
    try {
      // Check if Graph Explorer stores tokens in a known location
      // This is a fallback if network interception doesn't work
      console.log('Checking for tokens in storage...');
    } catch (e) {
      console.log('Error checking storage:', e);
    }
  }
  
  // Run check after page loads
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', checkStorageForTokens);
  } else {
    checkStorageForTokens();
  }
})();

