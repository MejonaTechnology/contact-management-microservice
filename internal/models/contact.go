package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// ContactStatus represents the status of a contact in the sales pipeline
type ContactStatus string

const (
	StatusNew          ContactStatus = "new"
	StatusContacted    ContactStatus = "contacted"
	StatusQualified    ContactStatus = "qualified"
	StatusProposal     ContactStatus = "proposal"
	StatusNegotiation  ContactStatus = "negotiation"
	StatusClosedWon    ContactStatus = "closed_won"
	StatusClosedLost   ContactStatus = "closed_lost"
	StatusOnHold       ContactStatus = "on_hold"
	StatusNurturing    ContactStatus = "nurturing"
)

// ContactPriority represents the priority level of a contact
type ContactPriority string

const (
	PriorityLow    ContactPriority = "low"
	PriorityMedium ContactPriority = "medium"
	PriorityHigh   ContactPriority = "high"
	PriorityUrgent ContactPriority = "urgent"
)

// PreferredContactMethod represents how the contact prefers to be contacted
type PreferredContactMethod string

const (
	ContactMethodEmail    PreferredContactMethod = "email"
	ContactMethodPhone    PreferredContactMethod = "phone"
	ContactMethodSMS      PreferredContactMethod = "sms"
	ContactMethodWhatsApp PreferredContactMethod = "whatsapp"
)

// JSONMap is a custom type for JSON fields
type JSONMap map[string]interface{}

// Value implements the driver Valuer interface
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql Scanner interface
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONMap", value)
	}
	return json.Unmarshal(bytes, j)
}

// JSONArray is a custom type for JSON array fields
type JSONArray []interface{}

// Value implements the driver Valuer interface
func (j JSONArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql Scanner interface
func (j *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONArray", value)
	}
	return json.Unmarshal(bytes, j)
}

// Contact represents the main contact entity with comprehensive contact management
type Contact struct {
	ID                    uint                   `json:"id" gorm:"primaryKey"`
	
	// Basic Information
	FirstName             string                 `json:"first_name" gorm:"column:first_name;size:100;not null" binding:"required,min=2,max=100"`
	LastName              *string                `json:"last_name" gorm:"column:last_name;size:100"`
	Email                 string                 `json:"email" gorm:"column:email;size:255;not null;index" binding:"required,email"`
	Phone                 *string                `json:"phone" gorm:"column:phone;size:20;index"`
	Company               *string                `json:"company" gorm:"column:company;size:200"`
	JobTitle              *string                `json:"job_title" gorm:"column:job_title;size:100"`
	Website               *string                `json:"website" gorm:"column:website;size:255"`
	
	// Address Information
	AddressLine1          *string                `json:"address_line1" gorm:"column:address_line1;size:255"`
	AddressLine2          *string                `json:"address_line2" gorm:"column:address_line2;size:255"`
	City                  *string                `json:"city" gorm:"column:city;size:100"`
	State                 *string                `json:"state" gorm:"column:state;size:100"`
	PostalCode            *string                `json:"postal_code" gorm:"column:postal_code;size:20"`
	Country               string                 `json:"country" gorm:"column:country;size:100;default:India"`
	
	// Contact Details
	ContactTypeID         uint                   `json:"contact_type_id" gorm:"column:contact_type_id;not null"`
	ContactSourceID       uint                   `json:"contact_source_id" gorm:"column:contact_source_id;not null"`
	Subject               *string                `json:"subject" gorm:"column:subject;size:500"`
	Message               *string                `json:"message" gorm:"column:message;type:text"`
	PreferredContactMethod PreferredContactMethod `json:"preferred_contact_method" gorm:"column:preferred_contact_method;default:email"`
	
	// Lead Management
	Status                ContactStatus          `json:"status" gorm:"column:status;default:new;index"`
	Priority              ContactPriority        `json:"priority" gorm:"column:priority;default:medium;index"`
	LeadScore             int                    `json:"lead_score" gorm:"column:lead_score;default:0;index"`
	EstimatedValue        float64                `json:"estimated_value" gorm:"column:estimated_value;type:decimal(12,2);default:0.00"`
	Probability           int                    `json:"probability" gorm:"column:probability;default:0"`
	
	// Assignment and Ownership
	AssignedTo            *uint                  `json:"assigned_to" gorm:"column:assigned_to;index"`
	AssignedAt            *time.Time             `json:"assigned_at" gorm:"column:assigned_at"`
	AssignedBy            *uint                  `json:"assigned_by" gorm:"column:assigned_by"`
	
	// Communication Tracking
	LastContactDate       *time.Time             `json:"last_contact_date" gorm:"column:last_contact_date"`
	NextFollowupDate      *time.Time             `json:"next_followup_date" gorm:"column:next_followup_date;index"`
	ResponseTimeHours     int                    `json:"response_time_hours" gorm:"column:response_time_hours;default:0"`
	TotalInteractions     int                    `json:"total_interactions" gorm:"column:total_interactions;default:0"`
	EmailOpened           bool                   `json:"email_opened" gorm:"column:email_opened;default:false"`
	EmailClicked          bool                   `json:"email_clicked" gorm:"column:email_clicked;default:false"`
	
	// Lifecycle Tracking
	FirstContactDate      time.Time              `json:"first_contact_date" gorm:"column:first_contact_date;default:CURRENT_TIMESTAMP;index"`
	LastActivityDate      time.Time              `json:"last_activity_date" gorm:"column:last_activity_date;default:CURRENT_TIMESTAMP;index"`
	ConversionDate        *time.Time             `json:"conversion_date" gorm:"column:conversion_date"`
	ClosedDate            *time.Time             `json:"closed_date" gorm:"column:closed_date"`
	
	// Technical Fields
	IPAddress             *string                `json:"ip_address" gorm:"column:ip_address;size:45"`
	UserAgent             *string                `json:"user_agent" gorm:"column:user_agent;type:text"`
	ReferrerURL           *string                `json:"referrer_url" gorm:"column:referrer_url;size:500"`
	LandingPage           *string                `json:"landing_page" gorm:"column:landing_page;size:500"`
	UTMSource             *string                `json:"utm_source" gorm:"column:utm_source;size:100;index"`
	UTMMedium             *string                `json:"utm_medium" gorm:"column:utm_medium;size:100"`
	UTMCampaign           *string                `json:"utm_campaign" gorm:"column:utm_campaign;size:100"`
	UTMTerm               *string                `json:"utm_term" gorm:"column:utm_term;size:100"`
	UTMContent            *string                `json:"utm_content" gorm:"column:utm_content;size:100"`
	
	// Data Management
	IsVerified            bool                   `json:"is_verified" gorm:"column:is_verified;default:false"`
	IsDuplicate           bool                   `json:"is_duplicate" gorm:"column:is_duplicate;default:false;index"`
	OriginalContactID     *uint                  `json:"original_contact_id" gorm:"column:original_contact_id"`
	DataSource            string                 `json:"data_source" gorm:"column:data_source;size:100;default:form"`
	
	// Privacy and Compliance
	MarketingConsent      bool                   `json:"marketing_consent" gorm:"column:marketing_consent;default:false"`
	DataProcessingConsent bool                   `json:"data_processing_consent" gorm:"column:data_processing_consent;default:true"`
	GDPRConsent           bool                   `json:"gdpr_consent" gorm:"column:gdpr_consent;default:false"`
	Unsubscribed          bool                   `json:"unsubscribed" gorm:"column:unsubscribed;default:false"`
	DoNotCall             bool                   `json:"do_not_call" gorm:"column:do_not_call;default:false"`
	
	// Metadata
	Tags                  JSONMap                `json:"tags" gorm:"column:tags;type:json"`
	CustomFields          JSONMap                `json:"custom_fields" gorm:"column:custom_fields;type:json"`
	Notes                 *string                `json:"notes" gorm:"column:notes;type:text"`
	
	// Audit Fields
	CreatedAt             time.Time              `json:"created_at" gorm:"column:created_at;index"`
	UpdatedAt             time.Time              `json:"updated_at" gorm:"column:updated_at"`
	CreatedBy             *uint                  `json:"created_by" gorm:"column:created_by"`
	UpdatedBy             *uint                  `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt             *time.Time             `json:"deleted_at" gorm:"column:deleted_at;index"`
	
	// Relationships
	ContactType           *ContactType           `json:"contact_type,omitempty" gorm:"foreignKey:ContactTypeID"`
	ContactSource         *ContactSource         `json:"contact_source,omitempty" gorm:"foreignKey:ContactSourceID"`
	Activities            []ContactActivity      `json:"activities,omitempty" gorm:"foreignKey:ContactID"`
	TagAssignments        []ContactTagAssignment `json:"tag_assignments,omitempty" gorm:"foreignKey:ContactID"`
	// Communications        []ContactCommunication `json:"communications,omitempty" gorm:"foreignKey:ContactID"`
	Appointments          []Appointment          `json:"appointments,omitempty" gorm:"foreignKey:ContactID"`
}

// TableName specifies the table name for Contact
func (Contact) TableName() string {
	return "contacts"
}

// BeforeCreate sets default values before creating
func (c *Contact) BeforeCreate() error {
	now := time.Now()
	c.FirstContactDate = now
	c.LastActivityDate = now
	if c.Country == "" {
		c.Country = "India"
	}
	return nil
}

// BeforeUpdate updates the last activity date
func (c *Contact) BeforeUpdate() error {
	c.LastActivityDate = time.Now()
	return nil
}

// GetFullName returns the full name of the contact
func (c *Contact) GetFullName() string {
	if c.LastName != nil && *c.LastName != "" {
		return c.FirstName + " " + *c.LastName
	}
	return c.FirstName
}

// GetDisplayName returns a display-friendly name for the contact
func (c *Contact) GetDisplayName() string {
	name := c.GetFullName()
	if c.Company != nil && *c.Company != "" {
		return fmt.Sprintf("%s (%s)", name, *c.Company)
	}
	return name
}

// IsHighPriority checks if the contact is high priority
func (c *Contact) IsHighPriority() bool {
	return c.Priority == PriorityHigh || c.Priority == PriorityUrgent
}

// IsHotLead checks if the contact is a hot lead (high lead score)
func (c *Contact) IsHotLead() bool {
	return c.LeadScore >= 80
}

// DaysInStatus returns the number of days the contact has been in current status
func (c *Contact) DaysInStatus() int {
	return int(time.Since(c.UpdatedAt).Hours() / 24)
}

// ContactRequest represents the request structure for creating/updating contacts
type ContactRequest struct {
	FirstName             string                 `json:"first_name" binding:"required,min=2,max=100"`
	LastName              *string                `json:"last_name" binding:"omitempty,max=100"`
	Email                 string                 `json:"email" binding:"required,email"`
	Phone                 *string                `json:"phone" binding:"omitempty,e164"`
	Company               *string                `json:"company" binding:"omitempty,max=200"`
	JobTitle              *string                `json:"job_title" binding:"omitempty,max=100"`
	Website               *string                `json:"website" binding:"omitempty,url"`
	AddressLine1          *string                `json:"address_line1" binding:"omitempty,max=255"`
	AddressLine2          *string                `json:"address_line2" binding:"omitempty,max=255"`
	City                  *string                `json:"city" binding:"omitempty,max=100"`
	State                 *string                `json:"state" binding:"omitempty,max=100"`
	PostalCode            *string                `json:"postal_code" binding:"omitempty,max=20"`
	Country               *string                `json:"country" binding:"omitempty,max=100"`
	ContactTypeID         uint                   `json:"contact_type_id" binding:"required,min=1"`
	ContactSourceID       uint                   `json:"contact_source_id" binding:"required,min=1"`
	Subject               *string                `json:"subject" binding:"omitempty,max=500"`
	Message               *string                `json:"message" binding:"omitempty,max=5000"`
	PreferredContactMethod *PreferredContactMethod `json:"preferred_contact_method"`
	Priority              *ContactPriority       `json:"priority"`
	EstimatedValue        *float64               `json:"estimated_value" binding:"omitempty,min=0"`
	AssignedTo            *uint                  `json:"assigned_to"`
	NextFollowupDate      *time.Time             `json:"next_followup_date"`
	MarketingConsent      *bool                  `json:"marketing_consent"`
	DataProcessingConsent *bool                  `json:"data_processing_consent"`
	GDPRConsent           *bool                  `json:"gdpr_consent"`
	Tags                  JSONMap                `json:"tags"`
	CustomFields          JSONMap                `json:"custom_fields"`
	Notes                 *string                `json:"notes" binding:"omitempty,max=5000"`
}

// PublicContactRequest represents a simplified request for public contact forms
type PublicContactRequest struct {
	FirstName        string  `json:"first_name" binding:"required,min=2,max=100"`
	LastName         *string `json:"last_name" binding:"omitempty,max=100"`
	Email            string  `json:"email" binding:"required,email"`
	Phone            *string `json:"phone"`
	Company          *string `json:"company" binding:"omitempty,max=200"`
	Subject          *string `json:"subject" binding:"omitempty,max=500"`
	Message          string  `json:"message" binding:"required,min=10,max=5000"`
	ContactTypeID    *uint   `json:"contact_type_id"`
	ContactSourceID  *uint   `json:"contact_source_id"`
	MarketingConsent *bool   `json:"marketing_consent"`
	// Honeypot field for spam detection
	Website          string  `json:"website"` // Should be empty for real users
}

// ContactResponse represents the response structure for contacts
type ContactResponse struct {
	ID                    uint                   `json:"id"`
	FirstName             string                 `json:"first_name"`
	LastName              *string                `json:"last_name"`
	FullName              string                 `json:"full_name"`
	DisplayName           string                 `json:"display_name"`
	Email                 string                 `json:"email"`
	Phone                 *string                `json:"phone"`
	Company               *string                `json:"company"`
	JobTitle              *string                `json:"job_title"`
	Website               *string                `json:"website"`
	Country               string                 `json:"country"`
	ContactType           *ContactTypeResponse   `json:"contact_type,omitempty"`
	ContactSource         *ContactSourceResponse `json:"contact_source,omitempty"`
	Subject               *string                `json:"subject"`
	Status                ContactStatus          `json:"status"`
	Priority              ContactPriority        `json:"priority"`
	LeadScore             int                    `json:"lead_score"`
	EstimatedValue        float64                `json:"estimated_value"`
	AssignedTo            *uint                  `json:"assigned_to"`
	AssignedAt            *time.Time             `json:"assigned_at"`
	LastContactDate       *time.Time             `json:"last_contact_date"`
	NextFollowupDate      *time.Time             `json:"next_followup_date"`
	ResponseTimeHours     int                    `json:"response_time_hours"`
	TotalInteractions     int                    `json:"total_interactions"`
	FirstContactDate      time.Time              `json:"first_contact_date"`
	LastActivityDate      time.Time              `json:"last_activity_date"`
	Tags                  JSONMap                `json:"tags"`
	CustomFields          JSONMap                `json:"custom_fields"`
	Notes                 *string                `json:"notes"`
	CreatedAt             time.Time              `json:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at"`
	// Computed fields
	DaysInStatus          int                    `json:"days_in_status"`
	IsHighPriority        bool                   `json:"is_high_priority"`
	IsHotLead             bool                   `json:"is_hot_lead"`
}