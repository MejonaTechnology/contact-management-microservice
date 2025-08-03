-- Create contact_duplicate_groups for duplicate management
CREATE TABLE IF NOT EXISTS contact_duplicate_groups (
    id INT PRIMARY KEY AUTO_INCREMENT,
    
    -- Group Information
    group_hash VARCHAR(64) NOT NULL UNIQUE, -- Hash of duplicate criteria
    master_contact_id INT, -- The primary/master contact
    duplicate_count INT DEFAULT 0,
    
    -- Duplicate Detection Criteria
    match_criteria JSON, -- What criteria were used to find duplicates
    confidence_score DECIMAL(5,2) DEFAULT 0.00, -- Confidence in duplicate detection
    
    -- Resolution Status
    status ENUM('detected', 'reviewing', 'merged', 'ignored', 'split') DEFAULT 'detected',
    resolution_notes TEXT,
    
    -- Audit Information
    detected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP NULL,
    resolved_by INT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_group_hash (group_hash),
    INDEX idx_master_contact_id (master_contact_id),
    INDEX idx_status (status),
    INDEX idx_confidence_score (confidence_score),
    
    FOREIGN KEY (master_contact_id) REFERENCES contacts(id) ON DELETE SET NULL
);

-- Create contact_duplicate_members for tracking duplicate relationships
CREATE TABLE IF NOT EXISTS contact_duplicate_members (
    id INT PRIMARY KEY AUTO_INCREMENT,
    duplicate_group_id INT NOT NULL,
    contact_id INT NOT NULL,
    
    -- Member Information
    is_master BOOLEAN DEFAULT FALSE,
    similarity_score DECIMAL(5,2) DEFAULT 0.00,
    matching_fields JSON, -- Which fields matched
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE KEY unique_group_contact (duplicate_group_id, contact_id),
    INDEX idx_duplicate_group_id (duplicate_group_id),
    INDEX idx_contact_id (contact_id),
    INDEX idx_is_master (is_master),
    
    FOREIGN KEY (duplicate_group_id) REFERENCES contact_duplicate_groups(id) ON DELETE CASCADE,
    FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE
);

-- Create contact_merge_history for tracking merge operations
CREATE TABLE IF NOT EXISTS contact_merge_history (
    id INT PRIMARY KEY AUTO_INCREMENT,
    
    -- Merge Information
    master_contact_id INT NOT NULL,
    merged_contact_id INT NOT NULL,
    duplicate_group_id INT,
    
    -- Merge Details
    merge_strategy ENUM('keep_master', 'merge_fields', 'manual') DEFAULT 'merge_fields',
    merged_fields JSON, -- Which fields were merged and how
    conflicts_resolved JSON, -- How conflicts were resolved
    
    -- Data Preservation
    original_data JSON, -- Original merged contact data for potential rollback
    
    -- Audit Information
    merged_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    merged_by INT NOT NULL,
    merge_notes TEXT,
    
    INDEX idx_master_contact_id (master_contact_id),
    INDEX idx_merged_contact_id (merged_contact_id),
    INDEX idx_duplicate_group_id (duplicate_group_id),
    INDEX idx_merged_at (merged_at),
    
    FOREIGN KEY (master_contact_id) REFERENCES contacts(id) ON DELETE CASCADE,
    FOREIGN KEY (duplicate_group_id) REFERENCES contact_duplicate_groups(id) ON DELETE SET NULL
);

-- Create contact_import_batches for bulk import tracking
CREATE TABLE IF NOT EXISTS contact_import_batches (
    id INT PRIMARY KEY AUTO_INCREMENT,
    
    -- Batch Information
    batch_name VARCHAR(255) NOT NULL,
    file_name VARCHAR(255),
    file_size_bytes BIGINT DEFAULT 0,
    total_records INT DEFAULT 0,
    
    -- Processing Status
    status ENUM('pending', 'processing', 'completed', 'failed', 'cancelled') DEFAULT 'pending',
    progress_percentage DECIMAL(5,2) DEFAULT 0.00,
    
    -- Results
    records_processed INT DEFAULT 0,
    records_imported INT DEFAULT 0,
    records_updated INT DEFAULT 0,
    records_skipped INT DEFAULT 0,
    records_failed INT DEFAULT 0,
    duplicates_found INT DEFAULT 0,
    
    -- Configuration
    import_options JSON, -- Import configuration options
    field_mapping JSON, -- How CSV fields map to contact fields
    validation_rules JSON, -- Validation rules applied
    
    -- Error Handling
    error_log TEXT, -- Import errors and warnings
    failed_records JSON, -- Records that failed to import
    
    -- Audit Information
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL,
    imported_by INT NOT NULL,
    
    INDEX idx_status (status),
    INDEX idx_started_at (started_at),
    INDEX idx_imported_by (imported_by),
    INDEX idx_batch_name (batch_name)
);

-- Create contact_export_requests for export tracking
CREATE TABLE IF NOT EXISTS contact_export_requests (
    id INT PRIMARY KEY AUTO_INCREMENT,
    
    -- Export Information
    export_name VARCHAR(255) NOT NULL,
    export_format ENUM('csv', 'xlsx', 'json', 'xml') DEFAULT 'csv',
    
    -- Export Criteria
    filters JSON, -- Export filters applied
    fields JSON, -- Fields to include in export
    sort_criteria JSON, -- Sorting preferences
    
    -- Processing Status
    status ENUM('pending', 'processing', 'completed', 'failed', 'expired') DEFAULT 'pending',
    progress_percentage DECIMAL(5,2) DEFAULT 0.00,
    
    -- Results
    total_records INT DEFAULT 0,
    exported_records INT DEFAULT 0,
    file_path VARCHAR(500),
    file_size_bytes BIGINT DEFAULT 0,
    download_url VARCHAR(500),
    
    -- Expiration and Cleanup
    expires_at TIMESTAMP NULL,
    downloaded_count INT DEFAULT 0,
    last_downloaded_at TIMESTAMP NULL,
    
    -- Error Handling
    error_message TEXT,
    
    -- Audit Information
    requested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL,
    requested_by INT NOT NULL,
    
    INDEX idx_status (status),
    INDEX idx_requested_at (requested_at),
    INDEX idx_requested_by (requested_by),
    INDEX idx_expires_at (expires_at)
);

-- Create contact_field_history for tracking field changes
CREATE TABLE IF NOT EXISTS contact_field_history (
    id INT PRIMARY KEY AUTO_INCREMENT,
    contact_id INT NOT NULL,
    
    -- Field Change Information
    field_name VARCHAR(100) NOT NULL,
    old_value TEXT,
    new_value TEXT,
    change_type ENUM('create', 'update', 'delete') NOT NULL,
    
    -- Change Context
    changed_by INT,
    change_reason VARCHAR(255), -- Manual update, import, API, etc.
    change_source VARCHAR(100), -- admin, api, import, automation
    
    -- Validation and Quality
    validation_passed BOOLEAN DEFAULT TRUE,
    validation_errors JSON,
    data_quality_score DECIMAL(3,2) DEFAULT 1.00,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_contact_id (contact_id),
    INDEX idx_field_name (field_name),
    INDEX idx_changed_by (changed_by),
    INDEX idx_change_type (change_type),
    INDEX idx_created_at (created_at),
    
    FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE
);

-- Create contact_validation_rules for data quality
CREATE TABLE IF NOT EXISTS contact_validation_rules (
    id INT PRIMARY KEY AUTO_INCREMENT,
    
    -- Rule Information
    name VARCHAR(255) NOT NULL,
    description TEXT,
    field_name VARCHAR(100) NOT NULL,
    
    -- Validation Configuration
    rule_type ENUM('required', 'format', 'length', 'range', 'custom', 'uniqueness') NOT NULL,
    validation_pattern VARCHAR(500), -- Regex pattern or validation rule
    error_message VARCHAR(255),
    severity ENUM('error', 'warning', 'info') DEFAULT 'error',
    
    -- Rule Execution
    is_active BOOLEAN DEFAULT TRUE,
    execution_order INT DEFAULT 0,
    applies_to_imports BOOLEAN DEFAULT TRUE,
    applies_to_api BOOLEAN DEFAULT TRUE,
    applies_to_manual BOOLEAN DEFAULT TRUE,
    
    -- Usage Tracking
    usage_count INT DEFAULT 0,
    last_used_at TIMESTAMP NULL,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by INT,
    
    INDEX idx_field_name (field_name),
    INDEX idx_rule_type (rule_type),
    INDEX idx_is_active (is_active),
    INDEX idx_execution_order (execution_order)
);

-- Insert default validation rules
INSERT INTO contact_validation_rules (name, description, field_name, rule_type, validation_pattern, error_message, severity) VALUES
('Email Format', 'Validate email address format', 'email', 'format', '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$', 'Invalid email format', 'error'),
('Phone Format', 'Validate phone number format', 'phone', 'format', '^[\+]?[1-9][\d]{0,15}$', 'Invalid phone number format', 'warning'),
('Required First Name', 'First name is required', 'first_name', 'required', '', 'First name is required', 'error'),
('Required Email', 'Email address is required', 'email', 'required', '', 'Email address is required', 'error'),
('Email Uniqueness', 'Email must be unique', 'email', 'uniqueness', '', 'Email address already exists', 'error'),
('Name Length', 'Name must be reasonable length', 'first_name', 'length', '2,100', 'Name must be between 2 and 100 characters', 'warning'),
('Company Length', 'Company name length validation', 'company', 'length', '0,200', 'Company name must be less than 200 characters', 'warning'),
('Website Format', 'Website URL format validation', 'website', 'format', '^https?:\/\/.+', 'Website must be a valid URL', 'warning');