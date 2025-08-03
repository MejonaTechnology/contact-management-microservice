package auth

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the claims in our JWT token
type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

var (
	accessTokenSecret  = []byte(getEnv("JWT_ACCESS_SECRET", "mejona-contact-service-access-secret-2024"))
	refreshTokenSecret = []byte(getEnv("JWT_REFRESH_SECRET", "mejona-contact-service-refresh-secret-2024"))
	accessTokenTTL     = time.Duration(getEnvInt("JWT_ACCESS_TTL_MINUTES", 60)) * time.Minute
	refreshTokenTTL    = time.Duration(getEnvInt("JWT_REFRESH_TTL_HOURS", 168)) * time.Hour // 7 days
)

// GenerateTokenPair generates both access and refresh tokens
func GenerateTokenPair(userID uint, email, role string) (*TokenPair, error) {
	// Generate access token
	accessToken, err := GenerateAccessToken(userID, email, role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %v", err)
	}

	// Generate refresh token
	refreshToken, err := GenerateRefreshToken(userID, email, role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %v", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessTokenTTL.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// GenerateAccessToken generates a JWT access token
func GenerateAccessToken(userID uint, email, role string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "mejona-contact-service",
			Subject:   fmt.Sprintf("user:%d", userID),
			ID:        fmt.Sprintf("%d-%d", userID, time.Now().Unix()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(accessTokenSecret)
}

// GenerateRefreshToken generates a JWT refresh token
func GenerateRefreshToken(userID uint, email, role string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "mejona-contact-service",
			Subject:   fmt.Sprintf("user:%d", userID),
			ID:        fmt.Sprintf("refresh-%d-%d", userID, time.Now().Unix()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(refreshTokenSecret)
}

// ValidateAccessToken validates and parses an access token
func ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	return validateToken(tokenString, accessTokenSecret)
}

// ValidateRefreshToken validates and parses a refresh token
func ValidateRefreshToken(tokenString string) (*JWTClaims, error) {
	return validateToken(tokenString, refreshTokenSecret)
}

// validateToken validates a JWT token with the given secret
func validateToken(tokenString string, secret []byte) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token expired")
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, errors.New("token not valid yet")
		}
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("authorization header must be in format 'Bearer <token>'")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", errors.New("token cannot be empty")
	}

	return token, nil
}

// RefreshAccessToken generates a new access token using a valid refresh token
func RefreshAccessToken(refreshToken string) (*TokenPair, error) {
	claims, err := ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %v", err)
	}

	// Generate new token pair
	return GenerateTokenPair(claims.UserID, claims.Email, claims.Role)
}

// HasPermission checks if a role has a specific permission
func HasPermission(role, permission string) bool {
	permissions := getPermissionsForRole(role)
	for _, p := range permissions {
		if p == permission || p == "*" {
			return true
		}
	}
	return false
}

// getPermissionsForRole returns the permissions for a given role
func getPermissionsForRole(role string) []string {
	rolePermissions := map[string][]string{
		"admin": {
			"*", // Admin has all permissions
		},
		"hr_manager": {
			"contacts:read",
			"contacts:write",
			"contacts:update",
			"contacts:assign",
			"activities:read",
			"activities:write",
			"analytics:read",
			"search:read",
			"search:write",
			"bulk:read",
			"bulk:write",
		},
		"editor": {
			"contacts:read",
			"contacts:write",
			"contacts:update",
			"activities:read",
			"activities:write",
			"search:read",
			"search:write",
			"blogs:read",
			"blogs:write",
		},
		"content_writer": {
			"contacts:read",
			"activities:read",
			"search:read",
			"blogs:read",
			"blogs:write",
		},
	}

	if permissions, exists := rolePermissions[role]; exists {
		return permissions
	}

	return []string{} // No permissions for unknown roles
}

// GetUserRoles returns all available user roles
func GetUserRoles() []string {
	return []string{"admin", "hr_manager", "editor", "content_writer"}
}

// IsValidRole checks if a role is valid
func IsValidRole(role string) bool {
	validRoles := GetUserRoles()
	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	return false
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue := parseInt(value); intValue > 0 {
			return intValue
		}
	}
	return defaultValue
}

func parseInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}

// TokenInfo represents information about a token
type TokenInfo struct {
	UserID    uint      `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	ExpiresAt time.Time `json:"expires_at"`
	IssuedAt  time.Time `json:"issued_at"`
}

// GetTokenInfo extracts information from a token without validating it
func GetTokenInfo(tokenString string) (*TokenInfo, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &JWTClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok {
		return &TokenInfo{
			UserID:    claims.UserID,
			Email:     claims.Email,
			Role:      claims.Role,
			ExpiresAt: claims.ExpiresAt.Time,
			IssuedAt:  claims.IssuedAt.Time,
		}, nil
	}

	return nil, errors.New("invalid token claims")
}