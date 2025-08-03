-- Migration: Create assignment tables
-- Created: 2025-01-01 12:00:00
-- Description: Creates tables for contact assignment and routing functionality

-- Assignment Rules Table
CREATE TABLE IF NOT EXISTS assignment_rules (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type ENUM('round_robin', 'load_based', 'skill_based', 'geography_based', 'value_based', 'custom') NOT NULL,
    status ENUM('active', 'inactive', 'paused') DEFAULT 'active',
    priority INT DEFAULT 0,
    
    -- Rule Configuration
    conditions JSON,
    settings JSON,
    
    -- Assignment Targets
    assignee_ids JSON,
    fallback_user_id INT UNSIGNED,
    
    -- Business Hours and Availability
    business_hours_enabled BOOLEAN DEFAULT FALSE,
    business_hours_start VARCHAR(5), -- "09:00"
    business_hours_end VARCHAR(5),   -- "17:00"
    working_days JSON,               -- ["mon","tue","wed","thu","fri"]
    timezone VARCHAR(50) DEFAULT 'UTC',
    
    -- Rate Limiting
    max_assignments_per_hour INT,
    max_assignments_per_day INT,
    
    -- Tracking
    total_assignments INT DEFAULT 0,
    successful_assignments INT DEFAULT 0,
    last_assignment_at TIMESTAMP NULL,
    
    -- Audit Fields
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by INT UNSIGNED,
    updated_by INT UNSIGNED,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_assignment_rules_status (status),
    INDEX idx_assignment_rules_type (type),
    INDEX idx_assignment_rules_priority (priority),
    INDEX idx_assignment_rules_deleted (deleted_at),
    
    FOREIGN KEY (fallback_user_id) REFERENCES admin_users(id) ON DELETE SET NULL,
    FOREIGN KEY (created_by) REFERENCES admin_users(id) ON DELETE SET NULL,
    FOREIGN KEY (updated_by) REFERENCES admin_users(id) ON DELETE SET NULL
);

-- Contact Assignments Table
CREATE TABLE IF NOT EXISTS contact_assignments (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    contact_id INT UNSIGNED NOT NULL,
    assigned_to_id INT UNSIGNED NOT NULL,
    assigned_by_id INT UNSIGNED, -- Null for automatic assignments
    rule_id INT UNSIGNED,        -- Which rule triggered this assignment
    
    -- Assignment Details
    assignment_type VARCHAR(50) DEFAULT 'automatic', -- automatic, manual
    assignment_reason TEXT,
    priority ENUM('low', 'medium', 'high', 'urgent') DEFAULT 'medium',
    
    -- Status Tracking
    status VARCHAR(50) DEFAULT 'active', -- active, reassigned, completed, cancelled, unassigned
    accepted_at TIMESTAMP NULL,
    first_response_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    
    -- Performance Metrics
    response_time_hours INT DEFAULT 0,
    resolution_time_hours INT DEFAULT 0,
    interaction_count INT DEFAULT 0,
    
    -- Audit Fields
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_contact_assignments_contact (contact_id),
    INDEX idx_contact_assignments_assigned_to (assigned_to_id),
    INDEX idx_contact_assignments_status (status),
    INDEX idx_contact_assignments_rule (rule_id),
    INDEX idx_contact_assignments_created (created_at),
    
    FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE,
    FOREIGN KEY (assigned_to_id) REFERENCES admin_users(id) ON DELETE CASCADE,
    FOREIGN KEY (assigned_by_id) REFERENCES admin_users(id) ON DELETE SET NULL,
    FOREIGN KEY (rule_id) REFERENCES assignment_rules(id) ON DELETE SET NULL
);

-- User Workload Table
CREATE TABLE IF NOT EXISTS user_workloads (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL UNIQUE,
    
    -- Current Workload
    active_contacts INT DEFAULT 0,
    pending_contacts INT DEFAULT 0,
    overdue_contacts INT DEFAULT 0,
    total_contacts INT DEFAULT 0,
    
    -- Daily Metrics
    today_assignments INT DEFAULT 0,
    today_responses INT DEFAULT 0,
    today_completions INT DEFAULT 0,
    
    -- Weekly Metrics
    weekly_assignments INT DEFAULT 0,
    weekly_responses INT DEFAULT 0,
    weekly_completions INT DEFAULT 0,
    
    -- Performance Metrics
    avg_response_time_hours DECIMAL(8,2) DEFAULT 0.00,
    avg_resolution_time_hours DECIMAL(8,2) DEFAULT 0.00,
    conversion_rate DECIMAL(5,4) DEFAULT 0.0000,
    
    -- Availability
    is_available BOOLEAN DEFAULT TRUE,
    max_daily_assignments INT,
    max_active_contacts INT,
    
    -- Skills and Specialties
    skills JSON,
    territories JSON,
    contact_types JSON,
    
    -- Audit Fields
    last_calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_user_workloads_user (user_id),
    INDEX idx_user_workloads_available (is_available),
    INDEX idx_user_workloads_calculated (last_calculated_at),
    
    FOREIGN KEY (user_id) REFERENCES admin_users(id) ON DELETE CASCADE
);

-- Assignment History Table
CREATE TABLE IF NOT EXISTS assignment_history (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    contact_id INT UNSIGNED NOT NULL,
    from_user_id INT UNSIGNED, -- Null for initial assignments
    to_user_id INT UNSIGNED NOT NULL,
    changed_by_id INT UNSIGNED,
    rule_id INT UNSIGNED,
    
    -- Change Details
    change_type VARCHAR(50) NOT NULL, -- assigned, reassigned, unassigned
    change_reason TEXT,
    previous_status VARCHAR(50),
    new_status VARCHAR(50),
    
    -- Context
    business_context JSON,
    system_context JSON,
    
    -- Audit Fields
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_assignment_history_contact (contact_id),
    INDEX idx_assignment_history_from_user (from_user_id),
    INDEX idx_assignment_history_to_user (to_user_id),
    INDEX idx_assignment_history_rule (rule_id),
    INDEX idx_assignment_history_created (created_at),
    
    FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE,
    FOREIGN KEY (from_user_id) REFERENCES admin_users(id) ON DELETE SET NULL,
    FOREIGN KEY (to_user_id) REFERENCES admin_users(id) ON DELETE CASCADE,
    FOREIGN KEY (changed_by_id) REFERENCES admin_users(id) ON DELETE SET NULL,
    FOREIGN KEY (rule_id) REFERENCES assignment_rules(id) ON DELETE SET NULL
);