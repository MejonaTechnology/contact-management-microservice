-- Create lead_scoring_rules table for configurable lead scoring
CREATE TABLE IF NOT EXISTS lead_scoring_rules (
    id INT PRIMARY KEY AUTO_INCREMENT,
    
    -- Rule Information
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category ENUM('demographic', 'behavioral', 'engagement', 'firmographic', 'lifecycle') NOT NULL,
    
    -- Rule Configuration
    field_name VARCHAR(100) NOT NULL, -- Field to evaluate
    operator ENUM('equals', 'not_equals', 'contains', 'not_contains', 'greater_than', 'less_than', 'between', 'in', 'not_in', 'exists', 'not_exists') NOT NULL,
    field_value TEXT, -- Value to compare against
    
    -- Scoring
    score_points INT NOT NULL, -- Points to add/subtract
    score_type ENUM('add', 'subtract', 'set') DEFAULT 'add',
    max_score_per_rule INT DEFAULT NULL, -- Maximum points this rule can contribute
    
    -- Rule Execution
    is_active BOOLEAN DEFAULT TRUE,
    execution_order INT DEFAULT 0, -- Order of rule execution
    frequency ENUM('once', 'daily', 'weekly', 'on_change', 'always') DEFAULT 'once',
    
    -- Conditions and Dependencies
    conditions JSON, -- Additional conditions for rule execution
    dependent_rules JSON, -- Rules that must be satisfied first
    
    -- Time-based Rules
    time_window_days INT DEFAULT NULL, -- Time window for the rule
    decay_rate DECIMAL(5,2) DEFAULT 0.00, -- Score decay over time
    
    -- Usage Tracking
    usage_count INT DEFAULT 0,
    last_used_at TIMESTAMP NULL,
    
    -- Audit Fields
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by INT,
    updated_by INT,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_category (category),
    INDEX idx_field_name (field_name),
    INDEX idx_is_active (is_active),
    INDEX idx_execution_order (execution_order),
    INDEX idx_score_points (score_points)
);

-- Insert default lead scoring rules
INSERT INTO lead_scoring_rules (name, description, category, field_name, operator, field_value, score_points, score_type, execution_order) VALUES
-- Demographic Scoring
('Email Domain - Business', 'Business email domains get higher scores', 'demographic', 'email', 'not_contains', '@gmail.com,@yahoo.com,@hotmail.com,@outlook.com', 10, 'add', 1),
('Phone Number Provided', 'Contacts who provide phone numbers', 'demographic', 'phone', 'exists', '', 15, 'add', 2),
('Company Information', 'Contacts who provide company details', 'demographic', 'company', 'exists', '', 12, 'add', 3),
('Job Title - Decision Maker', 'Decision maker job titles', 'demographic', 'job_title', 'contains', 'CEO,CTO,Manager,Director,President,VP,Head', 20, 'add', 4),

-- Firmographic Scoring
('Enterprise Contact', 'Large company contacts', 'firmographic', 'tags', 'contains', 'Enterprise', 25, 'add', 5),
('Technology Industry', 'Technology sector contacts', 'firmographic', 'tags', 'contains', 'Technology', 15, 'add', 6),
('Local Business', 'Local area businesses', 'firmographic', 'tags', 'contains', 'Local', 10, 'add', 7),

-- Behavioral Scoring
('Website Visit Frequency', 'Multiple website visits', 'behavioral', 'custom_fields->visit_count', 'greater_than', '3', 15, 'add', 8),
('Downloaded Resources', 'Downloaded case studies, whitepapers', 'behavioral', 'custom_fields->downloads', 'greater_than', '0', 20, 'add', 9),
('Pricing Page Visit', 'Visited pricing or service pages', 'behavioral', 'landing_page', 'contains', 'pricing,services,packages', 18, 'add', 10),

-- Engagement Scoring
('Email Opened', 'Opened marketing emails', 'engagement', 'email_opened', 'equals', 'true', 8, 'add', 11),
('Email Clicked', 'Clicked links in emails', 'engagement', 'email_clicked', 'equals', 'true', 12, 'add', 12),
('Response to Email', 'Replied to emails', 'engagement', 'total_interactions', 'greater_than', '0', 25, 'add', 13),
('Social Media Engagement', 'Social media interaction', 'engagement', 'utm_medium', 'contains', 'social', 10, 'add', 14),

-- Lifecycle Scoring
('New Contact Bonus', 'Bonus for new contacts', 'lifecycle', 'status', 'equals', 'new', 5, 'add', 15),
('Contacted Status', 'Moved to contacted status', 'lifecycle', 'status', 'equals', 'contacted', 10, 'add', 16),
('Qualified Lead', 'Successfully qualified lead', 'lifecycle', 'status', 'equals', 'qualified', 30, 'add', 17),
('High Value Potential', 'High estimated value', 'lifecycle', 'estimated_value', 'greater_than', '50000', 35, 'add', 18),

-- Negative Scoring
('Personal Email Penalty', 'Personal email domains get lower scores', 'demographic', 'email', 'contains', '@gmail.com,@yahoo.com,@hotmail.com', -5, 'add', 19),
('No Response Penalty', 'No response after multiple attempts', 'engagement', 'total_interactions', 'equals', '0', -10, 'add', 20),
('Unsubscribed Penalty', 'Unsubscribed from communications', 'engagement', 'unsubscribed', 'equals', 'true', -50, 'add', 21),
('Closed Lost Penalty', 'Previously closed as lost', 'lifecycle', 'status', 'equals', 'closed_lost', -30, 'add', 22);

-- Create contact_scores table for tracking score history
CREATE TABLE IF NOT EXISTS contact_scores (
    id INT PRIMARY KEY AUTO_INCREMENT,
    contact_id INT NOT NULL,
    
    -- Score Information
    current_score INT DEFAULT 0,
    previous_score INT DEFAULT 0,
    score_change INT DEFAULT 0,
    
    -- Score Components
    demographic_score INT DEFAULT 0,
    behavioral_score INT DEFAULT 0,
    engagement_score INT DEFAULT 0,
    firmographic_score INT DEFAULT 0,
    lifecycle_score INT DEFAULT 0,
    
    -- Score Categories
    score_grade ENUM('A', 'B', 'C', 'D', 'F') DEFAULT 'C',
    score_category ENUM('hot', 'warm', 'cold', 'frozen') DEFAULT 'cold',
    
    -- Calculation Details
    calculation_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    rules_applied JSON, -- Rules that were applied
    calculation_metadata JSON, -- Additional calculation details
    
    -- Decay and Time-based Adjustments
    decay_applied BOOLEAN DEFAULT FALSE,
    decay_amount INT DEFAULT 0,
    last_activity_date TIMESTAMP NULL,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_contact_id (contact_id),
    INDEX idx_current_score (current_score),
    INDEX idx_score_grade (score_grade),
    INDEX idx_score_category (score_category),
    INDEX idx_calculation_date (calculation_date),
    
    FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE
);

-- Create score_calculation_log for audit trail
CREATE TABLE IF NOT EXISTS score_calculation_log (
    id INT PRIMARY KEY AUTO_INCREMENT,
    contact_id INT NOT NULL,
    rule_id INT,
    
    -- Calculation Details
    rule_name VARCHAR(255),
    field_evaluated VARCHAR(100),
    field_value TEXT,
    score_change INT NOT NULL,
    previous_score INT DEFAULT 0,
    new_score INT DEFAULT 0,
    
    -- Execution Context
    calculation_batch_id VARCHAR(100), -- Group related calculations
    execution_time_ms INT DEFAULT 0,
    
    -- Metadata
    calculation_metadata JSON,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_contact_id (contact_id),
    INDEX idx_rule_id (rule_id),
    INDEX idx_calculation_batch_id (calculation_batch_id),
    INDEX idx_created_at (created_at),
    
    FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE,
    FOREIGN KEY (rule_id) REFERENCES lead_scoring_rules(id) ON DELETE SET NULL
);

-- Create score_thresholds for defining score ranges
CREATE TABLE IF NOT EXISTS score_thresholds (
    id INT PRIMARY KEY AUTO_INCREMENT,
    
    -- Threshold Configuration
    name VARCHAR(100) NOT NULL,
    min_score INT NOT NULL,
    max_score INT NOT NULL,
    grade ENUM('A', 'B', 'C', 'D', 'F') NOT NULL,
    category ENUM('hot', 'warm', 'cold', 'frozen') NOT NULL,
    color VARCHAR(7) DEFAULT '#007bff', -- UI color
    
    -- Actions and Automation
    auto_assign_rules JSON, -- Automatic assignment rules
    notification_rules JSON, -- Notification triggers
    follow_up_rules JSON, -- Follow-up automation
    
    -- Description and Usage
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_score_range (min_score, max_score),
    INDEX idx_grade (grade),
    INDEX idx_category (category),
    INDEX idx_is_active (is_active)
);

-- Insert default score thresholds
INSERT INTO score_thresholds (name, min_score, max_score, grade, category, color, description) VALUES
('Hot Leads (A)', 80, 100, 'A', 'hot', '#dc3545', 'High-priority leads with strong buying signals'),
('Warm Leads (B)', 60, 79, 'B', 'warm', '#fd7e14', 'Good potential leads requiring follow-up'),
('Moderate Leads (C)', 40, 59, 'C', 'warm', '#ffc107', 'Average leads needing nurturing'),
('Cold Leads (D)', 20, 39, 'D', 'cold', '#6c757d', 'Low-priority leads requiring long-term nurturing'),
('Poor Leads (F)', 0, 19, 'F', 'frozen', '#343a40', 'Very low quality leads, consider removing');