package models

import (
	"time"
)

// ContactSource represents different sources where contacts come from
type ContactSource struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	Name           string    `json:"name" gorm:"column:name;size:100;not null;uniqueIndex" binding:"required,min=2,max=100"`
	Description    *string   `json:"description" gorm:"column:description;type:text"`
	UTMSource      *string   `json:"utm_source" gorm:"column:utm_source;size:100"`
	UTMMedium      *string   `json:"utm_medium" gorm:"column:utm_medium;size:100"`
	UTMCampaign    *string   `json:"utm_campaign" gorm:"column:utm_campaign;size:100"`
	ConversionRate float64   `json:"conversion_rate" gorm:"column:conversion_rate;type:decimal(5,2);default:0.00"`
	CostPerLead    float64   `json:"cost_per_lead" gorm:"column:cost_per_lead;type:decimal(10,2);default:0.00"`
	IsActive       bool      `json:"is_active" gorm:"column:is_active;default:true"`
	SortOrder      int       `json:"sort_order" gorm:"column:sort_order;default:0"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// TableName specifies the table name for ContactSource
func (ContactSource) TableName() string {
	return "contact_sources"
}

// ContactSourceRequest represents the request structure for creating/updating contact sources
type ContactSourceRequest struct {
	Name           string   `json:"name" binding:"required,min=2,max=100"`
	Description    *string  `json:"description" binding:"omitempty,max=1000"`
	UTMSource      *string  `json:"utm_source" binding:"omitempty,max=100"`
	UTMMedium      *string  `json:"utm_medium" binding:"omitempty,max=100"`
	UTMCampaign    *string  `json:"utm_campaign" binding:"omitempty,max=100"`
	ConversionRate *float64 `json:"conversion_rate" binding:"omitempty,min=0,max=100"`
	CostPerLead    *float64 `json:"cost_per_lead" binding:"omitempty,min=0"`
	IsActive       *bool    `json:"is_active"`
	SortOrder      *int     `json:"sort_order"`
}

// ContactSourceResponse represents the response structure for contact sources
type ContactSourceResponse struct {
	ID             uint      `json:"id"`
	Name           string    `json:"name"`
	Description    *string   `json:"description"`
	UTMSource      *string   `json:"utm_source"`
	UTMMedium      *string   `json:"utm_medium"`
	UTMCampaign    *string   `json:"utm_campaign"`
	ConversionRate float64   `json:"conversion_rate"`
	CostPerLead    float64   `json:"cost_per_lead"`
	IsActive       bool      `json:"is_active"`
	SortOrder      int       `json:"sort_order"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	// Analytics fields
	TotalContacts     int     `json:"total_contacts,omitempty"`
	ConvertedContacts int     `json:"converted_contacts,omitempty"`
	ActualROI         float64 `json:"actual_roi,omitempty"`
}