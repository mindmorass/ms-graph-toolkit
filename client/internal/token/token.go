package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenInfo contains information about a JWT token
type TokenInfo struct {
	ExpiresAt    time.Time
	IsExpired    bool
	TimeUntilExp time.Duration
	ExpiresSoon  bool
}

// ParseToken extracts expiration information from a JWT token
func ParseToken(tokenString string) (*TokenInfo, error) {
	// Parse without verification since we only need to read claims
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Extract expiration time
	exp, ok := claims["exp"]
	if !ok {
		return nil, fmt.Errorf("token does not contain expiration claim")
	}

	var expTime time.Time
	switch v := exp.(type) {
	case float64:
		expTime = time.Unix(int64(v), 0)
	case int64:
		expTime = time.Unix(v, 0)
	default:
		return nil, fmt.Errorf("invalid expiration claim type")
	}

	now := time.Now()
	timeUntilExp := expTime.Sub(now)
	isExpired := now.After(expTime)
	expiresSoon := !isExpired && timeUntilExp < 10*time.Minute

	return &TokenInfo{
		ExpiresAt:     expTime,
		IsExpired:     isExpired,
		TimeUntilExp:  timeUntilExp,
		ExpiresSoon:   expiresSoon,
	}, nil
}

// IsExpired checks if a token is expired
func IsExpired(tokenString string) (bool, error) {
	info, err := ParseToken(tokenString)
	if err != nil {
		return false, err
	}
	return info.IsExpired, nil
}

// IsExpiringSoon checks if a token is expiring within 10 minutes
func IsExpiringSoon(tokenString string) (bool, error) {
	info, err := ParseToken(tokenString)
	if err != nil {
		return false, err
	}
	return info.ExpiresSoon, nil
}

// GetExpirationTime returns the expiration time of a token
func GetExpirationTime(tokenString string) (time.Time, error) {
	info, err := ParseToken(tokenString)
	if err != nil {
		return time.Time{}, err
	}
	return info.ExpiresAt, nil
}

// GetTimeUntilExpiration returns the duration until token expiration
func GetTimeUntilExpiration(tokenString string) (time.Duration, error) {
	info, err := ParseToken(tokenString)
	if err != nil {
		return 0, err
	}
	return info.TimeUntilExp, nil
}

