package handlers

import (
	"contact-service/internal/models"
	"contact-service/pkg/auth"
	"contact-service/pkg/database"
	"contact-service/pkg/logger"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	db *gorm.DB
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		db: database.DB,
	}
}

// Login godoc
// @Summary User login
// @Description Authenticate user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "User credentials"
// @Success 200 {object} APIResponse{data=LoginResponse}
// @Failure 400 {object} APIResponse
// @Failure 401 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.LogSecurityEvent("invalid_login_request", nil, c.ClientIP(), map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request data", err.Error()))
		return
	}

	// Find user by email
	var user models.AdminUser
	if err := h.db.Where("email = ? AND is_active = ?", strings.ToLower(req.Email), true).
		First(&user).Error; err != nil {
		logger.LogSecurityEvent("login_failed", nil, c.ClientIP(), map[string]interface{}{
			"email":  req.Email,
			"reason": "user_not_found",
		})
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Invalid credentials", ""))
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		logger.LogSecurityEvent("login_failed", &user.ID, c.ClientIP(), map[string]interface{}{
			"email":  req.Email,
			"reason": "invalid_password",
		})
		
		// Update failed login attempts
		h.updateFailedLoginAttempts(&user, c.ClientIP())
		
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Invalid credentials", ""))
		return
	}

	// Check if account has too many failed attempts (simple check)
	if user.LoginAttempts >= 5 {
		logger.LogSecurityEvent("login_blocked", &user.ID, c.ClientIP(), map[string]interface{}{
			"email":          req.Email,
			"reason":         "too_many_attempts",
			"login_attempts": user.LoginAttempts,
		})
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Too many failed attempts", "Please try again later"))
		return
	}

	// Reset failed attempts on successful login
	if user.LoginAttempts > 0 {
		h.db.Model(&user).Updates(map[string]interface{}{
			"login_attempts": 0,
		})
	}

	// Generate token pair
	tokenPair, err := auth.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		logger.Error("Failed to generate tokens", err, map[string]interface{}{
			"user_id": user.ID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Login failed", ""))
		return
	}

	// Update user login info
	now := time.Now()
	h.db.Model(&user).Updates(map[string]interface{}{
		"last_login_at":    now,
		"last_activity_at": now,
	})

	// Log successful login
	logger.LogSecurityEvent("login_success", &user.ID, c.ClientIP(), map[string]interface{}{
		"email": user.Email,
		"role":  user.Role,
	})

	response := &LoginResponse{
		User: UserResponse{
			ID        : user.ID,
			Email     : user.Email,
			Name      : user.Name,
			Role      : user.Role,
			IsActive  : user.IsActive,
			CreatedAt : user.CreatedAt,
			UpdatedAt : user.UpdatedAt,
		},
		Tokens: *tokenPair,
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Login successful", response))
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Refresh an expired access token using a valid refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} APIResponse{data=auth.TokenPair}
// @Failure 400 {object} APIResponse
// @Failure 401 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request data", err.Error()))
		return
	}

	// Refresh the token
	tokenPair, err := auth.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		logger.LogSecurityEvent("token_refresh_failed", nil, c.ClientIP(), map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Invalid refresh token", ""))
		return
	}

	// Get user info from token for logging
	if claims, err := auth.ValidateRefreshToken(req.RefreshToken); err == nil {
		logger.LogSecurityEvent("token_refreshed", &claims.UserID, c.ClientIP(), map[string]interface{}{
			"user_id": claims.UserID,
			"email":   claims.Email,
		})
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Token refreshed successfully", tokenPair))
}

// Logout godoc
// @Summary User logout
// @Description Logout user and invalidate token
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse
// @Security BearerAuth
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userID := getUserIDFromContext(c)
	
	if userID != nil {
		logger.LogSecurityEvent("logout", userID, c.ClientIP(), map[string]interface{}{
			"user_id": *userID,
		})
	}

	// In a production system, you might want to:
	// 1. Add the token to a blacklist
	// 2. Store token revocation in Redis/database
	// For now, we'll just return success since JWT tokens are stateless
	
	c.JSON(http.StatusOK, NewSuccessResponse("Logout successful", nil))
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get current user's profile information
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=UserResponse}
// @Failure 401 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	var user models.AdminUser
	if err := h.db.First(&user, *userID).Error; err != nil {
		logger.Error("Failed to get user profile", err, map[string]interface{}{
			"user_id": *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get profile", ""))
		return
	}

	response := UserResponse{
		ID        : user.ID,
		Email     : user.Email,
		Name      : user.Name,
		Role      : user.Role,
		IsActive  : user.IsActive,
		CreatedAt : user.CreatedAt,
		UpdatedAt : user.UpdatedAt,
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Profile retrieved successfully", response))
}

// ChangePassword godoc
// @Summary Change user password
// @Description Change current user's password
// @Tags auth
// @Accept json
// @Produce json
// @Param password body ChangePasswordRequest true "Password change data"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 401 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request data", err.Error()))
		return
	}

	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	var user models.AdminUser
	if err := h.db.First(&user, *userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse("User not found", ""))
		return
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		logger.LogSecurityEvent("password_change_failed", userID, c.ClientIP(), map[string]interface{}{
			"reason": "invalid_current_password",
		})
		c.JSON(http.StatusBadRequest, NewErrorResponse("Current password is incorrect", ""))
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Failed to hash password", err, map[string]interface{}{
			"user_id": *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to change password", ""))
		return
	}

	// Update password
	if err := h.db.Model(&user).Update("password_hash", string(hashedPassword)).Error; err != nil {
		logger.Error("Failed to update password", err, map[string]interface{}{
			"user_id": *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to change password", ""))
		return
	}

	logger.LogSecurityEvent("password_changed", userID, c.ClientIP(), map[string]interface{}{
		"user_id": *userID,
	})

	c.JSON(http.StatusOK, NewSuccessResponse("Password changed successfully", nil))
}

// ValidateToken godoc
// @Summary Validate token
// @Description Validate if the current token is valid
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=TokenValidationResponse}
// @Failure 401 {object} APIResponse
// @Security BearerAuth
// @Router /auth/validate [get]
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	userID := getUserIDFromContext(c)
	userEmail, _ := c.Get("user_email")
	userRole, _ := c.Get("user_role")

	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	response := &TokenValidationResponse{
		Valid:  true,
		UserID: *userID,
		Email:  userEmail.(string),
		Role:   userRole.(string),
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Token is valid", response))
}

// Helper methods

func (h *AuthHandler) updateFailedLoginAttempts(user *models.AdminUser, clientIP string) {
	attempts := user.LoginAttempts + 1
	
	updates := map[string]interface{}{
		"login_attempts": attempts,
	}

	// Log warning after 3 failed attempts
	if attempts >= 3 {
		logger.LogSecurityEvent("multiple_failed_attempts", &user.ID, clientIP, map[string]interface{}{
			"failed_attempts": attempts,
		})
	}

	h.db.Model(user).Updates(updates)
}

// Request/Response types

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginResponse struct {
	User   UserResponse      `json:"user"`
	Tokens auth.TokenPair    `json:"tokens"`
}

type UserResponse struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,min=6"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

type TokenValidationResponse struct {
	Valid  bool   `json:"valid"`
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}