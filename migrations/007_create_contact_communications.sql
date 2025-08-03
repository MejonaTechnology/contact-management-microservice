-- Create contact_communications table for tracking all communications
CREATE TABLE IF NOT EXISTS contact_communications (
    id INT PRIMARY KEY AUTO_INCREMENT,
    contact_id INT NOT NULL,
    
    -- Communication Basic Information
    communication_type ENUM('email', 'sms', 'phone_call', 'whatsapp', 'chat', 'letter', 'fax', 'video_call', 'in_person') NOT NULL,
    direction ENUM('inbound', 'outbound') NOT NULL,
    subject VARCHAR(500),
    content TEXT,
    
    -- Status and Tracking
    status ENUM('draft', 'sending', 'sent', 'delivered', 'read', 'replied', 'failed', 'bounced') DEFAULT 'draft',
    priority ENUM('low', 'medium', 'high', 'urgent') DEFAULT 'medium',
    
    -- Sender and Recipient Information  
    from_email VARCHAR(255),
    to_email VARCHAR(255),
    cc_emails JSON, -- Array of CC email addresses
    bcc_emails JSON, -- Array of BCC email addresses
    from_phone VARCHAR(20),
    to_phone VARCHAR(20),
    
    -- Email Specific Fields
    email_message_id VARCHAR(255), -- Email Message-ID header
    email_thread_id VARCHAR(255), -- For email threading
    email_template_id INT, -- Reference to email template used
    html_content TEXT, -- HTML version of email
    plain_content TEXT, -- Plain text version
    
    -- SMS Specific Fields
    sms_provider VARCHAR(50), -- twilio, aws_sns, etc.
    sms_provider_id VARCHAR(100), -- Provider's message ID
    sms_segments INT DEFAULT 1, -- Number of SMS segments
    
    -- Phone Call Specific Fields
    call_duration_seconds INT DEFAULT 0,
    call_recording_url VARCHAR(500),
    call_provider VARCHAR(50), -- Provider used for the call
    call_provider_id VARCHAR(100),
    
    -- Delivery and Engagement Tracking
    sent_at TIMESTAMP NULL,
    delivered_at TIMESTAMP NULL,
    opened_at TIMESTAMP NULL,
    first_opened_at TIMESTAMP NULL,
    last_opened_at TIMESTAMP NULL,
    open_count INT DEFAULT 0,
    clicked_at TIMESTAMP NULL,
    first_clicked_at TIMESTAMP NULL,
    last_clicked_at TIMESTAMP NULL,
    click_count INT DEFAULT 0,
    replied_at TIMESTAMP NULL,
    
    -- Response and Follow-up
    response_required BOOLEAN DEFAULT FALSE,
    response_due_date TIMESTAMP NULL,
    follow_up_required BOOLEAN DEFAULT FALSE,
    follow_up_date TIMESTAMP NULL,
    follow_up_completed BOOLEAN DEFAULT FALSE,
    
    -- Related Information
    parent_communication_id INT, -- For replies and follow-ups
    campaign_id INT, -- Marketing campaign reference
    template_id INT, -- Template used
    automation_id INT, -- Automation workflow reference
    
    -- File Attachments
    attachments JSON, -- Array of attachment information
    attachment_count INT DEFAULT 0,
    
    -- Cost and Analytics
    cost DECIMAL(8,4) DEFAULT 0.0000, -- Cost of sending (for SMS, calls)
    revenue_attributed DECIMAL(12,2) DEFAULT 0.00, -- Revenue attributed to this communication
    
    -- Integration and External References
    external_id VARCHAR(255), -- External system reference
    external_system VARCHAR(100), -- Which external system
    sync_status ENUM('pending', 'synced', 'failed') DEFAULT 'synced',
    
    -- User and Assignment
    sent_by INT, -- User who sent the communication
    assigned_to INT, -- User responsible for follow-up
    
    -- Metadata and Tags
    tags JSON,
    metadata JSON, -- Additional structured data
    
    -- Audit Fields
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by INT,
    updated_by INT,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes for Performance
    INDEX idx_contact_id (contact_id),
    INDEX idx_communication_type (communication_type),
    INDEX idx_direction (direction),
    INDEX idx_status (status),
    INDEX idx_sent_at (sent_at),
    INDEX idx_delivered_at (delivered_at),
    INDEX idx_opened_at (opened_at),
    INDEX idx_replied_at (replied_at),
    INDEX idx_sent_by (sent_by),
    INDEX idx_assigned_to (assigned_to),
    INDEX idx_email_message_id (email_message_id),
    INDEX idx_email_thread_id (email_thread_id),
    INDEX idx_parent_communication_id (parent_communication_id),
    INDEX idx_created_at (created_at),
    INDEX idx_deleted_at (deleted_at),
    
    -- Composite Indexes
    INDEX idx_contact_type (contact_id, communication_type),
    INDEX idx_contact_status (contact_id, status),
    INDEX idx_type_direction (communication_type, direction),
    INDEX idx_status_sent (status, sent_at),
    
    -- Full-text Search
    FULLTEXT idx_content_search (subject, content, plain_content),
    
    -- Foreign Key Constraints
    FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_communication_id) REFERENCES contact_communications(id) ON DELETE SET NULL
);

-- Create communication_recipients table for detailed recipient tracking
CREATE TABLE IF NOT EXISTS communication_recipients (
    id INT PRIMARY KEY AUTO_INCREMENT,
    communication_id INT NOT NULL,
    
    -- Recipient Information
    recipient_type ENUM('to', 'cc', 'bcc') NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(20),
    name VARCHAR(255),
    
    -- Delivery Status per Recipient
    status ENUM('pending', 'sent', 'delivered', 'failed', 'bounced', 'complained') DEFAULT 'pending',
    delivered_at TIMESTAMP NULL,
    opened_at TIMESTAMP NULL,
    clicked_at TIMESTAMP NULL,
    
    -- Tracking Information
    open_count INT DEFAULT 0,
    click_count INT DEFAULT 0,
    bounce_reason TEXT,
    failure_reason TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_communication_id (communication_id),
    INDEX idx_recipient_type (recipient_type),
    INDEX idx_email (email),
    INDEX idx_phone (phone),
    INDEX idx_status (status),
    
    FOREIGN KEY (communication_id) REFERENCES contact_communications(id) ON DELETE CASCADE
);

-- Create communication_templates table for reusable templates
CREATE TABLE IF NOT EXISTS communication_templates (
    id INT PRIMARY KEY AUTO_INCREMENT,
    
    -- Template Information
    name VARCHAR(255) NOT NULL,
    description TEXT,
    communication_type ENUM('email', 'sms', 'whatsapp') NOT NULL,
    category VARCHAR(100), -- welcome, follow_up, proposal, etc.
    
    -- Template Content
    subject VARCHAR(500), -- For email templates
    html_content TEXT, -- HTML version
    plain_content TEXT, -- Plain text version
    variables JSON, -- Available template variables
    
    -- Configuration
    is_active BOOLEAN DEFAULT TRUE,
    is_system BOOLEAN DEFAULT FALSE, -- System vs user templates
    usage_count INT DEFAULT 0,
    
    -- Personalization
    personalization_rules JSON, -- Rules for personalizing content
    
    -- Metadata
    tags JSON,
    
    -- Audit Fields
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by INT,
    updated_by INT,
    
    INDEX idx_name (name),
    INDEX idx_communication_type (communication_type),
    INDEX idx_category (category),
    INDEX idx_is_active (is_active),
    INDEX idx_usage_count (usage_count)
);