package models

import (
	"time"
)

// AdminUser represents an admin user in the system - matches existing database structure
type AdminUser struct {
	ID                   uint       `json:"id" gorm:"primaryKey"`
	Email                string     `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash         string     `json:"-" gorm:"not null"`
	Name                 string     `json:"name" gorm:"not null"`
	Role                 string     `json:"role" gorm:"type:enum('admin','editor','hr_manager','content_writer');default:'editor'"`
	AvatarURL            *string    `json:"avatar_url"`
	Phone                *string    `json:"phone"`
	JobTitle             *string    `json:"job_title"`
	Department           *string    `json:"department"`
	Location             *string    `json:"location"`
	Bio                  *string    `json:"bio" gorm:"type:text"`
	IsActive             bool       `json:"is_active" gorm:"default:true"`
	LoginAttempts        int        `json:"login_attempts" gorm:"default:0"`
	TwoFactorEnabled     bool       `json:"two_factor_enabled" gorm:"default:false"`
	LastLoginAt          *time.Time `json:"last_login_at"`
	LastActivityAt       *time.Time `json:"last_activity_at"`
	PasswordChangedAt    *time.Time `json:"password_changed_at"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	DeletedAt            *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName sets the table name
func (AdminUser) TableName() string {
	return "admin_users"
}

// GetFullName returns the user's full name
func (u *AdminUser) GetFullName() string {
	return u.Name
}

// GetDisplayName returns a display name for the user
func (u *AdminUser) GetDisplayName() string {
	if u.Name != "" {
		return u.Name
	}
	return u.Email
}

// IsAdmin checks if the user has admin role
func (u *AdminUser) IsAdmin() bool {
	return u.Role == "admin"
}

// IsManager checks if the user has manager role or higher
func (u *AdminUser) IsManager() bool {
	return u.Role == "admin" || u.Role == "hr_manager"
}

// CanAccessResource checks if the user can access a specific resource
func (u *AdminUser) CanAccessResource(resource, action string) bool {
	// This would integrate with the auth package's HasPermission function
	// For now, simplified logic based on existing roles
	switch u.Role {
	case "admin":
		return true
	case "hr_manager":
		return resource != "admin" // HR managers can access everything except admin functions
	case "editor":
		return resource == "contacts" || resource == "activities" || resource == "search" || resource == "blogs"
	case "content_writer":
		return resource == "blogs" || (resource == "contacts" && action == "read")
	default:
		return false
	}
}

// AdminUserRequest represents a request to create or update an admin user
type AdminUserRequest struct {
	Email            string  `json:"email" binding:"required,email"`
	Password         *string `json:"password,omitempty" binding:"omitempty,min=8"`
	Name             string  `json:"name" binding:"required,max=100"`
	Role             string  `json:"role" binding:"required,oneof=admin editor hr_manager content_writer"`
	AvatarURL        *string `json:"avatar_url,omitempty"`
	Phone            *string `json:"phone,omitempty"`
	JobTitle         *string `json:"job_title,omitempty"`
	Department       *string `json:"department,omitempty"`
	Location         *string `json:"location,omitempty"`
	Bio              *string `json:"bio,omitempty"`
	IsActive         *bool   `json:"is_active,omitempty"`
	TwoFactorEnabled *bool   `json:"two_factor_enabled,omitempty"`
}

// AdminUserResponse represents the API response format for admin users
type AdminUserResponse struct {
	ID               uint       `json:"id"`
	Email            string     `json:"email"`
	Name             string     `json:"name"`
	FullName         string     `json:"full_name"`
	DisplayName      string     `json:"display_name"`
	Role             string     `json:"role"`
	AvatarURL        *string    `json:"avatar_url,omitempty"`
	Phone            *string    `json:"phone,omitempty"`
	JobTitle         *string    `json:"job_title,omitempty"`
	Department       *string    `json:"department,omitempty"`
	Location         *string    `json:"location,omitempty"`
	Bio              *string    `json:"bio,omitempty"`
	IsActive         bool       `json:"is_active"`
	LoginAttempts    int        `json:"login_attempts"`
	TwoFactorEnabled bool       `json:"two_factor_enabled"`
	LastLoginAt      *time.Time `json:"last_login_at,omitempty"`
	LastActivityAt   *time.Time `json:"last_activity_at,omitempty"`
	PasswordChangedAt *time.Time `json:"password_changed_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	IsAdmin          bool       `json:"is_admin"`
	IsManager        bool       `json:"is_manager"`
}

// ToResponse converts AdminUser to AdminUserResponse
func (u *AdminUser) ToResponse() *AdminUserResponse {
	return &AdminUserResponse{
		ID:                u.ID,
		Email:             u.Email,
		Name:              u.Name,
		FullName:          u.GetFullName(),
		DisplayName:       u.GetDisplayName(),
		Role:              u.Role,
		AvatarURL:         u.AvatarURL,
		Phone:             u.Phone,
		JobTitle:          u.JobTitle,
		Department:        u.Department,
		Location:          u.Location,
		Bio:               u.Bio,
		IsActive:          u.IsActive,
		LoginAttempts:     u.LoginAttempts,
		TwoFactorEnabled:  u.TwoFactorEnabled,
		LastLoginAt:       u.LastLoginAt,
		LastActivityAt:    u.LastActivityAt,
		PasswordChangedAt: u.PasswordChangedAt,
		CreatedAt:         u.CreatedAt,
		UpdatedAt:         u.UpdatedAt,
		IsAdmin:           u.IsAdmin(),
		IsManager:         u.IsManager(),
	}
}