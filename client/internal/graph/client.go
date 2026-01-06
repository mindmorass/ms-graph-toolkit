package graph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"ms_graph/internal/token"
)

const (
	// BaseURL is the base URL for Microsoft Graph API
	BaseURL = "https://graph.microsoft.com/v1.0"
)

// Client represents a Microsoft Graph API client
type Client struct {
	accessToken string
	httpClient  *http.Client
	baseURL     string
	mu          sync.RWMutex // Protects accessToken updates
}

// ClientWithRefresh represents a Microsoft Graph API client with automatic token refresh
type ClientWithRefresh struct {
	*Client
	refreshToken string
	tenantID     string
	mu           sync.Mutex // Protects refresh operations
}

// NewClient creates a new Graph API client with the provided access token
func NewClient(accessToken string) *Client {
	return &Client{
		accessToken: accessToken,
		httpClient:  &http.Client{},
		baseURL:     BaseURL,
	}
}

// NewClientWithRefresh creates a new Graph API client with automatic token refresh capability
func NewClientWithRefresh(accessToken, refreshToken, tenantID string) *ClientWithRefresh {
	return &ClientWithRefresh{
		Client: &Client{
			accessToken: accessToken,
			httpClient:  &http.Client{},
			baseURL:     BaseURL,
		},
		refreshToken: refreshToken,
		tenantID:     tenantID,
	}
}

// checkAndRefreshToken checks if token is expired or expiring soon and refreshes if needed
func (c *ClientWithRefresh) checkAndRefreshToken() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check token expiration
	tokenInfo, err := token.ParseToken(c.accessToken)
	if err != nil {
		// If we can't parse the token, try to refresh anyway if we have a refresh token
		if c.refreshToken == "" {
			return fmt.Errorf("failed to parse token and no refresh token available: %w", err)
		}
	} else {
		// If token is not expired and not expiring soon, no need to refresh
		if !tokenInfo.IsExpired && !tokenInfo.ExpiresSoon {
			return nil
		}
	}

	// Refresh token if we have one
	if c.refreshToken == "" {
		if tokenInfo != nil && tokenInfo.IsExpired {
			return fmt.Errorf("token is expired and no refresh token available. Please get a new token from Graph Explorer")
		}
		return nil
	}

	// Attempt to refresh
	tokenResp, err := refreshToken(c.refreshToken, c.tenantID)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update access token
	c.Client.mu.Lock()
	c.Client.accessToken = tokenResp.AccessToken
	c.Client.mu.Unlock()

	// Update refresh token if a new one is provided (token rotation)
	if tokenResp.RefreshToken != "" {
		c.refreshToken = tokenResp.RefreshToken
	}

	return nil
}

// refreshTokenOn401 attempts to refresh token and retry the request on 401 errors
func (c *ClientWithRefresh) refreshTokenOn401(endpoint string, method string, body io.Reader, result interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.refreshToken == "" {
		return fmt.Errorf("received 401 error and no refresh token available for automatic refresh")
	}

	// Attempt to refresh
	tokenResp, err := refreshToken(c.refreshToken, c.tenantID)
	if err != nil {
		return fmt.Errorf("received 401 error and failed to refresh token: %w", err)
	}

	// Update access token
	c.Client.mu.Lock()
	c.Client.accessToken = tokenResp.AccessToken
	c.Client.mu.Unlock()

	// Update refresh token if a new one is provided
	if tokenResp.RefreshToken != "" {
		c.refreshToken = tokenResp.RefreshToken
	}

	// Retry the original request
	url := c.baseURL + endpoint
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Errorf("failed to create retry request: %w", err)
	}

	c.Client.mu.RLock()
	req.Header.Set("Authorization", "Bearer "+c.Client.accessToken)
	c.Client.mu.RUnlock()
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute retry request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read retry response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(respBody, &errorResp); err != nil {
			return fmt.Errorf("API error after refresh (status %d): %s", resp.StatusCode, string(respBody))
		}
		return fmt.Errorf("API error after refresh: %s - %s", errorResp.Error.Code, errorResp.Error.Message)
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal retry response: %w", err)
		}
	}

	return nil
}

// Get performs a GET request to the specified endpoint
func (c *Client) Get(endpoint string, result interface{}) error {
	c.mu.RLock()
	accessToken := c.accessToken
	c.mu.RUnlock()

	url := c.baseURL + endpoint
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
		return fmt.Errorf("API error: %s - %s", errorResp.Error.Code, errorResp.Error.Message)
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// Get performs a GET request with automatic token refresh
func (c *ClientWithRefresh) Get(endpoint string, result interface{}) error {
	// Check and refresh token before request
	if err := c.checkAndRefreshToken(); err != nil {
		return fmt.Errorf("token check failed: %w", err)
	}

	// Perform the request
	c.Client.mu.RLock()
	accessToken := c.Client.accessToken
	c.Client.mu.RUnlock()

	url := c.baseURL + endpoint
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle 401 errors by refreshing and retrying
	if resp.StatusCode == 401 {
		return c.refreshTokenOn401(endpoint, "GET", nil, result)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
		return fmt.Errorf("API error: %s - %s", errorResp.Error.Code, errorResp.Error.Message)
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// Post performs a POST request to the specified endpoint
func (c *Client) Post(endpoint string, payload interface{}, result interface{}) error {
	c.mu.RLock()
	accessToken := c.accessToken
	c.mu.RUnlock()

	url := c.baseURL + endpoint

	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			return fmt.Errorf("failed to encode payload: %w", err)
		}
	}

	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(respBody, &errorResp); err != nil {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
		}
		return fmt.Errorf("API error: %s - %s", errorResp.Error.Code, errorResp.Error.Message)
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// Post performs a POST request with automatic token refresh
func (c *ClientWithRefresh) Post(endpoint string, payload interface{}, result interface{}) error {
	// Check and refresh token before request
	if err := c.checkAndRefreshToken(); err != nil {
		return fmt.Errorf("token check failed: %w", err)
	}

	// Prepare body
	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			return fmt.Errorf("failed to encode payload: %w", err)
		}
	}

	// Perform the request
	c.Client.mu.RLock()
	accessToken := c.Client.accessToken
	c.Client.mu.RUnlock()

	url := c.baseURL + endpoint
	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle 401 errors by refreshing and retrying
	if resp.StatusCode == 401 {
		var retryBody bytes.Buffer
		if payload != nil {
			json.NewEncoder(&retryBody).Encode(payload)
		}
		return c.refreshTokenOn401(endpoint, "POST", &retryBody, result)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(respBody, &errorResp); err != nil {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
		}
		return fmt.Errorf("API error: %s - %s", errorResp.Error.Code, errorResp.Error.Message)
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// Patch performs a PATCH request to the specified endpoint
func (c *Client) Patch(endpoint string, payload interface{}, result interface{}) error {
	c.mu.RLock()
	accessToken := c.accessToken
	c.mu.RUnlock()

	url := c.baseURL + endpoint

	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			return fmt.Errorf("failed to encode payload: %w", err)
		}
	}

	req, err := http.NewRequest("PATCH", url, &body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(respBody, &errorResp); err != nil {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
		}
		return fmt.Errorf("API error: %s - %s", errorResp.Error.Code, errorResp.Error.Message)
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// Patch performs a PATCH request with automatic token refresh
func (c *ClientWithRefresh) Patch(endpoint string, payload interface{}, result interface{}) error {
	// Check and refresh token before request
	if err := c.checkAndRefreshToken(); err != nil {
		return fmt.Errorf("token check failed: %w", err)
	}

	// Prepare body
	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			return fmt.Errorf("failed to encode payload: %w", err)
		}
	}

	// Perform the request
	c.Client.mu.RLock()
	accessToken := c.Client.accessToken
	c.Client.mu.RUnlock()

	url := c.baseURL + endpoint
	req, err := http.NewRequest("PATCH", url, &body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle 401 errors by refreshing and retrying
	if resp.StatusCode == 401 {
		var retryBody bytes.Buffer
		if payload != nil {
			json.NewEncoder(&retryBody).Encode(payload)
		}
		return c.refreshTokenOn401(endpoint, "PATCH", &retryBody, result)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(respBody, &errorResp); err != nil {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
		}
		return fmt.Errorf("API error: %s - %s", errorResp.Error.Code, errorResp.Error.Message)
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// Delete performs a DELETE request to the specified endpoint
func (c *Client) Delete(endpoint string) error {
	c.mu.RLock()
	accessToken := c.accessToken
	c.mu.RUnlock()

	url := c.baseURL + endpoint
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
		return fmt.Errorf("API error: %s - %s", errorResp.Error.Code, errorResp.Error.Message)
	}

	return nil
}

// Delete performs a DELETE request with automatic token refresh
func (c *ClientWithRefresh) Delete(endpoint string) error {
	// Check and refresh token before request
	if err := c.checkAndRefreshToken(); err != nil {
		return fmt.Errorf("token check failed: %w", err)
	}

	// Perform the request
	c.Client.mu.RLock()
	accessToken := c.Client.accessToken
	c.Client.mu.RUnlock()

	url := c.baseURL + endpoint
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Handle 401 errors by refreshing and retrying
	if resp.StatusCode == 401 {
		return c.refreshTokenOn401(endpoint, "DELETE", nil, nil)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
		return fmt.Errorf("API error: %s - %s", errorResp.Error.Code, errorResp.Error.Message)
	}

	return nil
}

// refreshToken refreshes an access token using a refresh token
func refreshToken(refreshToken, tenantID string) (*TokenResponse, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("refresh token is required")
	}

	// Default to "common" if tenant ID is not provided
	if tenantID == "" {
		tenantID = "common"
	}

	endpoint := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantID)

	// Prepare form data
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("scope", "https://graph.microsoft.com/.default")

	// Create request
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("token refresh failed (status %d): %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("token refresh failed: %s - %s", errorResp.Error.Code, errorResp.Error.Message)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("token response does not contain access_token")
	}

	return &tokenResp, nil
}

