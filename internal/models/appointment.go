package models

import (
	"time"
)

// AppointmentType represents different types of appointments
type AppointmentType string

const (
	AppointmentConsultation AppointmentType = "consultation"
	AppointmentDemo         AppointmentType = "demo"
	AppointmentMeeting      AppointmentType = "meeting"
	AppointmentCall         AppointmentType = "call"
	AppointmentPresentation AppointmentType = "presentation"
	AppointmentFollowUp     AppointmentType = "follow_up"
	AppointmentOther        AppointmentType = "other"
)

// AppointmentStatus represents the status of an appointment
type AppointmentStatus string

const (
	AppointmentRequested   AppointmentStatus = "requested"
	AppointmentConfirmed   AppointmentStatus = "confirmed"
	AppointmentRescheduled AppointmentStatus = "rescheduled"
	AppointmentCompleted   AppointmentStatus = "completed"
	AppointmentCancelled   AppointmentStatus = "cancelled"
	AppointmentNoShow      AppointmentStatus = "no_show"
)

// MeetingType represents how the meeting will be conducted
type MeetingType string

const (
	MeetingInPerson  MeetingType = "in_person"
	MeetingVideoCall MeetingType = "video_call"
	MeetingPhoneCall MeetingType = "phone_call"
	MeetingHybrid    MeetingType = "hybrid"
)

// AppointmentOutcome represents the outcome of a completed appointment
type AppointmentOutcome string

const (
	OutcomeSuccessful         AppointmentOutcome = "successful"
	OutcomeNeedsFollowUp     AppointmentOutcome = "needs_follow_up"
	OutcomeNotInterested     AppointmentOutcome = "not_interested"
	OutcomeRescheduleNeeded  AppointmentOutcome = "reschedule_needed"
	OutcomeNoShow            AppointmentOutcome = "no_show"
)

// Appointment represents scheduled meetings and appointments with contacts
type Appointment struct {
	ID                        uint               `json:"id" gorm:"primaryKey"`
	ContactID                 uint               `json:"contact_id" gorm:"column:contact_id;not null;index"`
	
	// Appointment Basic Information
	Title                     string             `json:"title" gorm:"column:title;size:255;not null" binding:"required,min=2,max=255"`
	Description               *string            `json:"description" gorm:"column:description;type:text"`
	AppointmentType           AppointmentType    `json:"appointment_type" gorm:"column:appointment_type;default:consultation"`
	
	// Scheduling Information
	ScheduledDate             time.Time          `json:"scheduled_date" gorm:"column:scheduled_date;not null;index"`
	ScheduledTime             string             `json:"scheduled_time" gorm:"column:scheduled_time;not null"` // TIME format
	Timezone                  string             `json:"timezone" gorm:"column:timezone;size:50;default:Asia/Kolkata"`
	DurationMinutes           int                `json:"duration_minutes" gorm:"column:duration_minutes;default:60"`
	
	// Status and Management
	Status                    AppointmentStatus  `json:"status" gorm:"column:status;default:requested;index"`
	Priority                  ContactPriority    `json:"priority" gorm:"column:priority;default:medium"`
	
	// Location and Meeting Details
	MeetingType               MeetingType        `json:"meeting_type" gorm:"column:meeting_type;default:video_call"`
	Location                  *string            `json:"location" gorm:"column:location;size:500"`
	MeetingLink               *string            `json:"meeting_link" gorm:"column:meeting_link;size:500"`
	MeetingID                 *string            `json:"meeting_id" gorm:"column:meeting_id;size:100"`
	MeetingPassword           *string            `json:"meeting_password" gorm:"column:meeting_password;size:100"`
	PhoneNumber               *string            `json:"phone_number" gorm:"column:phone_number;size:20"`
	
	// Assignment and Participants
	AssignedTo                uint               `json:"assigned_to" gorm:"column:assigned_to;not null;index"`
	Participants              JSONMap            `json:"participants" gorm:"column:participants;type:json"`
	
	// Confirmation and Communication
	ConfirmationSent          bool               `json:"confirmation_sent" gorm:"column:confirmation_sent;default:false"`
	ConfirmationSentAt        *time.Time         `json:"confirmation_sent_at" gorm:"column:confirmation_sent_at"`
	ReminderSent              bool               `json:"reminder_sent" gorm:"column:reminder_sent;default:false"`
	ReminderSentAt            *time.Time         `json:"reminder_sent_at" gorm:"column:reminder_sent_at"`
	
	// Rescheduling Information
	OriginalScheduledDate     *time.Time         `json:"original_scheduled_date" gorm:"column:original_scheduled_date"`
	OriginalScheduledTime     *string            `json:"original_scheduled_time" gorm:"column:original_scheduled_time;size:8"`
	RescheduleCount           int                `json:"reschedule_count" gorm:"column:reschedule_count;default:0"`
	RescheduleReason          *string            `json:"reschedule_reason" gorm:"column:reschedule_reason;type:text"`
	
	// Completion and Follow-up
	CompletedAt               *time.Time         `json:"completed_at" gorm:"column:completed_at"`
	CompletionNotes           *string            `json:"completion_notes" gorm:"column:completion_notes;type:text"`
	Outcome                   *AppointmentOutcome `json:"outcome" gorm:"column:outcome"`
	NextAction                *string            `json:"next_action" gorm:"column:next_action;type:text"`
	NextAppointmentSuggested  bool               `json:"next_appointment_suggested" gorm:"column:next_appointment_suggested;default:false"`
	
	// Business Information
	EstimatedValue            float64            `json:"estimated_value" gorm:"column:estimated_value;type:decimal(12,2);default:0.00"`
	ActualValue               float64            `json:"actual_value" gorm:"column:actual_value;type:decimal(12,2);default:0.00"`
	ConversionProbability     int                `json:"conversion_probability" gorm:"column:conversion_probability;default:0"`
	
	// Preparation and Requirements
	PreparationNotes          *string            `json:"preparation_notes" gorm:"column:preparation_notes;type:text"`
	ClientRequirements        *string            `json:"client_requirements" gorm:"column:client_requirements;type:text"`
	MaterialsNeeded           JSONMap            `json:"materials_needed" gorm:"column:materials_needed;type:json"`
	Agenda                    JSONMap            `json:"agenda" gorm:"column:agenda;type:json"`
	
	// Integration Data
	CalendarEventID           *string            `json:"calendar_event_id" gorm:"column:calendar_event_id;size:255"`
	ExternalMeetingID         *string            `json:"external_meeting_id" gorm:"column:external_meeting_id;size:255"`
	BookingSource             string             `json:"booking_source" gorm:"column:booking_source;size:100;default:manual"`
	
	// Metadata and Custom Fields
	Tags                      JSONMap            `json:"tags" gorm:"column:tags;type:json"`
	CustomFields              JSONMap            `json:"custom_fields" gorm:"column:custom_fields;type:json"`
	
	// Audit and Tracking
	CreatedAt                 time.Time          `json:"created_at" gorm:"column:created_at;index"`
	UpdatedAt                 time.Time          `json:"updated_at" gorm:"column:updated_at"`
	CreatedBy                 *uint              `json:"created_by" gorm:"column:created_by"`
	UpdatedBy                 *uint              `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt                 *time.Time         `json:"deleted_at" gorm:"column:deleted_at;index"`
	
	// Relationships
	Contact                   *Contact           `json:"contact,omitempty" gorm:"foreignKey:ContactID"`
	Attendees                 []AppointmentAttendee `json:"attendees,omitempty" gorm:"foreignKey:AppointmentID"`
	Reminders                 []AppointmentReminder `json:"reminders,omitempty" gorm:"foreignKey:AppointmentID"`
}

// TableName specifies the table name for Appointment
func (Appointment) TableName() string {
	return "appointments"
}

// GetScheduledDateTime combines scheduled date and time into a single datetime
func (a *Appointment) GetScheduledDateTime() (time.Time, error) {
	dateStr := a.ScheduledDate.Format("2006-01-02")
	datetimeStr := dateStr + " " + a.ScheduledTime
	return time.Parse("2006-01-02 15:04:05", datetimeStr)
}

// IsToday checks if the appointment is scheduled for today
func (a *Appointment) IsToday() bool {
	today := time.Now().Format("2006-01-02")
	scheduledDay := a.ScheduledDate.Format("2006-01-02")
	return today == scheduledDay
}

// IsUpcoming checks if the appointment is scheduled for the future
func (a *Appointment) IsUpcoming() bool {
	scheduledDateTime, err := a.GetScheduledDateTime()
	if err != nil {
		return false
	}
	return scheduledDateTime.After(time.Now())
}

// IsOverdue checks if the appointment is overdue
func (a *Appointment) IsOverdue() bool {
	if a.Status == AppointmentCompleted || a.Status == AppointmentCancelled {
		return false
	}
	scheduledDateTime, err := a.GetScheduledDateTime()
	if err != nil {
		return false
	}
	return scheduledDateTime.Before(time.Now())
}

// AppointmentAttendee represents people attending the appointment
type AppointmentAttendee struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	AppointmentID    uint      `json:"appointment_id" gorm:"column:appointment_id;not null;index"`
	
	// Attendee Information
	AttendeeType     string    `json:"attendee_type" gorm:"column:attendee_type;not null"` // contact, employee, external
	AttendeeID       *uint     `json:"attendee_id" gorm:"column:attendee_id;index"`
	Name             string    `json:"name" gorm:"column:name;size:255;not null"`
	Email            *string   `json:"email" gorm:"column:email;size:255;index"`
	Phone            *string   `json:"phone" gorm:"column:phone;size:20"`
	Role             *string   `json:"role" gorm:"column:role;size:100"`
	
	// Attendance Tracking
	InvitationSent   bool      `json:"invitation_sent" gorm:"column:invitation_sent;default:false"`
	InvitationSentAt *time.Time `json:"invitation_sent_at" gorm:"column:invitation_sent_at"`
	ResponseStatus   string    `json:"response_status" gorm:"column:response_status;default:pending"` // pending, accepted, declined, tentative
	ResponseDate     *time.Time `json:"response_date" gorm:"column:response_date"`
	Attended         bool      `json:"attended" gorm:"column:attended;default:false"`
	JoinedAt         *time.Time `json:"joined_at" gorm:"column:joined_at"`
	LeftAt           *time.Time `json:"left_at" gorm:"column:left_at"`
	
	// Additional Information
	Notes            *string   `json:"notes" gorm:"column:notes;type:text"`
	
	CreatedAt        time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"column:updated_at"`
	
	// Relationships
	Appointment      *Appointment `json:"appointment,omitempty" gorm:"foreignKey:AppointmentID"`
}

// TableName specifies the table name for AppointmentAttendee
func (AppointmentAttendee) TableName() string {
	return "appointment_attendees"
}

// AppointmentReminder represents automated reminder tracking
type AppointmentReminder struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	AppointmentID      uint      `json:"appointment_id" gorm:"column:appointment_id;not null;index"`
	
	// Reminder Configuration
	ReminderType       string    `json:"reminder_type" gorm:"column:reminder_type;not null"` // email, sms, push, call
	ReminderTimeMinutes int      `json:"reminder_time_minutes" gorm:"column:reminder_time_minutes;not null"`
	
	// Status and Execution
	Status             string    `json:"status" gorm:"column:status;default:scheduled"` // scheduled, sent, failed, cancelled
	ScheduledSendTime  time.Time `json:"scheduled_send_time" gorm:"column:scheduled_send_time;not null;index"`
	ActualSendTime     *time.Time `json:"actual_send_time" gorm:"column:actual_send_time"`
	
	// Content and Recipient
	RecipientEmail     *string   `json:"recipient_email" gorm:"column:recipient_email;size:255"`
	RecipientPhone     *string   `json:"recipient_phone" gorm:"column:recipient_phone;size:20"`
	Subject            *string   `json:"subject" gorm:"column:subject;size:255"`
	Message            *string   `json:"message" gorm:"column:message;type:text"`
	
	// Delivery Tracking
	DeliveryStatus     *string   `json:"delivery_status" gorm:"column:delivery_status;size:50"`
	DeliveryError      *string   `json:"delivery_error" gorm:"column:delivery_error;type:text"`
	Opened             bool      `json:"opened" gorm:"column:opened;default:false"`
	Clicked            bool      `json:"clicked" gorm:"column:clicked;default:false"`
	
	CreatedAt          time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"column:updated_at"`
	
	// Relationships
	Appointment        *Appointment `json:"appointment,omitempty" gorm:"foreignKey:AppointmentID"`
}

// TableName specifies the table name for AppointmentReminder
func (AppointmentReminder) TableName() string {
	return "appointment_reminders"
}

// Request and Response structures

// AppointmentRequest represents the request structure for creating/updating appointments
type AppointmentRequest struct {
	ContactID                uint                `json:"contact_id" binding:"required,min=1"`
	Title                    string              `json:"title" binding:"required,min=2,max=255"`
	Description              *string             `json:"description" binding:"omitempty,max=5000"`
	AppointmentType          *AppointmentType    `json:"appointment_type"`
	ScheduledDate            string              `json:"scheduled_date" binding:"required"` // YYYY-MM-DD format
	ScheduledTime            string              `json:"scheduled_time" binding:"required"` // HH:MM:SS format
	Timezone                 *string             `json:"timezone"`
	DurationMinutes          *int                `json:"duration_minutes" binding:"omitempty,min=15,max=480"`
	Priority                 *ContactPriority    `json:"priority"`
	MeetingType              *MeetingType        `json:"meeting_type"`
	Location                 *string             `json:"location" binding:"omitempty,max=500"`
	MeetingLink              *string             `json:"meeting_link" binding:"omitempty,url"`
	MeetingID                *string             `json:"meeting_id" binding:"omitempty,max=100"`
	MeetingPassword          *string             `json:"meeting_password" binding:"omitempty,max=100"`
	PhoneNumber              *string             `json:"phone_number"`
	AssignedTo               uint                `json:"assigned_to" binding:"required,min=1"`
	EstimatedValue           *float64            `json:"estimated_value" binding:"omitempty,min=0"`
	ConversionProbability    *int                `json:"conversion_probability" binding:"omitempty,min=0,max=100"`
	PreparationNotes         *string             `json:"preparation_notes" binding:"omitempty,max=5000"`
	ClientRequirements       *string             `json:"client_requirements" binding:"omitempty,max=5000"`
	MaterialsNeeded          JSONMap             `json:"materials_needed"`
	Agenda                   JSONMap             `json:"agenda"`
	Tags                     JSONMap             `json:"tags"`
	CustomFields             JSONMap             `json:"custom_fields"`
}

// PublicAppointmentRequest represents a simplified request for public appointment booking
type PublicAppointmentRequest struct {
	FirstName        string              `json:"first_name" binding:"required,min=2,max=100"`
	LastName         *string             `json:"last_name" binding:"omitempty,max=100"`
	Email            string              `json:"email" binding:"required,email"`
	Phone            *string             `json:"phone"`
	Company          *string             `json:"company" binding:"omitempty,max=200"`
	Title            string              `json:"title" binding:"required,min=2,max=255"`
	Description      *string             `json:"description" binding:"omitempty,max=1000"`
	AppointmentType  *AppointmentType    `json:"appointment_type"`
	PreferredDate    string              `json:"preferred_date" binding:"required"`
	PreferredTime    string              `json:"preferred_time" binding:"required"`
	MeetingType      *MeetingType        `json:"meeting_type"`
	// Honeypot field for spam detection
	Website          string              `json:"website"` // Should be empty for real users
}

// AppointmentResponse represents the response structure for appointments
type AppointmentResponse struct {
	ID                        uint                `json:"id"`
	ContactID                 uint                `json:"contact_id"`
	Title                     string              `json:"title"`
	Description               *string             `json:"description"`
	AppointmentType           AppointmentType     `json:"appointment_type"`
	ScheduledDate             time.Time           `json:"scheduled_date"`
	ScheduledTime             string              `json:"scheduled_time"`
	ScheduledDateTime         time.Time           `json:"scheduled_datetime"`
	Timezone                  string              `json:"timezone"`
	DurationMinutes           int                 `json:"duration_minutes"`
	Status                    AppointmentStatus   `json:"status"`
	Priority                  ContactPriority     `json:"priority"`
	MeetingType               MeetingType         `json:"meeting_type"`
	Location                  *string             `json:"location"`
	MeetingLink               *string             `json:"meeting_link"`
	MeetingID                 *string             `json:"meeting_id"`
	PhoneNumber               *string             `json:"phone_number"`
	AssignedTo                uint                `json:"assigned_to"`
	ConfirmationSent          bool                `json:"confirmation_sent"`
	ReminderSent              bool                `json:"reminder_sent"`
	RescheduleCount           int                 `json:"reschedule_count"`
	CompletedAt               *time.Time          `json:"completed_at"`
	Outcome                   *AppointmentOutcome `json:"outcome"`
	EstimatedValue            float64             `json:"estimated_value"`
	ActualValue               float64             `json:"actual_value"`
	ConversionProbability     int                 `json:"conversion_probability"`
	Tags                      JSONMap             `json:"tags"`
	CustomFields              JSONMap             `json:"custom_fields"`
	CreatedAt                 time.Time           `json:"created_at"`
	UpdatedAt                 time.Time           `json:"updated_at"`
	// Computed fields
	IsToday                   bool                `json:"is_today"`
	IsUpcoming                bool                `json:"is_upcoming"`
	IsOverdue                 bool                `json:"is_overdue"`
	// Related data
	Contact                   *ContactResponse    `json:"contact,omitempty"`
	Attendees                 []AppointmentAttendee `json:"attendees,omitempty"`
}

// AppointmentUpdateRequest represents a request to update appointment status
type AppointmentUpdateRequest struct {
	Status      string  `json:"status" binding:"required"`
	Notes       *string `json:"notes,omitempty"`
	CancelReason *string `json:"cancel_reason,omitempty"`
}

// RescheduleRequest represents a request to reschedule an appointment
type RescheduleRequest struct {
	NewStartTime time.Time `json:"new_start_time" binding:"required"`
	NewEndTime   time.Time `json:"new_end_time" binding:"required"`
	Reason       *string   `json:"reason,omitempty"`
	Notes        *string   `json:"notes,omitempty"`
}

// AvailabilitySlotRequest represents a request to find available slots
type AvailabilitySlotRequest struct {
	UserID            uint      `json:"user_id" binding:"required"`
	StartDate         time.Time `json:"start_date" binding:"required"`
	EndDate           time.Time `json:"end_date" binding:"required"`
	Duration          int       `json:"duration" binding:"required,min=15"` // in minutes
	BufferTime        *int      `json:"buffer_time,omitempty"` // in minutes
	Timezone          *string   `json:"timezone,omitempty"`
	BusinessHoursOnly *bool     `json:"business_hours_only,omitempty"`
}

// AvailabilitySlot represents an available time slot
type AvailabilitySlot struct {
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Duration    int       `json:"duration"` // in minutes
	IsAvailable bool      `json:"is_available"`
	UserID      uint      `json:"user_id"`
	Date        time.Time `json:"date"`
}

// AvailabilityResponse represents user availability information
type AvailabilityResponse struct {
	UserID          uint               `json:"user_id"`
	Date            time.Time          `json:"date"`
	IsAvailable     bool               `json:"is_available"`
	WorkingHours    *WorkingHours      `json:"working_hours,omitempty"`
	AvailableSlots  []AvailabilitySlot `json:"available_slots"`
	BusySlots       []AvailabilitySlot `json:"busy_slots"`
	TotalAvailable  int                `json:"total_available_slots"`
	Timezone        string             `json:"timezone"`
}

// WorkingHours represents working hours for a user
type WorkingHours struct {
	StartTime string `json:"start_time"` // e.g., "09:00"
	EndTime   string `json:"end_time"`   // e.g., "17:00"
	Timezone  string `json:"timezone"`
}