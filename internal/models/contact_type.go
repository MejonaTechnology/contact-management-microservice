package models

import (
	"time"
)

// ContactType represents different types of contact inquiries
type ContactType struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"column:name;size:100;not null;uniqueIndex" binding:"required,min=2,max=100"`
	Description *string   `json:"description" gorm:"column:description;type:text"`
	Color       string    `json:"color" gorm:"column:color;size:7;default:#007bff" binding:"omitempty,hexcolor"`
	Icon        string    `json:"icon" gorm:"column:icon;size:50;default:contact"`
	IsActive    bool      `json:"is_active" gorm:"column:is_active;default:true"`
	SortOrder   int       `json:"sort_order" gorm:"column:sort_order;default:0"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// TableName specifies the table name for ContactType
func (ContactType) TableName() string {
	return "contact_types"
}

// BeforeCreate sets default values before creating
func (ct *ContactType) BeforeCreate() error {
	if ct.Color == "" {
		ct.Color = "#007bff"
	}
	if ct.Icon == "" {
		ct.Icon = "contact"
	}
	return nil
}

// ContactTypeRequest represents the request structure for creating/updating contact types
type ContactTypeRequest struct {
	Name        string  `json:"name" binding:"required,min=2,max=100"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
	Color       string  `json:"color" binding:"omitempty,hexcolor"`
	Icon        string  `json:"icon" binding:"omitempty,max=50"`
	IsActive    *bool   `json:"is_active"`
	SortOrder   *int    `json:"sort_order"`
}

// ContactTypeResponse represents the response structure for contact types
type ContactTypeResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Color       string    `json:"color"`
	Icon        string    `json:"icon"`
	IsActive    bool      `json:"is_active"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	UsageCount  int       `json:"usage_count,omitempty"` // For analytics
}