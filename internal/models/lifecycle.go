package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// LifecycleStage represents stages in the contact lifecycle
type LifecycleStage string

const (
	StageUnknown      LifecycleStage = "unknown"
	StageSuspect      LifecycleStage = "suspect"
	StageProspect     LifecycleStage = "prospect"
	StageMarketingQL  LifecycleStage = "marketing_qualified_lead"
	StageSalesQL      LifecycleStage = "sales_qualified_lead"
	StageOpportunity  LifecycleStage = "opportunity"
	StageCustomer     LifecycleStage = "customer"
	StageEvangelist   LifecycleStage = "evangelist"
	StageOther        LifecycleStage = "other"
)

// StatusTransitionType represents the type of status transition
type StatusTransitionType string

const (
	TransitionAutomatic StatusTransitionType = "automatic"
	TransitionManual    StatusTransitionType = "manual"
	TransitionScheduled StatusTransitionType = "scheduled"
	TransitionTriggered StatusTransitionType = "triggered"
)

// ScoringCriteria represents individual scoring criteria
type ScoringCriteria struct {
	Name        string      `json:"name"`
	Field       string      `json:"field"`
	Operator    string      `json:"operator"` // equals, contains, greater_than, less_than, in, not_in
	Value       interface{} `json:"value"`
	Score       int         `json:"score"`
	Weight      float64     `json:"weight"`
	Description string      `json:"description"`
}

// ScoringCriteriaList is a slice of scoring criteria that can be serialized to JSON
type ScoringCriteriaList []ScoringCriteria

// Value implements the driver Valuer interface for database storage
func (scl ScoringCriteriaList) Value() (driver.Value, error) {
	if scl == nil {
		return nil, nil
	}
	return json.Marshal(scl)
}

// Scan implements the sql Scanner interface for database retrieval
func (scl *ScoringCriteriaList) Scan(value interface{}) error {
	if value == nil {
		*scl = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into ScoringCriteriaList", value)
	}
	return json.Unmarshal(bytes, scl)
}

// LeadScoringRule represents rules for calculating lead scores
type LeadScoringRule struct {
	ID              uint                `json:"id" gorm:"primaryKey"`
	Name            string              `json:"name" gorm:"size:255;not null"`
	Description     *string             `json:"description" gorm:"type:text"`
	IsActive        bool                `json:"is_active" gorm:"default:true;index"`
	Priority        int                 `json:"priority" gorm:"default:0;index"`
	
	// Scoring Configuration
	Category        string              `json:"category" gorm:"size:100"` // demographic, behavioral, engagement, firmographic
	BaseScore       int                 `json:"base_score" gorm:"default:0"`
	MaxScore        int                 `json:"max_score" gorm:"default:100"`
	Criteria        ScoringCriteriaList `json:"criteria" gorm:"type:json"`
	
	// Conditions for when this rule applies
	ApplicableWhen  AssignmentConditions `json:"applicable_when" gorm:"type:json"`
	
	// Tracking
	TimesApplied    int                 `json:"times_applied" gorm:"default:0"`
	LastAppliedAt   *time.Time          `json:"last_applied_at"`
	
	// Audit Fields
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	CreatedBy       *uint               `json:"created_by"`
	UpdatedBy       *uint               `json:"updated_by"`
	DeletedAt       *time.Time          `json:"deleted_at" gorm:"index"`
}

// TableName specifies the table name for LeadScoringRule
func (LeadScoringRule) TableName() string {
	return "lead_scoring_rules"
}

// StatusTransitionRule represents rules for automatic status transitions
type StatusTransitionRule struct {
	ID              uint                    `json:"id" gorm:"primaryKey"`
	Name            string                  `json:"name" gorm:"size:255;not null"`
	Description     *string                 `json:"description" gorm:"type:text"`
	IsActive        bool                    `json:"is_active" gorm:"default:true;index"`
	Priority        int                     `json:"priority" gorm:"default:0;index"`
	
	// Transition Configuration
	FromStatus      ContactStatus           `json:"from_status" gorm:"not null"`
	ToStatus        ContactStatus           `json:"to_status" gorm:"not null"`
	TransitionType  StatusTransitionType    `json:"transition_type" gorm:"default:automatic"`
	
	// Conditions for transition
	Conditions      AssignmentConditions    `json:"conditions" gorm:"type:json"`
	RequiredScore   int                     `json:"required_score" gorm:"default:0"`
	DaysInStatus    int                     `json:"days_in_status" gorm:"default:0"`
	
	// Actions to perform on transition
	Actions         JSONMap                 `json:"actions" gorm:"type:json"`
	NotifyUsers     JSONArray               `json:"notify_users" gorm:"type:json"`
	
	// Tracking
	TimesTriggered  int                     `json:"times_triggered" gorm:"default:0"`
	LastTriggeredAt *time.Time              `json:"last_triggered_at"`
	
	// Audit Fields
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
	CreatedBy       *uint                   `json:"created_by"`
	UpdatedBy       *uint                   `json:"updated_by"`
	DeletedAt       *time.Time              `json:"deleted_at" gorm:"index"`
}

// TableName specifies the table name for StatusTransitionRule
func (StatusTransitionRule) TableName() string {
	return "status_transition_rules"
}

// ContactLifecycle represents the lifecycle tracking for a contact
type ContactLifecycle struct {
	ID              uint                    `json:"id" gorm:"primaryKey"`
	ContactID       uint                    `json:"contact_id" gorm:"not null;uniqueIndex;index"`
	
	// Current State
	CurrentStage    LifecycleStage          `json:"current_stage" gorm:"default:unknown;index"`
	CurrentStatus   ContactStatus           `json:"current_status" gorm:"index"`
	CurrentScore    int                     `json:"current_score" gorm:"default:0;index"`
	
	// Lifecycle Progression
	StageEnteredAt  time.Time               `json:"stage_entered_at" gorm:"default:CURRENT_TIMESTAMP"`
	StatusEnteredAt time.Time               `json:"status_entered_at" gorm:"default:CURRENT_TIMESTAMP"`
	LastScoredAt    time.Time               `json:"last_scored_at" gorm:"default:CURRENT_TIMESTAMP"`
	
	// Scoring Breakdown
	DemographicScore    int                 `json:"demographic_score" gorm:"default:0"`
	BehavioralScore     int                 `json:"behavioral_score" gorm:"default:0"`
	EngagementScore     int                 `json:"engagement_score" gorm:"default:0"`
	FirmographicScore   int                 `json:"firmographic_score" gorm:"default:0"`
	
	// Scoring History
	ScoreHistory        JSONArray           `json:"score_history" gorm:"type:json"`
	ScoringFactors      JSONMap             `json:"scoring_factors" gorm:"type:json"`
	
	// Milestones
	FirstEngagementAt   *time.Time          `json:"first_engagement_at"`
	QualificationAt     *time.Time          `json:"qualification_at"`
	OpportunityAt       *time.Time          `json:"opportunity_at"`
	ConversionAt        *time.Time          `json:"conversion_at"`
	
	// Performance Metrics
	DaysInCurrentStage  int                 `json:"days_in_current_stage" gorm:"default:0"`
	DaysInCurrentStatus int                 `json:"days_in_current_status" gorm:"default:0"`
	TotalLifecycleDays  int                 `json:"total_lifecycle_days" gorm:"default:0"`
	
	// Velocity Metrics
	StageVelocity       JSONMap             `json:"stage_velocity" gorm:"type:json"` // Days spent in each stage
	ConversionRate      float64             `json:"conversion_rate" gorm:"type:decimal(5,4);default:0.0000"`
	
	// Audit Fields
	CreatedAt           time.Time           `json:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at"`
	
	// Relationships
	Contact             *Contact            `json:"contact,omitempty" gorm:"foreignKey:ContactID"`
}

// TableName specifies the table name for ContactLifecycle
func (ContactLifecycle) TableName() string {
	return "contact_lifecycles"
}

// LifecycleEvent represents events in the contact lifecycle
type LifecycleEvent struct {
	ID              uint                    `json:"id" gorm:"primaryKey"`
	ContactID       uint                    `json:"contact_id" gorm:"not null;index"`
	LifecycleID     uint                    `json:"lifecycle_id" gorm:"not null;index"`
	
	// Event Details
	EventType       string                  `json:"event_type" gorm:"size:100;not null;index"` // score_change, status_change, stage_change, milestone, action
	EventName       string                  `json:"event_name" gorm:"size:255;not null"`
	EventDescription string                 `json:"event_description" gorm:"type:text"`
	
	// Before/After State
	PreviousValue   *string                 `json:"previous_value" gorm:"size:255"`
	NewValue        string                  `json:"new_value" gorm:"size:255;not null"`
	ChangeAmount    *int                    `json:"change_amount"` // For score changes
	
	// Context
	TriggerType     string                  `json:"trigger_type" gorm:"size:50"` // automatic, manual, scheduled, api
	TriggerSource   string                  `json:"trigger_source" gorm:"size:100"` // rule_id, user_id, system
	TriggerData     JSONMap                 `json:"trigger_data" gorm:"type:json"`
	
	// User who caused the event (if manual)
	TriggeredBy     *uint                   `json:"triggered_by"`
	
	// Audit Fields
	CreatedAt       time.Time               `json:"created_at"`
	
	// Relationships
	Contact         *Contact                `json:"contact,omitempty" gorm:"foreignKey:ContactID"`
	Lifecycle       *ContactLifecycle       `json:"lifecycle,omitempty" gorm:"foreignKey:LifecycleID"`
	TriggeredByUser *AdminUser              `json:"triggered_by_user,omitempty" gorm:"foreignKey:TriggeredBy"`
}

// TableName specifies the table name for LifecycleEvent
func (LifecycleEvent) TableName() string {
	return "lifecycle_events"
}

// ScoreSnapshot represents a snapshot of a contact's score at a point in time
type ScoreSnapshot struct {
	ContactID       uint        `json:"contact_id"`
	Timestamp       time.Time   `json:"timestamp"`
	TotalScore      int         `json:"total_score"`
	DemoScore       int         `json:"demo_score"`
	BehavioralScore int         `json:"behavioral_score"`
	EngagementScore int         `json:"engagement_score"`
	FirmoScore      int         `json:"firmo_score"`
	Factors         JSONMap     `json:"factors"`
}

// StatusTransition represents a status transition event
type StatusTransition struct {
	ContactID       uint                    `json:"contact_id"`
	FromStatus      ContactStatus           `json:"from_status"`
	ToStatus        ContactStatus           `json:"to_status"`
	TransitionType  StatusTransitionType    `json:"transition_type"`
	RuleID          *uint                   `json:"rule_id"`
	Reason          string                  `json:"reason"`
	TriggeredBy     *uint                   `json:"triggered_by"`
	Timestamp       time.Time               `json:"timestamp"`
}

// Request/Response types

// LeadScoringRuleRequest represents the request structure for creating/updating lead scoring rules
type LeadScoringRuleRequest struct {
	Name            string               `json:"name" binding:"required,min=3,max=255"`
	Description     *string              `json:"description"`
	IsActive        *bool                `json:"is_active"`
	Priority        *int                 `json:"priority"`
	Category        string               `json:"category" binding:"required"`
	BaseScore       *int                 `json:"base_score"`
	MaxScore        *int                 `json:"max_score"`
	Criteria        ScoringCriteriaList  `json:"criteria" binding:"required"`
	ApplicableWhen  AssignmentConditions `json:"applicable_when"`
}

// StatusTransitionRuleRequest represents the request structure for creating/updating status transition rules
type StatusTransitionRuleRequest struct {
	Name            string                  `json:"name" binding:"required,min=3,max=255"`
	Description     *string                 `json:"description"`
	IsActive        *bool                   `json:"is_active"`
	Priority        *int                    `json:"priority"`
	FromStatus      ContactStatus           `json:"from_status" binding:"required"`
	ToStatus        ContactStatus           `json:"to_status" binding:"required"`
	TransitionType  *StatusTransitionType   `json:"transition_type"`
	Conditions      AssignmentConditions    `json:"conditions"`
	RequiredScore   *int                    `json:"required_score"`
	DaysInStatus    *int                    `json:"days_in_status"`
	Actions         JSONMap                 `json:"actions"`
	NotifyUsers     JSONArray               `json:"notify_users"`
}

// ContactScoringRequest represents the request to manually score a contact
type ContactScoringRequest struct {
	ContactID   uint    `json:"contact_id" binding:"required"`
	ForceRescore bool   `json:"force_rescore"`
	Reason      string  `json:"reason"`
}

// StatusChangeRequest represents the request to manually change contact status
type StatusChangeRequest struct {
	ContactID   uint          `json:"contact_id" binding:"required"`
	NewStatus   ContactStatus `json:"new_status" binding:"required"`
	Reason      string        `json:"reason" binding:"required"`
	ForceChange bool          `json:"force_change"`
}

// BulkStatusChangeRequest represents the request to change status for multiple contacts
type BulkStatusChangeRequest struct {
	ContactIDs  []uint        `json:"contact_ids" binding:"required,min=1"`
	NewStatus   ContactStatus `json:"new_status" binding:"required"`
	Reason      string        `json:"reason" binding:"required"`
	ForceChange bool          `json:"force_change"`
}

// ContactLifecycleResponse represents the response structure for contact lifecycle
type ContactLifecycleResponse struct {
	ID                  uint                    `json:"id"`
	ContactID           uint                    `json:"contact_id"`
	Contact             *ContactResponse        `json:"contact,omitempty"`
	CurrentStage        LifecycleStage          `json:"current_stage"`
	CurrentStatus       ContactStatus           `json:"current_status"`
	CurrentScore        int                     `json:"current_score"`
	StageEnteredAt      time.Time               `json:"stage_entered_at"`
	StatusEnteredAt     time.Time               `json:"status_entered_at"`
	LastScoredAt        time.Time               `json:"last_scored_at"`
	DemographicScore    int                     `json:"demographic_score"`
	BehavioralScore     int                     `json:"behavioral_score"`
	EngagementScore     int                     `json:"engagement_score"`
	FirmographicScore   int                     `json:"firmographic_score"`
	ScoreHistory        JSONArray               `json:"score_history"`
	ScoringFactors      JSONMap                 `json:"scoring_factors"`
	FirstEngagementAt   *time.Time              `json:"first_engagement_at"`
	QualificationAt     *time.Time              `json:"qualification_at"`
	OpportunityAt       *time.Time              `json:"opportunity_at"`
	ConversionAt        *time.Time              `json:"conversion_at"`
	DaysInCurrentStage  int                     `json:"days_in_current_stage"`
	DaysInCurrentStatus int                     `json:"days_in_current_status"`
	TotalLifecycleDays  int                     `json:"total_lifecycle_days"`
	StageVelocity       JSONMap                 `json:"stage_velocity"`
	ConversionRate      float64                 `json:"conversion_rate"`
	CreatedAt           time.Time               `json:"created_at"`
	UpdatedAt           time.Time               `json:"updated_at"`
	// Computed fields
	ScoreGrade          string                  `json:"score_grade"` // A, B, C, D, F
	QualificationStatus string                  `json:"qualification_status"` // qualified, unqualified, pending
	NextSuggestedAction string                  `json:"next_suggested_action"`
}

// ScoringAnalysisResponse represents detailed scoring analysis
type ScoringAnalysisResponse struct {
	ContactID           uint                    `json:"contact_id"`
	TotalScore          int                     `json:"total_score"`
	MaxPossibleScore    int                     `json:"max_possible_score"`
	ScorePercentage     float64                 `json:"score_percentage"`
	Grade               string                  `json:"grade"`
	CategoryBreakdown   map[string]int          `json:"category_breakdown"`
	AppliedRules        []string                `json:"applied_rules"`
	ScoringFactors      map[string]interface{}  `json:"scoring_factors"`
	Recommendations     []string                `json:"recommendations"`
	LastUpdated         time.Time               `json:"last_updated"`
}