# Microsoft Graph API Tools

A monorepo containing tools for working with Microsoft Graph API, including a Go client library and a Chrome extension for token extraction.

## Repository Structure

```
ms_graph/
├── client/          # Go client library for Microsoft Graph API
│   ├── cmd/        # Example application
│   ├── internal/   # Internal packages
│   └── README.md   # Client documentation
├── extension/       # Chrome extension for token extraction
│   └── README.md   # Extension documentation
└── README.md       # This file
```

## Components

### 1. Go Client (`client/`)

A Go client library for interacting with Microsoft Graph API with automatic token expiration checking and renewal.

**Features:**
- Token-based authentication
- Automatic token expiration checking
- Automatic token refresh using refresh tokens
- User profile management
- Extensible architecture

**Quick Start:**
```bash
cd client
export MS_GRAPH_ACCESS_TOKEN=your_token
export MS_GRAPH_REFRESH_TOKEN=your_refresh_token  # Optional
go run cmd/main.go
```

See [client/README.md](client/README.md) for detailed documentation.

### 2. Chrome Extension (`extension/`)

A Chrome browser extension that automatically extracts access and refresh tokens from Microsoft Graph Explorer.

**Features:**
- Automatic token extraction from Graph Explorer
- Easy-to-use popup interface
- One-click copy as environment variables
- Secure local storage

**Quick Start:**
1. Open Chrome → `chrome://extensions/`
2. Enable "Developer mode"
3. Click "Load unpacked" → Select `extension/` directory
4. Open Graph Explorer and make a request
5. Click extension icon to view/copy tokens

See [extension/README.md](extension/README.md) for detailed documentation.

## Getting Started

### Complete Workflow

1. **Install the Chrome Extension**
   - Follow instructions in [extension/README.md](extension/README.md)
   - Extract tokens from Graph Explorer

2. **Set Up Environment Variables**
   ```bash
   export MS_GRAPH_ACCESS_TOKEN="your_access_token"
   export MS_GRAPH_REFRESH_TOKEN="your_refresh_token"
   export MS_GRAPH_TENANT_ID="your_tenant_id"  # Optional
   ```

3. **Use the Go Client**
   ```bash
   cd client
   go run cmd/main.go
   ```

## Development

### Go Client

```bash
cd client
go build ./...
go test ./...
```

### Chrome Extension

1. Make changes to files in `extension/`
2. Go to `chrome://extensions/`
3. Click refresh on the extension card
4. Test changes

## CI/CD

This repository includes GitHub Actions workflows for automated building and releasing:

- **Path-based triggers**: Workflows only run when files in the respective directory change
- **Multi-platform builds**: Go client builds for Linux, Windows, and macOS (amd64/arm64)
- **Automatic releases**: Creates GitHub releases on pushes to main/master branch
- **CI checks**: Runs tests and validations on pull requests

See [.github/workflows/README.md](.github/workflows/README.md) for detailed documentation.

### Workflows

- `client.yml` - Builds and releases Go client for all platforms
- `extension.yml` - Packages and releases Chrome extension
- `ci.yml` - Runs tests and validations

## License

This project is provided as-is for educational and development purposes.

