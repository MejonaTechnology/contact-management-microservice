-- Create main contacts table with comprehensive contact management
CREATE TABLE IF NOT EXISTS contacts (
    id INT PRIMARY KEY AUTO_INCREMENT,
    
    -- Basic Information
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100),
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    company VARCHAR(200),
    job_title VARCHAR(100),
    website VARCHAR(255),
    
    -- Address Information
    address_line1 VARCHAR(255),
    address_line2 VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100) DEFAULT 'India',
    
    -- Contact Details
    contact_type_id INT NOT NULL,
    contact_source_id INT NOT NULL,
    subject VARCHAR(500),
    message TEXT,
    preferred_contact_method ENUM('email', 'phone', 'sms', 'whatsapp') DEFAULT 'email',
    
    -- Lead Management
    status ENUM('new', 'contacted', 'qualified', 'proposal', 'negotiation', 'closed_won', 'closed_lost', 'on_hold', 'nurturing') DEFAULT 'new',
    priority ENUM('low', 'medium', 'high', 'urgent') DEFAULT 'medium',
    lead_score INT DEFAULT 0, -- 0-100 scoring system
    estimated_value DECIMAL(12,2) DEFAULT 0.00,
    probability INT DEFAULT 0, -- 0-100 probability of conversion
    
    -- Assignment and Ownership
    assigned_to INT, -- Foreign key to employees/users table
    assigned_at TIMESTAMP NULL,
    assigned_by INT, -- Who assigned this contact
    
    -- Communication Tracking
    last_contact_date TIMESTAMP NULL,
    next_followup_date TIMESTAMP NULL,
    response_time_hours INT DEFAULT 0, -- Time to first response
    total_interactions INT DEFAULT 0,
    email_opened BOOLEAN DEFAULT FALSE,
    email_clicked BOOLEAN DEFAULT FALSE,
    
    -- Lifecycle Tracking
    first_contact_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_activity_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    conversion_date TIMESTAMP NULL,
    closed_date TIMESTAMP NULL,
    
    -- Technical Fields
    ip_address VARCHAR(45),
    user_agent TEXT,
    referrer_url VARCHAR(500),
    landing_page VARCHAR(500),
    utm_source VARCHAR(100),
    utm_medium VARCHAR(100),
    utm_campaign VARCHAR(100),
    utm_term VARCHAR(100),
    utm_content VARCHAR(100),
    
    -- Data Management
    is_verified BOOLEAN DEFAULT FALSE,
    is_duplicate BOOLEAN DEFAULT FALSE,
    original_contact_id INT NULL, -- Reference to original if duplicate
    data_source VARCHAR(100) DEFAULT 'form', -- form, api, import, manual
    
    -- Privacy and Compliance
    marketing_consent BOOLEAN DEFAULT FALSE,
    data_processing_consent BOOLEAN DEFAULT TRUE,
    gdpr_consent BOOLEAN DEFAULT FALSE,
    unsubscribed BOOLEAN DEFAULT FALSE,
    do_not_call BOOLEAN DEFAULT FALSE,
    
    -- Metadata
    tags JSON, -- Flexible tagging system
    custom_fields JSON, -- Custom data fields
    notes TEXT,
    
    -- Audit Fields
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by INT,
    updated_by INT,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes for Performance
    INDEX idx_email (email),
    INDEX idx_phone (phone),
    INDEX idx_status (status),
    INDEX idx_priority (priority),
    INDEX idx_assigned_to (assigned_to),
    INDEX idx_contact_type (contact_type_id),
    INDEX idx_contact_source (contact_source_id),
    INDEX idx_lead_score (lead_score),
    INDEX idx_next_followup (next_followup_date),
    INDEX idx_first_contact_date (first_contact_date),
    INDEX idx_last_activity_date (last_activity_date),
    INDEX idx_created_at (created_at),
    INDEX idx_deleted_at (deleted_at),
    INDEX idx_utm_source (utm_source),
    INDEX idx_is_duplicate (is_duplicate),
    
    -- Composite Indexes
    INDEX idx_status_assigned (status, assigned_to),
    INDEX idx_type_source (contact_type_id, contact_source_id),
    INDEX idx_name_search (first_name, last_name),
    
    -- Full-text Search Index
    FULLTEXT idx_search_text (first_name, last_name, email, company, subject, message),
    
    -- Foreign Key Constraints
    FOREIGN KEY (contact_type_id) REFERENCES contact_types(id) ON DELETE RESTRICT,
    FOREIGN KEY (contact_source_id) REFERENCES contact_sources(id) ON DELETE RESTRICT,
    FOREIGN KEY (original_contact_id) REFERENCES contacts(id) ON DELETE SET NULL
);