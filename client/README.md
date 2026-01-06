# Microsoft Graph API Go Client

A Go client library for interacting with Microsoft Graph API to manage Office 365 resources. This project provides a simple interface for working with Microsoft Graph API using an existing access token, with automatic token expiration checking and renewal capabilities.

## Features

- Token-based authentication (no OAuth flow required)
- Automatic token expiration checking
- Automatic token refresh using refresh tokens
- User profile management
- Extensible client architecture for additional Graph API operations

## Prerequisites

- Go 1.19 or later
- A valid Microsoft Graph API access token

## Installation

1. Clone or navigate to this repository:
   ```bash
   cd ms_graph
   ```

2. The project uses Go modules, so dependencies will be automatically managed.

## Configuration

### Basic Configuration (Access Token Only)

Set your Microsoft Graph API access token as an environment variable:

```bash
export MS_GRAPH_ACCESS_TOKEN=your_access_token_here
```

### Advanced Configuration (With Automatic Token Refresh)

For automatic token renewal, you'll need both an access token and a refresh token:

```bash
export MS_GRAPH_ACCESS_TOKEN=your_access_token_here
export MS_GRAPH_REFRESH_TOKEN=your_refresh_token_here
export MS_GRAPH_TENANT_ID=your_tenant_id  # Optional, defaults to "common"
```

#### Getting a Refresh Token from Graph Explorer

1. Open [Microsoft Graph Explorer](https://developer.microsoft.com/graph/graph-explorer)
2. Sign in and authenticate
3. Open your browser's Developer Tools (F12 or right-click → Inspect)
4. Go to the **Network** tab
5. Make a request in Graph Explorer (e.g., click "Run query")
6. Look for a request to `login.microsoftonline.com` or find the token request
7. In the request/response, find the `refresh_token` field
8. Copy the refresh token value

**Note:** The refresh token is obtained once and can be reused for all future token renewals. All renewals happen automatically via API calls - no browser interaction needed after the initial setup.

## Usage

### Running the Example Application

The example application retrieves and displays your profile information:

```bash
go run cmd/main.go
```

### Using the Library

#### Basic Client (No Automatic Refresh)

```go
package main

import (
    "fmt"
    "os"
    "ms_graph/internal/graph"
    "ms_graph/internal/profile"
)

func main() {
    // Create client with access token
    token := os.Getenv("MS_GRAPH_ACCESS_TOKEN")
    client := graph.NewClient(token)
    
    // Get your profile
    user, err := profile.GetMyProfile(client)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("User: %s (%s)\n", user.DisplayName, user.Mail)
}
```

#### Client with Automatic Token Refresh

```go
package main

import (
    "fmt"
    "os"
    "ms_graph/internal/graph"
    "ms_graph/internal/profile"
)

func main() {
    // Create client with automatic token refresh
    client := graph.NewClientWithRefresh(
        os.Getenv("MS_GRAPH_ACCESS_TOKEN"),
        os.Getenv("MS_GRAPH_REFRESH_TOKEN"),
        os.Getenv("MS_GRAPH_TENANT_ID"), // Optional
    )
    
    // Client automatically checks expiration and refreshes if needed
    // All renewals happen via API calls - no browser interaction needed
    user, err := profile.GetMyProfile(client.Client)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("User: %s (%s)\n", user.DisplayName, user.Mail)
}
```

## Project Structure

```
ms_graph/
├── cmd/
│   └── main.go                 # Example application
├── internal/
│   ├── graph/
│   │   ├── client.go           # Core Graph API client with refresh support
│   │   └── types.go            # Type definitions
│   ├── token/
│   │   └── token.go            # JWT parsing and validation
│   ├── auth/
│   │   └── refresh.go          # Token refresh using OAuth2 endpoint
│   └── profile/
│       └── profile.go           # Profile operations
├── go.mod                      # Go module definition
└── README.md                   # This file
```

## API Reference

### Graph Client

The `graph.Client` provides methods for making HTTP requests to Microsoft Graph API:

- `Get(endpoint string, result interface{}) error` - GET request
- `Post(endpoint string, payload interface{}, result interface{}) error` - POST request
- `Patch(endpoint string, payload interface{}, result interface{}) error` - PATCH request
- `Delete(endpoint string) error` - DELETE request

### Client with Automatic Refresh

The `graph.ClientWithRefresh` extends the basic client with automatic token refresh:

- Automatically checks token expiration before each request
- Refreshes token if expired or expiring soon (within 10 minutes)
- Handles 401 errors by refreshing and retrying the request
- Updates tokens seamlessly in the background

**Constructor:**
- `NewClientWithRefresh(accessToken, refreshToken, tenantID string) *ClientWithRefresh`

All HTTP methods (Get, Post, Patch, Delete) are automatically enhanced with refresh capabilities.

### Profile Operations

- `GetMyProfile(client *graph.Client) (*graph.User, error)` - Get current user's profile
- `GetUserProfile(client *graph.Client, userID string) (*graph.User, error)` - Get user profile by ID

## Getting Tokens

### Access Token

This project assumes you already have a valid access token. To obtain one, you can:

1. Use Azure CLI: `az account get-access-token --resource https://graph.microsoft.com`
2. Use Microsoft Graph Explorer: https://developer.microsoft.com/graph/graph-explorer
3. Use your organization's authentication system

### Refresh Token (For Automatic Renewal)

To enable automatic token renewal, you need a refresh token. The easiest way is to extract it from Graph Explorer:

1. Open [Graph Explorer](https://developer.microsoft.com/graph/graph-explorer) and sign in
2. Open browser DevTools (F12) → Network tab
3. Make a request in Graph Explorer
4. Find the token request in the Network tab
5. Look for `refresh_token` in the response
6. Copy the refresh token value

**Important:** The refresh token is obtained **once** and reused for all future renewals. All token renewals happen automatically via API calls to Microsoft's token endpoint - no browser interaction needed after initial setup.

## Token Management

### Token Expiration Checking

The client automatically checks token expiration before making API requests:
- Parses JWT tokens to extract expiration time
- Warns if token is expiring soon (within 10 minutes)
- Provides clear error messages if token is expired

### Automatic Token Refresh

When using `ClientWithRefresh`:
- Token is automatically refreshed if expired or expiring soon
- 401 errors trigger automatic refresh and retry
- New access tokens are seamlessly updated
- Refresh token rotation is handled automatically (if new refresh token provided)

### Manual Token Validation

You can also manually check token expiration:

```go
import "ms_graph/internal/token"

tokenInfo, err := token.ParseToken(accessToken)
if err == nil {
    fmt.Printf("Expires at: %s\n", tokenInfo.ExpiresAt)
    fmt.Printf("Time until expiration: %v\n", tokenInfo.TimeUntilExp)
    if tokenInfo.IsExpired {
        fmt.Println("Token is expired")
    }
}
```

## Error Handling

The client handles API errors and returns descriptive error messages. Errors from the Microsoft Graph API are parsed and returned with their error codes and messages. When using automatic refresh, 401 errors are automatically handled by refreshing the token and retrying the request.

## License

This project is provided as-is for educational and development purposes.

