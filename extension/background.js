// Microsoft Graph Token Extractor - Background Service Worker
// Uses webRequest API to intercept network requests

chrome.webRequest.onCompleted.addListener(
  function(details) {
    // Only process token endpoint responses
    if (details.url.includes('login.microsoftonline.com') && 
        details.url.includes('/token') &&
        details.method === 'POST') {
      
      console.log('Token request detected:', details.url);
      
      // Get the response body using chrome.webRequest.getResponseBody
      // Note: This requires the webRequestBlocking permission in Manifest V2
      // In Manifest V3, we need to use a different approach
      
      // For Manifest V3, we'll rely on the content script to intercept
      // But we can also try to get response headers
      if (details.responseHeaders) {
        console.log('Response headers available');
      }
    }
  },
  {
    urls: [
      "https://login.microsoftonline.com/*/oauth2/v2.0/token",
      "https://login.microsoftonline.com/common/oauth2/v2.0/token",
      "https://login.microsoftonline.com/organizations/oauth2/v2.0/token"
    ]
  },
  ["responseHeaders"]
);

// Listen for messages from content script
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === 'tokensExtracted') {
    console.log('Tokens extracted notification received');
    // Could trigger a notification or badge update here
  }
  return true;
});

console.log('Microsoft Graph Token Extractor: Background service worker loaded');

