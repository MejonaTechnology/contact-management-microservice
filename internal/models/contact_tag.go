package models

import (
	"time"
)

// ContactTag represents tags for flexible contact categorization
type ContactTag struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"column:name;size:100;not null;uniqueIndex" binding:"required,min=2,max=100"`
	Description *string   `json:"description" gorm:"column:description;type:text"`
	Color       string    `json:"color" gorm:"column:color;size:7;default:#007bff" binding:"omitempty,hexcolor"`
	Category    *string   `json:"category" gorm:"column:category;size:50;index"`
	IsSystem    bool      `json:"is_system" gorm:"column:is_system;default:false"`
	UsageCount  int       `json:"usage_count" gorm:"column:usage_count;default:0;index"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at"`
	CreatedBy   *uint     `json:"created_by" gorm:"column:created_by"`
}

// TableName specifies the table name for ContactTag
func (ContactTag) TableName() string {
	return "contact_tags"
}

// BeforeCreate sets default values before creating
func (ct *ContactTag) BeforeCreate() error {
	if ct.Color == "" {
		ct.Color = "#007bff"
	}
	return nil
}

// ContactTagAssignment represents the many-to-many relationship between contacts and tags
type ContactTagAssignment struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	ContactID  uint      `json:"contact_id" gorm:"column:contact_id;not null;index"`
	TagID      uint      `json:"tag_id" gorm:"column:tag_id;not null;index"`
	AssignedAt time.Time `json:"assigned_at" gorm:"column:assigned_at;default:CURRENT_TIMESTAMP"`
	AssignedBy *uint     `json:"assigned_by" gorm:"column:assigned_by"`
	
	// Relationships
	Contact    *Contact    `json:"contact,omitempty" gorm:"foreignKey:ContactID"`
	Tag        *ContactTag `json:"tag,omitempty" gorm:"foreignKey:TagID"`
}

// TableName specifies the table name for ContactTagAssignment
func (ContactTagAssignment) TableName() string {
	return "contact_tag_assignments"
}

// ContactTagRequest represents the request structure for creating/updating tags
type ContactTagRequest struct {
	Name        string  `json:"name" binding:"required,min=2,max=100"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
	Color       string  `json:"color" binding:"omitempty,hexcolor"`
	Category    *string `json:"category" binding:"omitempty,max=50"`
}

// ContactTagResponse represents the response structure for tags
type ContactTagResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Color       string    `json:"color"`
	Category    *string   `json:"category"`
	IsSystem    bool      `json:"is_system"`
	UsageCount  int       `json:"usage_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ContactTagAssignmentRequest represents the request for assigning tags to contacts
type ContactTagAssignmentRequest struct {
	ContactID uint   `json:"contact_id" binding:"required,min=1"`
	TagIDs    []uint `json:"tag_ids" binding:"required,min=1,dive,min=1"`
}

// ContactTagAssignmentResponse represents the response for tag assignments
type ContactTagAssignmentResponse struct {
	ID         uint                `json:"id"`
	ContactID  uint                `json:"contact_id"`
	TagID      uint                `json:"tag_id"`
	AssignedAt time.Time           `json:"assigned_at"`
	Tag        *ContactTagResponse `json:"tag,omitempty"`
}