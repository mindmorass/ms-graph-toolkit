package graph

// User represents a Microsoft Graph user object
type User struct {
	ID                string   `json:"id"`
	DisplayName       string   `json:"displayName"`
	GivenName         string   `json:"givenName"`
	Surname           string   `json:"surname"`
	Mail              string   `json:"mail"`
	UserPrincipalName string   `json:"userPrincipalName"`
	JobTitle          string   `json:"jobTitle"`
	Department        string   `json:"department"`
	OfficeLocation    string   `json:"officeLocation"`
	MobilePhone       string   `json:"mobilePhone"`
	BusinessPhones    []string `json:"businessPhones"`
}

// ErrorResponse represents an error response from Microsoft Graph API
type ErrorResponse struct {
	Error Error `json:"error"`
}

// Error represents an error object within an ErrorResponse
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// TokenResponse represents a response from the OAuth2 token endpoint
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope"`
}

