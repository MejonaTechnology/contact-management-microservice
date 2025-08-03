-- Create contact_activities table for tracking all interactions and activities
CREATE TABLE IF NOT EXISTS contact_activities (
    id INT PRIMARY KEY AUTO_INCREMENT,
    contact_id INT NOT NULL,
    
    -- Activity Information
    activity_type ENUM(
        'status_change', 'assignment', 'note_added', 'email_sent', 'email_received', 
        'call_made', 'call_received', 'sms_sent', 'meeting_scheduled', 'meeting_held',
        'proposal_sent', 'contract_sent', 'payment_received', 'follow_up',
        'document_shared', 'quote_sent', 'demo_scheduled', 'demo_completed',
        'lead_qualified', 'lead_scored', 'converted', 'lost', 'reopened'
    ) NOT NULL,
    
    title VARCHAR(255) NOT NULL,
    description TEXT,
    outcome TEXT, -- Result of the activity
    
    -- Activity Details
    activity_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    duration_minutes INT DEFAULT 0, -- For calls, meetings, etc.
    
    -- Status and Priority
    status ENUM('pending', 'in_progress', 'completed', 'cancelled') DEFAULT 'completed',
    priority ENUM('low', 'medium', 'high', 'urgent') DEFAULT 'medium',
    
    -- Related Information
    related_entity_type VARCHAR(50), -- email, call, meeting, document, etc.
    related_entity_id INT, -- ID of related entity
    external_reference VARCHAR(255), -- External system reference
    
    -- Communication Details
    direction ENUM('inbound', 'outbound', 'internal') DEFAULT 'outbound',
    channel ENUM('email', 'phone', 'sms', 'whatsapp', 'chat', 'in_person', 'video_call', 'other') DEFAULT 'email',
    
    -- Scheduling Information
    scheduled_date TIMESTAMP NULL,
    reminder_date TIMESTAMP NULL,
    completed_date TIMESTAMP NULL,
    
    -- User and Assignment
    performed_by INT NOT NULL, -- User who performed the activity
    assigned_to INT, -- User assigned to handle follow-up
    
    -- Tracking and Analytics
    is_billable BOOLEAN DEFAULT FALSE,
    billable_amount DECIMAL(10,2) DEFAULT 0.00,
    cost DECIMAL(10,2) DEFAULT 0.00,
    
    -- Metadata
    tags JSON,
    metadata JSON, -- Additional structured data
    attachments JSON, -- File attachments information
    
    -- Audit Fields
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by INT,
    updated_by INT,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_contact_id (contact_id),
    INDEX idx_activity_type (activity_type),
    INDEX idx_activity_date (activity_date),
    INDEX idx_performed_by (performed_by),
    INDEX idx_assigned_to (assigned_to),
    INDEX idx_status (status),
    INDEX idx_scheduled_date (scheduled_date),
    INDEX idx_reminder_date (reminder_date),
    INDEX idx_created_at (created_at),
    INDEX idx_deleted_at (deleted_at),
    
    -- Composite Indexes
    INDEX idx_contact_type (contact_id, activity_type),
    INDEX idx_contact_date (contact_id, activity_date),
    INDEX idx_user_date (performed_by, activity_date),
    
    -- Foreign Key Constraints
    FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE
);