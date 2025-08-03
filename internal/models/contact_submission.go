package models

import (
	"strings"
	"time"
)

// ContactSubmission represents simplified contact submissions for dashboard compatibility
// This bridges the gap between the comprehensive CRM Contact model and dashboard's simple needs
type ContactSubmission struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"column:name;size:200;not null"`
	Email        string    `json:"email" gorm:"column:email;size:255;not null;index"`
	Phone        *string   `json:"phone" gorm:"column:phone;size:20"`
	Subject      *string   `json:"subject" gorm:"column:subject;size:500"`
	Message      string    `json:"message" gorm:"column:message;type:text;not null"`
	Source       *string   `json:"source" gorm:"column:source;size:100"`
	Status       string    `json:"status" gorm:"column:status;size:50;default:new;index"`
	AssignedTo   *int      `json:"assigned_to" gorm:"column:assigned_to"`
	ResponseSent bool      `json:"response_sent" gorm:"column:response_sent;default:false"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at;index"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// TableName specifies the table name for ContactSubmission
func (ContactSubmission) TableName() string {
	return "contact_submissions"
}

// ContactSubmissionRequest represents the request for creating contact submissions
type ContactSubmissionRequest struct {
	Name            string  `json:"name" binding:"required,min=2,max=200"`
	Email           string  `json:"email" binding:"required,email"`
	Phone           *string `json:"phone"`
	Subject         *string `json:"subject" binding:"omitempty,max=500"`
	Message         string  `json:"message" binding:"required,min=10,max=5000"`
	Source          *string `json:"source"`
	// Honeypot field for spam detection
	Website         string  `json:"website"` // Should be empty for real users
}

// ToContact converts ContactSubmission to full Contact model
func (cs *ContactSubmission) ToContact() *Contact {
	// Split name into first and last name
	firstName := cs.Name
	var lastName *string
	
	// Simple name splitting logic
	if len(cs.Name) > 0 {
		parts := strings.Fields(cs.Name)
		if len(parts) > 1 {
			firstName = parts[0]
			lastNameStr := strings.Join(parts[1:], " ")
			lastName = &lastNameStr
		}
	}
	
	// Convert status
	status := StatusNew
	switch cs.Status {
	case "contacted", "in_progress":
		status = StatusContacted
	case "resolved", "closed":
		status = StatusClosedWon
	case "spam":
		status = StatusClosedLost
	}
	
	contact := &Contact{
		FirstName:             firstName,
		LastName:              lastName,
		Email:                 cs.Email,
		Phone:                 cs.Phone,
		Subject:               cs.Subject,
		Message:               &cs.Message,
		Status:                status,
		Priority:              PriorityMedium,
		ContactTypeID:         1, // Default to General Inquiry
		ContactSourceID:       1, // Default to Website
		DataSource:            "form",
		Country:               "India",
		PreferredContactMethod: ContactMethodEmail,
		CreatedAt:             cs.CreatedAt,
		UpdatedAt:             cs.UpdatedAt,
	}
	
	if cs.AssignedTo != nil {
		assignedTo := uint(*cs.AssignedTo)
		contact.AssignedTo = &assignedTo
	}
	
	if cs.Source != nil {
		// Map source to UTM source
		contact.UTMSource = cs.Source
	}
	
	return contact
}

// FromContact creates ContactSubmission from full Contact model
func (cs *ContactSubmission) FromContact(c *Contact) {
	cs.Name = c.GetFullName()
	cs.Email = c.Email
	cs.Phone = c.Phone
	cs.Subject = c.Subject
	if c.Message != nil {
		cs.Message = *c.Message
	}
	
	// Convert status back to simple format
	switch c.Status {
	case StatusNew:
		cs.Status = "new"
	case StatusContacted, StatusQualified:
		cs.Status = "in_progress"
	case StatusClosedWon:
		cs.Status = "resolved"
	case StatusClosedLost:
		cs.Status = "spam"
	default:
		cs.Status = "new"
	}
	
	if c.AssignedTo != nil {
		assignedTo := int(*c.AssignedTo)
		cs.AssignedTo = &assignedTo
	}
	
	if c.UTMSource != nil {
		cs.Source = c.UTMSource
	}
	
	cs.CreatedAt = c.CreatedAt
	cs.UpdatedAt = c.UpdatedAt
}
