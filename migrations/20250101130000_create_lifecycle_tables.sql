-- Migration: Create lifecycle management tables
-- Created: 2025-01-01 13:00:00
-- Description: Creates tables for contact lifecycle management, lead scoring, and status transitions

-- Lead Scoring Rules Table
CREATE TABLE IF NOT EXISTS lead_scoring_rules (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    priority INT DEFAULT 0,
    
    -- Scoring Configuration
    category VARCHAR(100), -- demographic, behavioral, engagement, firmographic
    base_score INT DEFAULT 0,
    max_score INT DEFAULT 100,
    criteria JSON,
    
    -- Conditions for when this rule applies
    applicable_when JSON,
    
    -- Tracking
    times_applied INT DEFAULT 0,
    last_applied_at TIMESTAMP NULL,
    
    -- Audit Fields
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by INT UNSIGNED,
    updated_by INT UNSIGNED,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_lead_scoring_rules_active (is_active),
    INDEX idx_lead_scoring_rules_priority (priority),
    INDEX idx_lead_scoring_rules_category (category),
    INDEX idx_lead_scoring_rules_deleted (deleted_at),
    
    FOREIGN KEY (created_by) REFERENCES admin_users(id) ON DELETE SET NULL,
    FOREIGN KEY (updated_by) REFERENCES admin_users(id) ON DELETE SET NULL
);

-- Status Transition Rules Table
CREATE TABLE IF NOT EXISTS status_transition_rules (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    priority INT DEFAULT 0,
    
    -- Transition Configuration
    from_status ENUM('new', 'contacted', 'qualified', 'proposal', 'negotiation', 'closed_won', 'closed_lost', 'on_hold', 'nurturing') NOT NULL,
    to_status ENUM('new', 'contacted', 'qualified', 'proposal', 'negotiation', 'closed_won', 'closed_lost', 'on_hold', 'nurturing') NOT NULL,
    transition_type ENUM('automatic', 'manual', 'scheduled', 'triggered') DEFAULT 'automatic',
    
    -- Conditions for transition
    conditions JSON,
    required_score INT DEFAULT 0,
    days_in_status INT DEFAULT 0,
    
    -- Actions to perform on transition
    actions JSON,
    notify_users JSON,
    
    -- Tracking
    times_triggered INT DEFAULT 0,
    last_triggered_at TIMESTAMP NULL,
    
    -- Audit Fields
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by INT UNSIGNED,
    updated_by INT UNSIGNED,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_status_transition_rules_active (is_active),
    INDEX idx_status_transition_rules_priority (priority),
    INDEX idx_status_transition_rules_from_status (from_status),
    INDEX idx_status_transition_rules_to_status (to_status),
    INDEX idx_status_transition_rules_deleted (deleted_at),
    
    FOREIGN KEY (created_by) REFERENCES admin_users(id) ON DELETE SET NULL,
    FOREIGN KEY (updated_by) REFERENCES admin_users(id) ON DELETE SET NULL
);

-- Contact Lifecycles Table
CREATE TABLE IF NOT EXISTS contact_lifecycles (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    contact_id INT UNSIGNED NOT NULL UNIQUE,
    
    -- Current State
    current_stage ENUM('unknown', 'suspect', 'prospect', 'marketing_qualified_lead', 'sales_qualified_lead', 'opportunity', 'customer', 'evangelist', 'other') DEFAULT 'unknown',
    current_status ENUM('new', 'contacted', 'qualified', 'proposal', 'negotiation', 'closed_won', 'closed_lost', 'on_hold', 'nurturing'),
    current_score INT DEFAULT 0,
    
    -- Lifecycle Progression
    stage_entered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status_entered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_scored_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Scoring Breakdown
    demographic_score INT DEFAULT 0,
    behavioral_score INT DEFAULT 0,
    engagement_score INT DEFAULT 0,
    firmographic_score INT DEFAULT 0,
    
    -- Scoring History
    score_history JSON,
    scoring_factors JSON,
    
    -- Milestones
    first_engagement_at TIMESTAMP NULL,
    qualification_at TIMESTAMP NULL,
    opportunity_at TIMESTAMP NULL,
    conversion_at TIMESTAMP NULL,
    
    -- Performance Metrics
    days_in_current_stage INT DEFAULT 0,
    days_in_current_status INT DEFAULT 0,
    total_lifecycle_days INT DEFAULT 0,
    
    -- Velocity Metrics
    stage_velocity JSON, -- Days spent in each stage
    conversion_rate DECIMAL(5,4) DEFAULT 0.0000,
    
    -- Audit Fields
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_contact_lifecycles_contact (contact_id),
    INDEX idx_contact_lifecycles_stage (current_stage),
    INDEX idx_contact_lifecycles_status (current_status),
    INDEX idx_contact_lifecycles_score (current_score),
    INDEX idx_contact_lifecycles_stage_entered (stage_entered_at),
    INDEX idx_contact_lifecycles_status_entered (status_entered_at),
    
    FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE
);

-- Lifecycle Events Table
CREATE TABLE IF NOT EXISTS lifecycle_events (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    contact_id INT UNSIGNED NOT NULL,
    lifecycle_id INT UNSIGNED NOT NULL,
    
    -- Event Details
    event_type VARCHAR(100) NOT NULL, -- score_change, status_change, stage_change, milestone, action
    event_name VARCHAR(255) NOT NULL,
    event_description TEXT,
    
    -- Before/After State
    previous_value VARCHAR(255),
    new_value VARCHAR(255) NOT NULL,
    change_amount INT, -- For score changes
    
    -- Context
    trigger_type VARCHAR(50), -- automatic, manual, scheduled, api
    trigger_source VARCHAR(100), -- rule_id, user_id, system
    trigger_data JSON,
    
    -- User who caused the event (if manual)
    triggered_by INT UNSIGNED,
    
    -- Audit Fields
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_lifecycle_events_contact (contact_id),
    INDEX idx_lifecycle_events_lifecycle (lifecycle_id),
    INDEX idx_lifecycle_events_type (event_type),
    INDEX idx_lifecycle_events_triggered_by (triggered_by),
    INDEX idx_lifecycle_events_created (created_at),
    
    FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE,
    FOREIGN KEY (lifecycle_id) REFERENCES contact_lifecycles(id) ON DELETE CASCADE,
    FOREIGN KEY (triggered_by) REFERENCES admin_users(id) ON DELETE SET NULL
);