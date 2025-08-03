package middleware

import (
	"contact-service/pkg/auth"
	"contact-service/pkg/logger"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.LogSecurityEvent("missing_auth_header", nil, c.ClientIP(), map[string]interface{}{
				"path":   c.Request.URL.Path,
				"method": c.Request.Method,
			})
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization header required",
				"error": map[string]string{
					"code":    "MISSING_AUTH_HEADER",
					"message": "Authorization header is required",
				},
			})
			c.Abort()
			return
		}

		// Extract token from header
		token, err := auth.ExtractTokenFromHeader(authHeader)
		if err != nil {
			logger.LogSecurityEvent("invalid_auth_header", nil, c.ClientIP(), map[string]interface{}{
				"error": err.Error(),
				"path":  c.Request.URL.Path,
			})
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid authorization header",
				"error": map[string]string{
					"code":    "INVALID_AUTH_HEADER",
					"message": err.Error(),
				},
			})
			c.Abort()
			return
		}

		// Validate token
		claims, err := auth.ValidateAccessToken(token)
		if err != nil {
			logger.LogSecurityEvent("invalid_token", nil, c.ClientIP(), map[string]interface{}{
				"error": err.Error(),
				"path":  c.Request.URL.Path,
			})
			
			// Determine specific error type
			errorCode := "INVALID_TOKEN"
			if strings.Contains(err.Error(), "expired") {
				errorCode = "TOKEN_EXPIRED"
			}
			
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid or expired token",
				"error": map[string]string{
					"code":    errorCode,
					"message": err.Error(),
				},
			})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("token_claims", claims)

		// Log successful authentication
		logger.Debug("User authenticated successfully", map[string]interface{}{
			"user_id": claims.UserID,
			"email":   claims.Email,
			"role":    claims.Role,
			"path":    c.Request.URL.Path,
		})

		c.Next()
	}
}

// OptionalAuthMiddleware validates JWT tokens if present but doesn't require them
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth header, continue without authentication
			c.Next()
			return
		}

		// Extract and validate token if present
		token, err := auth.ExtractTokenFromHeader(authHeader)
		if err != nil {
			// Invalid header format, continue without authentication
			c.Next()
			return
		}

		claims, err := auth.ValidateAccessToken(token)
		if err != nil {
			// Invalid token, continue without authentication
			c.Next()
			return
		}

		// Set user context if token is valid
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("token_claims", claims)

		c.Next()
	}
}

// AdminOnly middleware requires admin role
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Access denied",
				"error": map[string]string{
					"code":    "ACCESS_DENIED",
					"message": "User role not found",
				},
			})
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok || userRole != "admin" {
			userID, _ := c.Get("user_id")
			logger.LogSecurityEvent("unauthorized_admin_access", userID.(*uint), c.ClientIP(), map[string]interface{}{
				"role": userRole,
				"path": c.Request.URL.Path,
			})
			
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Admin access required",
				"error": map[string]string{
					"code":    "ADMIN_REQUIRED",
					"message": "This endpoint requires admin privileges",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission middleware checks if user has specific permission
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Access denied",
				"error": map[string]string{
					"code":    "ACCESS_DENIED",
					"message": "User role not found",
				},
			})
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Access denied",
				"error": map[string]string{
					"code":    "ACCESS_DENIED",
					"message": "Invalid user role",
				},
			})
			c.Abort()
			return
		}

		if !auth.HasPermission(userRole, permission) {
			userID, _ := c.Get("user_id")
			logger.LogSecurityEvent("insufficient_permissions", userID.(*uint), c.ClientIP(), map[string]interface{}{
				"role":            userRole,
				"required_permission": permission,
				"path":            c.Request.URL.Path,
			})
			
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Insufficient permissions",
				"error": map[string]string{
					"code":    "INSUFFICIENT_PERMISSIONS",
					"message": "You don't have permission to access this resource",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ManagerOrAbove middleware requires manager role or higher
func ManagerOrAbove() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Access denied",
				"error": map[string]string{
					"code":    "ACCESS_DENIED",
					"message": "User role not found",
				},
			})
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Access denied",
			})
			c.Abort()
			return
		}

		allowedRoles := []string{"admin", "manager"}
		isAllowed := false
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			userID, _ := c.Get("user_id")
			logger.LogSecurityEvent("insufficient_role", userID.(*uint), c.ClientIP(), map[string]interface{}{
				"role":          userRole,
				"required_roles": allowedRoles,
				"path":          c.Request.URL.Path,
			})
			
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Manager access required",
				"error": map[string]string{
					"code":    "MANAGER_REQUIRED",
					"message": "This endpoint requires manager or admin privileges",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetCurrentUser returns user information from context
func GetCurrentUser(c *gin.Context) (*UserInfo, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return nil, false
	}

	email, _ := c.Get("user_email")
	role, _ := c.Get("user_role")

	user := &UserInfo{
		ID:    userID.(uint),
		Email: email.(string),
		Role:  role.(string),
	}

	return user, true
}

// UserInfo represents current user information
type UserInfo struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// IsAuthenticated checks if request is authenticated
func IsAuthenticated(c *gin.Context) bool {
	_, exists := c.Get("user_id")
	return exists
}

// GetUserRole returns the current user's role
func GetUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get("user_role")
	if !exists {
		return "", false
	}
	
	userRole, ok := role.(string)
	return userRole, ok
}

// HasRole checks if current user has specific role
func HasRole(c *gin.Context, requiredRole string) bool {
	role, exists := GetUserRole(c)
	return exists && role == requiredRole
}

// CanAccessResource checks if user can access a specific resource
func CanAccessResource(c *gin.Context, resourceType string, action string) bool {
	role, exists := GetUserRole(c)
	if !exists {
		return false
	}
	
	permission := resourceType + ":" + action
	return auth.HasPermission(role, permission)
}