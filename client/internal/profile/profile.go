package profile

import (
	"fmt"
	"ms_graph/internal/graph"
)

// GetMyProfile retrieves the current user's profile from Microsoft Graph API
func GetMyProfile(client *graph.Client) (*graph.User, error) {
	var user graph.User
	if err := client.Get("/me", &user); err != nil {
		return nil, fmt.Errorf("failed to get my profile: %w", err)
	}
	return &user, nil
}

// GetUserProfile retrieves a user's profile by ID from Microsoft Graph API
func GetUserProfile(client *graph.Client, userID string) (*graph.User, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}

	var user graph.User
	endpoint := fmt.Sprintf("/users/%s", userID)
	if err := client.Get(endpoint, &user); err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	return &user, nil
}

