# Microsoft Graph Token Extractor - Chrome Extension

A Chrome browser extension that automatically extracts access and refresh tokens from Microsoft Graph Explorer.

## Features

- Automatically monitors network requests in Graph Explorer
- Extracts access tokens and refresh tokens
- Easy-to-use popup interface to view and copy tokens
- One-click copy as environment variables for use with the Go client
- Secure local storage (tokens never leave your browser)

## Installation

### From Source

1. Open Chrome and navigate to `chrome://extensions/`
2. Enable "Developer mode" (toggle in top right)
3. Click "Load unpacked"
4. Select the `extension` directory from this repository
5. The extension icon should appear in your Chrome toolbar

## Usage

1. **Open Graph Explorer**
   - Navigate to https://developer.microsoft.com/graph/graph-explorer
   - Sign in with your Microsoft account

2. **Authenticate**
   - Click "Sign in" if prompted
   - Grant necessary permissions

3. **Make a Request**
   - Click "Run query" to make any API request
   - The extension will automatically extract tokens from the network request

4. **View and Copy Tokens**
   - Click the extension icon in your Chrome toolbar
   - View the extracted access token and refresh token
   - Click "Copy Access Token" or "Copy Refresh Token" to copy individual tokens
   - Click "Copy as Environment Variables" to get ready-to-use export commands

5. **Use with Go Client**
   ```bash
   # The extension provides these commands:
   export MS_GRAPH_ACCESS_TOKEN="your_access_token"
   export MS_GRAPH_REFRESH_TOKEN="your_refresh_token"
   ```

## How It Works

The extension uses content scripts to monitor network requests made by Graph Explorer. When it detects a token response from Microsoft's authentication endpoint, it:

1. Extracts the `access_token` and `refresh_token` from the response
2. Stores them securely in Chrome's local storage
3. Displays them in the popup interface for easy access

## Privacy & Security

- All token extraction happens locally in your browser
- Tokens are stored only in Chrome's local storage (never sent anywhere)
- The extension only monitors requests on Graph Explorer pages
- No external servers or services are involved

## Troubleshooting

**Tokens not appearing?**
- Make sure you're signed into Graph Explorer
- Try making a request (click "Run query")
- Refresh the Graph Explorer page and try again

**Extension not working?**
- Check that the extension is enabled in `chrome://extensions/`
- Make sure you're on the Graph Explorer page (https://developer.microsoft.com/graph/graph-explorer)
- Check the browser console for any errors (F12 → Console tab)

## Development

To modify the extension:

1. Make changes to the files in this directory
2. Go to `chrome://extensions/`
3. Click the refresh icon on the extension card
4. Test your changes

## Icons

The extension includes simple placeholder icons with a "G" (for Graph) on a Microsoft blue background (#0078d4). For production, you may want to create more polished icons:
- `icon16.png` - 16x16 pixels
- `icon48.png` - 48x48 pixels  
- `icon128.png` - 128x128 pixels

The current icons are valid PNG files and will work for installation and basic use.

