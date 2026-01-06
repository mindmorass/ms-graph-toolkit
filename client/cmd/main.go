package main

import (
	"fmt"
	"os"
	"time"
	"ms_graph/internal/graph"
	"ms_graph/internal/profile"
	"ms_graph/internal/token"
)

func main() {
	// Get access token from environment variable
	accessToken := os.Getenv("MS_GRAPH_ACCESS_TOKEN")
	if accessToken == "" {
		fmt.Fprintf(os.Stderr, "Error: MS_GRAPH_ACCESS_TOKEN environment variable is not set\n")
		fmt.Fprintf(os.Stderr, "Please set it with: export MS_GRAPH_ACCESS_TOKEN=your_token_here\n")
		os.Exit(1)
	}

	// Get refresh token and tenant ID (optional)
	refreshToken := os.Getenv("MS_GRAPH_REFRESH_TOKEN")
	tenantID := os.Getenv("MS_GRAPH_TENANT_ID")

	// Check token expiration and display info
	tokenInfo, err := token.ParseToken(accessToken)
	if err == nil {
		fmt.Println("=== Token Information ===")
		fmt.Printf("Expires At: %s\n", tokenInfo.ExpiresAt.Format(time.RFC3339))
		fmt.Printf("Time Until Expiration: %v\n", tokenInfo.TimeUntilExp.Round(time.Second))
		if tokenInfo.IsExpired {
			fmt.Println("⚠️  Token is EXPIRED")
		} else if tokenInfo.ExpiresSoon {
			fmt.Println("⚠️  Token is expiring soon (within 10 minutes)")
		} else {
			fmt.Println("✓ Token is valid")
		}
		fmt.Println()
	}

	// Create Graph API client with automatic refresh if refresh token is available
	var client *graph.Client
	if refreshToken != "" {
		fmt.Println("Using client with automatic token refresh...")
		client = graph.NewClientWithRefresh(accessToken, refreshToken, tenantID).Client
	} else {
		fmt.Println("Using basic client (no automatic refresh - token will be checked but not refreshed)")
		client = graph.NewClient(accessToken)
	}

	// Get current user profile
	fmt.Println("Fetching your profile...")
	user, err := profile.GetMyProfile(client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error retrieving profile: %v\n", err)
		os.Exit(1)
	}

	// Display user information
	fmt.Println("\n=== Your Profile ===")
	fmt.Printf("ID: %s\n", user.ID)
	fmt.Printf("Display Name: %s\n", user.DisplayName)
	fmt.Printf("Given Name: %s\n", user.GivenName)
	fmt.Printf("Surname: %s\n", user.Surname)
	fmt.Printf("Email: %s\n", user.Mail)
	fmt.Printf("User Principal Name: %s\n", user.UserPrincipalName)
	if user.JobTitle != "" {
		fmt.Printf("Job Title: %s\n", user.JobTitle)
	}
	if user.Department != "" {
		fmt.Printf("Department: %s\n", user.Department)
	}
	if user.OfficeLocation != "" {
		fmt.Printf("Office Location: %s\n", user.OfficeLocation)
	}
	if user.MobilePhone != "" {
		fmt.Printf("Mobile Phone: %s\n", user.MobilePhone)
	}
	if len(user.BusinessPhones) > 0 {
		fmt.Printf("Business Phones: %v\n", user.BusinessPhones)
	}
}

