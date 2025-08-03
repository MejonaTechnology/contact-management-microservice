package models

import (
	"time"
)

// ActivityType represents different types of contact activities
type ActivityType string

const (
	ActivityStatusChange    ActivityType = "status_change"
	ActivityAssignment      ActivityType = "assignment"
	ActivityNoteAdded       ActivityType = "note_added"
	ActivityEmailSent       ActivityType = "email_sent"
	ActivityEmailReceived   ActivityType = "email_received"
	ActivityCallMade        ActivityType = "call_made"
	ActivityCallReceived    ActivityType = "call_received"
	ActivitySMSSent         ActivityType = "sms_sent"
	ActivityMeetingScheduled ActivityType = "meeting_scheduled"
	ActivityMeetingHeld     ActivityType = "meeting_held"
	ActivityProposalSent    ActivityType = "proposal_sent"
	ActivityContractSent    ActivityType = "contract_sent"
	ActivityPaymentReceived ActivityType = "payment_received"
	ActivityFollowUp        ActivityType = "follow_up"
	ActivityDocumentShared  ActivityType = "document_shared"
	ActivityQuoteSent       ActivityType = "quote_sent"
	ActivityDemoScheduled   ActivityType = "demo_scheduled"
	ActivityDemoCompleted   ActivityType = "demo_completed"
	ActivityLeadQualified   ActivityType = "lead_qualified"
	ActivityLeadScored      ActivityType = "lead_scored"
	ActivityConverted       ActivityType = "converted"
	ActivityLost            ActivityType = "lost"
	ActivityReopened        ActivityType = "reopened"
)

// ActivityStatus represents the status of an activity
type ActivityStatus string

const (
	ActivityStatusPending    ActivityStatus = "pending"
	ActivityStatusInProgress ActivityStatus = "in_progress"
	ActivityStatusCompleted  ActivityStatus = "completed"
	ActivityStatusCancelled  ActivityStatus = "cancelled"
)

// ActivityDirection represents the direction of communication
type ActivityDirection string

const (
	DirectionInbound  ActivityDirection = "inbound"
	DirectionOutbound ActivityDirection = "outbound"
	DirectionInternal ActivityDirection = "internal"
)

// ActivityChannel represents the communication channel
type ActivityChannel string

const (
	ChannelEmail     ActivityChannel = "email"
	ChannelPhone     ActivityChannel = "phone"
	ChannelSMS       ActivityChannel = "sms"
	ChannelWhatsApp  ActivityChannel = "whatsapp"
	ChannelChat      ActivityChannel = "chat"
	ChannelInPerson  ActivityChannel = "in_person"
	ChannelVideoCall ActivityChannel = "video_call"
	ChannelOther     ActivityChannel = "other"
)

// ContactActivity represents all interactions and activities with a contact
type ContactActivity struct {
	ID                  uint              `json:"id" gorm:"primaryKey"`
	ContactID           uint              `json:"contact_id" gorm:"column:contact_id;not null;index"`
	
	// Activity Information
	ActivityType        ActivityType      `json:"activity_type" gorm:"column:activity_type;not null;index"`
	Title               string            `json:"title" gorm:"column:title;size:255;not null" binding:"required,min=2,max=255"`
	Description         *string           `json:"description" gorm:"column:description;type:text"`
	Outcome             *string           `json:"outcome" gorm:"column:outcome;type:text"`
	
	// Activity Details
	ActivityDate        time.Time         `json:"activity_date" gorm:"column:activity_date;default:CURRENT_TIMESTAMP;index"`
	DurationMinutes     int               `json:"duration_minutes" gorm:"column:duration_minutes;default:0"`
	
	// Status and Priority
	Status              ActivityStatus    `json:"status" gorm:"column:status;default:completed;index"`
	Priority            ContactPriority   `json:"priority" gorm:"column:priority;default:medium"`
	
	// Related Information
	RelatedEntityType   *string           `json:"related_entity_type" gorm:"column:related_entity_type;size:50"`
	RelatedEntityID     *uint             `json:"related_entity_id" gorm:"column:related_entity_id"`
	ExternalReference   *string           `json:"external_reference" gorm:"column:external_reference;size:255"`
	
	// Communication Details
	Direction           ActivityDirection `json:"direction" gorm:"column:direction;default:outbound"`
	Channel             ActivityChannel   `json:"channel" gorm:"column:channel;default:email"`
	
	// Scheduling Information
	ScheduledDate       *time.Time        `json:"scheduled_date" gorm:"column:scheduled_date;index"`
	ReminderDate        *time.Time        `json:"reminder_date" gorm:"column:reminder_date;index"`
	CompletedDate       *time.Time        `json:"completed_date" gorm:"column:completed_date"`
	
	// User and Assignment
	PerformedBy         uint              `json:"performed_by" gorm:"column:performed_by;not null;index"`
	AssignedTo          *uint             `json:"assigned_to" gorm:"column:assigned_to;index"`
	
	// Tracking and Analytics
	IsBillable          bool              `json:"is_billable" gorm:"column:is_billable;default:false"`
	BillableAmount      float64           `json:"billable_amount" gorm:"column:billable_amount;type:decimal(10,2);default:0.00"`
	Cost                float64           `json:"cost" gorm:"column:cost;type:decimal(10,2);default:0.00"`
	
	// Metadata
	Tags                JSONMap           `json:"tags" gorm:"column:tags;type:json"`
	Metadata            JSONMap           `json:"metadata" gorm:"column:metadata;type:json"`
	Attachments         JSONMap           `json:"attachments" gorm:"column:attachments;type:json"`
	
	// Audit Fields
	CreatedAt           time.Time         `json:"created_at" gorm:"column:created_at;index"`
	UpdatedAt           time.Time         `json:"updated_at" gorm:"column:updated_at"`
	CreatedBy           *uint             `json:"created_by" gorm:"column:created_by"`
	UpdatedBy           *uint             `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt           *time.Time        `json:"deleted_at" gorm:"column:deleted_at;index"`
	
	// Relationships
	Contact             *Contact          `json:"contact,omitempty" gorm:"foreignKey:ContactID"`
}

// TableName specifies the table name for ContactActivity
func (ContactActivity) TableName() string {
	return "contact_activities"
}

// BeforeCreate sets default values before creating
func (ca *ContactActivity) BeforeCreate() error {
	if ca.ActivityDate.IsZero() {
		ca.ActivityDate = time.Now()
	}
	return nil
}

// IsOverdue checks if a scheduled activity is overdue
func (ca *ContactActivity) IsOverdue() bool {
	if ca.Status == ActivityStatusCompleted || ca.Status == ActivityStatusCancelled {
		return false
	}
	if ca.ScheduledDate == nil {
		return false
	}
	return time.Now().After(*ca.ScheduledDate)
}

// IsUpcoming checks if an activity is scheduled for the next hour
func (ca *ContactActivity) IsUpcoming() bool {
	if ca.Status == ActivityStatusCompleted || ca.Status == ActivityStatusCancelled {
		return false
	}
	if ca.ScheduledDate == nil {
		return false
	}
	nextHour := time.Now().Add(time.Hour)
	return ca.ScheduledDate.Before(nextHour) && ca.ScheduledDate.After(time.Now())
}

// ContactActivityRequest represents the request structure for creating/updating activities
type ContactActivityRequest struct {
	ContactID         uint              `json:"contact_id" binding:"required,min=1"`
	ActivityType      ActivityType      `json:"activity_type" binding:"required"`
	Title             string            `json:"title" binding:"required,min=2,max=255"`
	Description       *string           `json:"description" binding:"omitempty,max=5000"`
	Outcome           *string           `json:"outcome" binding:"omitempty,max=1000"`
	ActivityDate      *time.Time        `json:"activity_date"`
	DurationMinutes   *int              `json:"duration_minutes" binding:"omitempty,min=0,max=1440"`
	Status            *ActivityStatus   `json:"status"`
	Priority          *ContactPriority  `json:"priority"`
	Direction         *ActivityDirection `json:"direction"`
	Channel           *ActivityChannel  `json:"channel"`
	ScheduledDate     *time.Time        `json:"scheduled_date"`
	ReminderDate      *time.Time        `json:"reminder_date"`
	AssignedTo        *uint             `json:"assigned_to"`
	IsBillable        *bool             `json:"is_billable"`
	BillableAmount    *float64          `json:"billable_amount" binding:"omitempty,min=0"`
	Cost              *float64          `json:"cost" binding:"omitempty,min=0"`
	Tags              JSONMap           `json:"tags"`
	Metadata          JSONMap           `json:"metadata"`
	Attachments       JSONMap           `json:"attachments"`
}

// ContactActivityResponse represents the response structure for activities
type ContactActivityResponse struct {
	ID                uint              `json:"id"`
	ContactID         uint              `json:"contact_id"`
	ActivityType      ActivityType      `json:"activity_type"`
	Title             string            `json:"title"`
	Description       *string           `json:"description"`
	Outcome           *string           `json:"outcome"`
	ActivityDate      time.Time         `json:"activity_date"`
	DurationMinutes   int               `json:"duration_minutes"`
	Status            ActivityStatus    `json:"status"`
	Priority          ContactPriority   `json:"priority"`
	Direction         ActivityDirection `json:"direction"`
	Channel           ActivityChannel   `json:"channel"`
	ScheduledDate     *time.Time        `json:"scheduled_date"`
	ReminderDate      *time.Time        `json:"reminder_date"`
	CompletedDate     *time.Time        `json:"completed_date"`
	PerformedBy       uint              `json:"performed_by"`
	AssignedTo        *uint             `json:"assigned_to"`
	IsBillable        bool              `json:"is_billable"`
	BillableAmount    float64           `json:"billable_amount"`
	Cost              float64           `json:"cost"`
	Tags              JSONMap           `json:"tags"`
	Metadata          JSONMap           `json:"metadata"`
	Attachments       JSONMap           `json:"attachments"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	// Computed fields
	IsOverdue         bool              `json:"is_overdue"`
	IsUpcoming        bool              `json:"is_upcoming"`
	// Related data
	Contact           *ContactResponse  `json:"contact,omitempty"`
}