package models

import (
	"time"
)

// SavedSearch represents a saved search query
type SavedSearch struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"not null;index"`
	Name        string    `json:"name" gorm:"not null;size:100"`
	Description *string   `json:"description,omitempty" gorm:"type:text"`
	Criteria    []byte    `json:"criteria" gorm:"type:json;not null"`
	IsPublic    bool      `json:"is_public" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName sets the table name
func (SavedSearch) TableName() string {
	return "saved_searches"
}

// SavedSearchResponse represents the API response format for saved searches
type SavedSearchResponse struct {
	ID          uint                   `json:"id"`
	Name        string                 `json:"name"`
	Description *string                `json:"description,omitempty"`
	Criteria    map[string]interface{} `json:"criteria"`
	IsPublic    bool                   `json:"is_public"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// SavedSearchRequest represents a request to save a search
type SavedSearchRequest struct {
	Name        string                 `json:"name" binding:"required,max=100"`
	Description *string                `json:"description,omitempty"`
	Criteria    map[string]interface{} `json:"criteria" binding:"required"`
	IsPublic    bool                   `json:"is_public"`
}