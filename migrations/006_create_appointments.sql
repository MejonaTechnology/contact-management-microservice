-- Create appointments table for scheduling management
CREATE TABLE IF NOT EXISTS appointments (
    id INT PRIMARY KEY AUTO_INCREMENT,
    contact_id INT NOT NULL,
    
    -- Appointment Basic Information
    title VARCHAR(255) NOT NULL,
    description TEXT,
    appointment_type ENUM('consultation', 'demo', 'meeting', 'call', 'presentation', 'follow_up', 'other') DEFAULT 'consultation',
    
    -- Scheduling Information
    scheduled_date DATE NOT NULL,
    scheduled_time TIME NOT NULL,
    timezone VARCHAR(50) DEFAULT 'Asia/Kolkata',
    duration_minutes INT DEFAULT 60,
    
    -- Status and Management
    status ENUM('requested', 'confirmed', 'rescheduled', 'completed', 'cancelled', 'no_show') DEFAULT 'requested',
    priority ENUM('low', 'medium', 'high', 'urgent') DEFAULT 'medium',
    
    -- Location and Meeting Details
    meeting_type ENUM('in_person', 'video_call', 'phone_call', 'hybrid') DEFAULT 'video_call',
    location VARCHAR(500), -- Physical address or meeting room
    meeting_link VARCHAR(500), -- Zoom, Teams, Meet link
    meeting_id VARCHAR(100), -- Meeting ID for video calls
    meeting_password VARCHAR(100), -- Meeting password if required
    phone_number VARCHAR(20), -- Phone number for calls
    
    -- Assignment and Participants
    assigned_to INT NOT NULL, -- Primary person handling the appointment
    participants JSON, -- Additional participants with roles
    
    -- Confirmation and Communication
    confirmation_sent BOOLEAN DEFAULT FALSE,
    confirmation_sent_at TIMESTAMP NULL,
    reminder_sent BOOLEAN DEFAULT FALSE,
    reminder_sent_at TIMESTAMP NULL,
    
    -- Rescheduling Information
    original_scheduled_date DATE,
    original_scheduled_time TIME,
    reschedule_count INT DEFAULT 0,
    reschedule_reason TEXT,
    
    -- Completion and Follow-up
    completed_at TIMESTAMP NULL,
    completion_notes TEXT,
    outcome ENUM('successful', 'needs_follow_up', 'not_interested', 'reschedule_needed', 'no_show') NULL,
    next_action TEXT,
    next_appointment_suggested BOOLEAN DEFAULT FALSE,
    
    -- Business Information
    estimated_value DECIMAL(12,2) DEFAULT 0.00,
    actual_value DECIMAL(12,2) DEFAULT 0.00,
    conversion_probability INT DEFAULT 0, -- 0-100
    
    -- Preparation and Requirements
    preparation_notes TEXT,
    client_requirements TEXT,
    materials_needed JSON, -- Presentations, documents, etc.
    agenda JSON, -- Meeting agenda items
    
    -- Integration Data
    calendar_event_id VARCHAR(255), -- Google Calendar, Outlook event ID
    external_meeting_id VARCHAR(255), -- External system reference
    booking_source VARCHAR(100) DEFAULT 'manual', -- manual, api, website, admin
    
    -- Metadata and Custom Fields
    tags JSON,
    custom_fields JSON,
    
    -- Audit and Tracking
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by INT,
    updated_by INT,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes for Performance
    INDEX idx_contact_id (contact_id),
    INDEX idx_scheduled_date (scheduled_date),
    INDEX idx_scheduled_time (scheduled_time),
    INDEX idx_status (status),
    INDEX idx_assigned_to (assigned_to),
    INDEX idx_appointment_type (appointment_type),
    INDEX idx_meeting_type (meeting_type),
    INDEX idx_confirmation_sent (confirmation_sent),
    INDEX idx_reminder_sent (reminder_sent),
    INDEX idx_created_at (created_at),
    INDEX idx_deleted_at (deleted_at),
    
    -- Composite Indexes
    INDEX idx_status_date (status, scheduled_date),
    INDEX idx_assigned_date (assigned_to, scheduled_date),
    INDEX idx_contact_status (contact_id, status),
    INDEX idx_date_time (scheduled_date, scheduled_time),
    
    -- Foreign Key Constraints
    FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE
);

-- Create appointment_attendees table for tracking who attended
CREATE TABLE IF NOT EXISTS appointment_attendees (
    id INT PRIMARY KEY AUTO_INCREMENT,
    appointment_id INT NOT NULL,
    
    -- Attendee Information
    attendee_type ENUM('contact', 'employee', 'external') NOT NULL,
    attendee_id INT, -- Reference to contact or employee ID
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(20),
    role VARCHAR(100), -- Their role in the meeting
    
    -- Attendance Tracking
    invitation_sent BOOLEAN DEFAULT FALSE,
    invitation_sent_at TIMESTAMP NULL,
    response_status ENUM('pending', 'accepted', 'declined', 'tentative') DEFAULT 'pending',
    response_date TIMESTAMP NULL,
    attended BOOLEAN DEFAULT FALSE,
    joined_at TIMESTAMP NULL,
    left_at TIMESTAMP NULL,
    
    -- Additional Information
    notes TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_appointment_id (appointment_id),
    INDEX idx_attendee_type (attendee_type),
    INDEX idx_attendee_id (attendee_id),
    INDEX idx_email (email),
    INDEX idx_response_status (response_status),
    
    FOREIGN KEY (appointment_id) REFERENCES appointments(id) ON DELETE CASCADE
);

-- Create appointment_reminders table for automated reminder tracking
CREATE TABLE IF NOT EXISTS appointment_reminders (
    id INT PRIMARY KEY AUTO_INCREMENT,
    appointment_id INT NOT NULL,
    
    -- Reminder Configuration
    reminder_type ENUM('email', 'sms', 'push', 'call') NOT NULL,
    reminder_time_minutes INT NOT NULL, -- Minutes before appointment
    
    -- Status and Execution
    status ENUM('scheduled', 'sent', 'failed', 'cancelled') DEFAULT 'scheduled',
    scheduled_send_time TIMESTAMP NOT NULL,
    actual_send_time TIMESTAMP NULL,
    
    -- Content and Recipient
    recipient_email VARCHAR(255),
    recipient_phone VARCHAR(20),
    subject VARCHAR(255),
    message TEXT,
    
    -- Delivery Tracking
    delivery_status VARCHAR(50),
    delivery_error TEXT,
    opened BOOLEAN DEFAULT FALSE,
    clicked BOOLEAN DEFAULT FALSE,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_appointment_id (appointment_id),
    INDEX idx_reminder_type (reminder_type),
    INDEX idx_status (status),
    INDEX idx_scheduled_send_time (scheduled_send_time),
    
    FOREIGN KEY (appointment_id) REFERENCES appointments(id) ON DELETE CASCADE
);