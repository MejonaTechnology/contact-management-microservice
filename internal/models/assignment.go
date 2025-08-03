package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// AssignmentRuleType represents the type of assignment rule
type AssignmentRuleType string

const (
	AssignmentRuleRoundRobin     AssignmentRuleType = "round_robin"
	AssignmentRuleLoadBased      AssignmentRuleType = "load_based"
	AssignmentRuleSkillBased     AssignmentRuleType = "skill_based"
	AssignmentRuleGeographyBased AssignmentRuleType = "geography_based"
	AssignmentRuleValueBased     AssignmentRuleType = "value_based"
	AssignmentRuleCustom         AssignmentRuleType = "custom"
)

// AssignmentRuleStatus represents the status of an assignment rule
type AssignmentRuleStatus string

const (
	AssignmentRuleActive   AssignmentRuleStatus = "active"
	AssignmentRuleInactive AssignmentRuleStatus = "inactive"
	AssignmentRulePaused   AssignmentRuleStatus = "paused"
)

// AssignmentCondition represents conditions for assignment rules
type AssignmentCondition struct {
	Field     string      `json:"field"`
	Operator  string      `json:"operator"` // equals, contains, greater_than, less_than, in, not_in
	Value     interface{} `json:"value"`
	ValueType string      `json:"value_type"` // string, number, boolean, array
}

// AssignmentConditions is a slice of conditions that can be serialized to JSON
type AssignmentConditions []AssignmentCondition

// Value implements the driver Valuer interface for database storage
func (ac AssignmentConditions) Value() (driver.Value, error) {
	if ac == nil {
		return nil, nil
	}
	return json.Marshal(ac)
}

// Scan implements the sql Scanner interface for database retrieval
func (ac *AssignmentConditions) Scan(value interface{}) error {
	if value == nil {
		*ac = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into AssignmentConditions", value)
	}
	return json.Unmarshal(bytes, ac)
}

// AssignmentRule represents rules for automatically assigning contacts to users
type AssignmentRule struct {
	ID          uint                   `json:"id" gorm:"primaryKey"`
	Name        string                 `json:"name" gorm:"size:255;not null"`
	Description *string                `json:"description" gorm:"type:text"`
	Type        AssignmentRuleType     `json:"type" gorm:"not null;index"`
	Status      AssignmentRuleStatus   `json:"status" gorm:"default:active;index"`
	Priority    int                    `json:"priority" gorm:"default:0;index"` // Higher number = higher priority
	
	// Rule Configuration
	Conditions  AssignmentConditions   `json:"conditions" gorm:"type:json"`
	Settings    JSONMap                `json:"settings" gorm:"type:json"` // Rule-specific settings
	
	// Assignment Targets
	AssigneeIDs JSONArray              `json:"assignee_ids" gorm:"type:json"` // Array of user IDs
	FallbackUserID *uint               `json:"fallback_user_id"`
	
	// Business Hours and Availability
	BusinessHoursEnabled bool         `json:"business_hours_enabled" gorm:"default:false"`
	BusinessHoursStart   *string      `json:"business_hours_start" gorm:"size:5"` // "09:00"
	BusinessHoursEnd     *string      `json:"business_hours_end" gorm:"size:5"`   // "17:00"
	WorkingDays          JSONArray    `json:"working_days" gorm:"type:json"`      // ["mon","tue","wed","thu","fri"]
	Timezone             string       `json:"timezone" gorm:"size:50;default:UTC"`
	
	// Rate Limiting
	MaxAssignmentsPerHour *int        `json:"max_assignments_per_hour"`
	MaxAssignmentsPerDay  *int        `json:"max_assignments_per_day"`
	
	// Tracking
	TotalAssignments     int          `json:"total_assignments" gorm:"default:0"`
	SuccessfulAssignments int         `json:"successful_assignments" gorm:"default:0"`
	LastAssignmentAt     *time.Time   `json:"last_assignment_at"`
	
	// Audit Fields
	CreatedAt            time.Time    `json:"created_at"`
	UpdatedAt            time.Time    `json:"updated_at"`
	CreatedBy            *uint        `json:"created_by"`
	UpdatedBy            *uint        `json:"updated_by"`
	DeletedAt            *time.Time   `json:"deleted_at" gorm:"index"`
}

// TableName specifies the table name for AssignmentRule
func (AssignmentRule) TableName() string {
	return "assignment_rules"
}

// ContactAssignment represents the assignment of a contact to a user
type ContactAssignment struct {
	ID           uint                  `json:"id" gorm:"primaryKey"`
	ContactID    uint                  `json:"contact_id" gorm:"not null;index"`
	AssignedToID uint                  `json:"assigned_to_id" gorm:"not null;index"`
	AssignedByID *uint                 `json:"assigned_by_id"` // Null for automatic assignments
	RuleID       *uint                 `json:"rule_id"`        // Which rule triggered this assignment
	
	// Assignment Details
	AssignmentType   string            `json:"assignment_type" gorm:"size:50;default:automatic"` // automatic, manual
	AssignmentReason string            `json:"assignment_reason" gorm:"type:text"`
	Priority         ContactPriority   `json:"priority" gorm:"default:medium"`
	
	// Status Tracking
	Status           string            `json:"status" gorm:"size:50;default:active;index"` // active, reassigned, completed, cancelled
	AcceptedAt       *time.Time        `json:"accepted_at"`
	FirstResponseAt  *time.Time        `json:"first_response_at"`
	CompletedAt      *time.Time        `json:"completed_at"`
	
	// Performance Metrics
	ResponseTimeHours    int            `json:"response_time_hours" gorm:"default:0"`
	ResolutionTimeHours  int            `json:"resolution_time_hours" gorm:"default:0"`
	InteractionCount     int            `json:"interaction_count" gorm:"default:0"`
	
	// Audit Fields
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	
	// Relationships
	Contact              *Contact       `json:"contact,omitempty" gorm:"foreignKey:ContactID"`
	AssignedTo           *AdminUser     `json:"assigned_to,omitempty" gorm:"foreignKey:AssignedToID"`
	AssignedBy           *AdminUser     `json:"assigned_by,omitempty" gorm:"foreignKey:AssignedByID"`
	Rule                 *AssignmentRule `json:"rule,omitempty" gorm:"foreignKey:RuleID"`
}

// TableName specifies the table name for ContactAssignment
func (ContactAssignment) TableName() string {
	return "contact_assignments"
}

// UserWorkload represents the current workload of a user for assignment purposes
type UserWorkload struct {
	ID                    uint       `json:"id" gorm:"primaryKey"`
	UserID                uint       `json:"user_id" gorm:"not null;uniqueIndex;index"`
	
	// Current Workload
	ActiveContacts        int        `json:"active_contacts" gorm:"default:0"`
	PendingContacts       int        `json:"pending_contacts" gorm:"default:0"`
	OverdueContacts       int        `json:"overdue_contacts" gorm:"default:0"`
	TotalContacts         int        `json:"total_contacts" gorm:"default:0"`
	
	// Daily Metrics
	TodayAssignments      int        `json:"today_assignments" gorm:"default:0"`
	TodayResponses        int        `json:"today_responses" gorm:"default:0"`
	TodayCompletions      int        `json:"today_completions" gorm:"default:0"`
	
	// Weekly Metrics
	WeeklyAssignments     int        `json:"weekly_assignments" gorm:"default:0"`
	WeeklyResponses       int        `json:"weekly_responses" gorm:"default:0"`
	WeeklyCompletions     int        `json:"weekly_completions" gorm:"default:0"`
	
	// Performance Metrics
	AvgResponseTimeHours  float64    `json:"avg_response_time_hours" gorm:"type:decimal(8,2);default:0.00"`
	AvgResolutionTimeHours float64   `json:"avg_resolution_time_hours" gorm:"type:decimal(8,2);default:0.00"`
	ConversionRate        float64    `json:"conversion_rate" gorm:"type:decimal(5,4);default:0.0000"`
	
	// Availability
	IsAvailable           bool       `json:"is_available" gorm:"default:true"`
	MaxDailyAssignments   *int       `json:"max_daily_assignments"`
	MaxActiveContacts     *int       `json:"max_active_contacts"`
	
	// Skills and Specialties
	Skills                JSONArray  `json:"skills" gorm:"type:json"`
	Territories           JSONArray  `json:"territories" gorm:"type:json"`
	ContactTypes          JSONArray  `json:"contact_types" gorm:"type:json"`
	
	// Audit Fields
	LastCalculatedAt      time.Time  `json:"last_calculated_at"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
	
	// Relationships
	User                  *AdminUser `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for UserWorkload
func (UserWorkload) TableName() string {
	return "user_workloads"
}

// AssignmentHistory represents the history of contact assignments
type AssignmentHistory struct {
	ID               uint                  `json:"id" gorm:"primaryKey"`
	ContactID        uint                  `json:"contact_id" gorm:"not null;index"`
	FromUserID       *uint                 `json:"from_user_id"` // Null for initial assignments
	ToUserID         uint                  `json:"to_user_id" gorm:"not null;index"`
	ChangedByID      *uint                 `json:"changed_by_id"`
	RuleID           *uint                 `json:"rule_id"`
	
	// Change Details
	ChangeType       string                `json:"change_type" gorm:"size:50;not null"` // assigned, reassigned, unassigned
	ChangeReason     string                `json:"change_reason" gorm:"type:text"`
	PreviousStatus   string                `json:"previous_status" gorm:"size:50"`
	NewStatus        string                `json:"new_status" gorm:"size:50"`
	
	// Context
	BusinessContext  JSONMap               `json:"business_context" gorm:"type:json"`
	SystemContext    JSONMap               `json:"system_context" gorm:"type:json"`
	
	// Audit Fields
	CreatedAt        time.Time             `json:"created_at"`
	
	// Relationships
	Contact          *Contact              `json:"contact,omitempty" gorm:"foreignKey:ContactID"`
	FromUser         *AdminUser            `json:"from_user,omitempty" gorm:"foreignKey:FromUserID"`
	ToUser           *AdminUser            `json:"to_user,omitempty" gorm:"foreignKey:ToUserID"`
	ChangedBy        *AdminUser            `json:"changed_by,omitempty" gorm:"foreignKey:ChangedByID"`
	Rule             *AssignmentRule       `json:"rule,omitempty" gorm:"foreignKey:RuleID"`
}

// TableName specifies the table name for AssignmentHistory
func (AssignmentHistory) TableName() string {
	return "assignment_history"
}

// Request/Response types

// AssignmentRuleRequest represents the request structure for creating/updating assignment rules
type AssignmentRuleRequest struct {
	Name                 string                 `json:"name" binding:"required,min=3,max=255"`
	Description          *string                `json:"description"`
	Type                 AssignmentRuleType     `json:"type" binding:"required"`
	Status               *AssignmentRuleStatus  `json:"status"`
	Priority             *int                   `json:"priority"`
	Conditions           AssignmentConditions   `json:"conditions"`
	Settings             JSONMap                `json:"settings"`
	AssigneeIDs          JSONArray              `json:"assignee_ids" binding:"required"`
	FallbackUserID       *uint                  `json:"fallback_user_id"`
	BusinessHoursEnabled *bool                  `json:"business_hours_enabled"`
	BusinessHoursStart   *string                `json:"business_hours_start"`
	BusinessHoursEnd     *string                `json:"business_hours_end"`
	WorkingDays          JSONArray              `json:"working_days"`
	Timezone             *string                `json:"timezone"`
	MaxAssignmentsPerHour *int                  `json:"max_assignments_per_hour"`
	MaxAssignmentsPerDay  *int                  `json:"max_assignments_per_day"`
}

// ContactAssignmentRequest represents the request structure for manual contact assignment
type ContactAssignmentRequest struct {
	ContactID        uint            `json:"contact_id" binding:"required"`
	AssignedToID     uint            `json:"assigned_to_id" binding:"required"`
	AssignmentReason string          `json:"assignment_reason"`
	Priority         ContactPriority `json:"priority"`
}

// BulkAssignmentRequest represents the request structure for bulk contact assignment
type BulkAssignmentRequest struct {
	ContactIDs       []uint          `json:"contact_ids" binding:"required,min=1"`
	AssignedToID     uint            `json:"assigned_to_id" binding:"required"`
	AssignmentReason string          `json:"assignment_reason"`
	Priority         ContactPriority `json:"priority"`
}

// UserWorkloadResponse represents the response structure for user workload
type UserWorkloadResponse struct {
	UserID                uint                  `json:"user_id"`
	User                  *AdminUserResponse    `json:"user,omitempty"`
	ActiveContacts        int                   `json:"active_contacts"`
	PendingContacts       int                   `json:"pending_contacts"`
	OverdueContacts       int                   `json:"overdue_contacts"`
	TotalContacts         int                   `json:"total_contacts"`
	TodayAssignments      int                   `json:"today_assignments"`
	TodayResponses        int                   `json:"today_responses"`
	TodayCompletions      int                   `json:"today_completions"`
	AvgResponseTimeHours  float64               `json:"avg_response_time_hours"`
	AvgResolutionTimeHours float64              `json:"avg_resolution_time_hours"`
	ConversionRate        float64               `json:"conversion_rate"`
	IsAvailable           bool                  `json:"is_available"`
	MaxDailyAssignments   *int                  `json:"max_daily_assignments"`
	MaxActiveContacts     *int                  `json:"max_active_contacts"`
	Skills                JSONArray             `json:"skills"`
	Territories           JSONArray             `json:"territories"`
	ContactTypes          JSONArray             `json:"contact_types"`
	LastCalculatedAt      time.Time             `json:"last_calculated_at"`
	WorkloadScore         float64               `json:"workload_score"` // Computed field
	AvailabilityScore     float64               `json:"availability_score"` // Computed field
}