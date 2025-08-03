-- Migration: Create analytics and activity tracking tables
-- Created: 2025-01-01 15:00:00
-- Description: Creates tables for activity logging, system alerts, and analytics tracking

-- Activity Log Table
CREATE TABLE IF NOT EXISTS activity_logs (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED,
    
    -- Activity Information
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id INT UNSIGNED,
    description TEXT NOT NULL,
    
    -- Request Context
    ip_address VARCHAR(45),
    user_agent TEXT,
    request_method VARCHAR(10),
    request_url VARCHAR(500),
    
    -- Metadata
    metadata JSON,
    
    -- Performance Tracking
    execution_time INT, -- Milliseconds
    response_status INT,
    
    -- Audit Fields
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_activity_logs_user (user_id),
    INDEX idx_activity_logs_action (action),
    INDEX idx_activity_logs_entity (entity_type, entity_id),
    INDEX idx_activity_logs_created (created_at),
    
    FOREIGN KEY (user_id) REFERENCES admin_users(id) ON DELETE SET NULL
);

-- System Alerts Table
CREATE TABLE IF NOT EXISTS system_alerts (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    
    -- Alert Information
    type VARCHAR(50) NOT NULL, -- info, warning, error, critical
    priority VARCHAR(20) NOT NULL DEFAULT 'medium', -- low, medium, high, urgent
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    
    -- Targeting
    user_id INT UNSIGNED, -- NULL for global alerts
    role VARCHAR(50), -- NULL for user-specific alerts
    
    -- Status
    is_read BOOLEAN DEFAULT FALSE,
    is_dismissed BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Actions
    action_url VARCHAR(500),
    action_label VARCHAR(100),
    
    -- Expiry
    expires_at TIMESTAMP NULL,
    
    -- Audit Fields
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by INT UNSIGNED,
    
    INDEX idx_system_alerts_user (user_id),
    INDEX idx_system_alerts_type (type),
    INDEX idx_system_alerts_priority (priority),
    INDEX idx_system_alerts_active (is_active),
    INDEX idx_system_alerts_read (is_read),
    INDEX idx_system_alerts_expires (expires_at),
    INDEX idx_system_alerts_created (created_at),
    
    FOREIGN KEY (user_id) REFERENCES admin_users(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES admin_users(id) ON DELETE SET NULL
);

-- Analytics Cache Table (for performance optimization)
CREATE TABLE IF NOT EXISTS analytics_cache (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    
    -- Cache Key
    cache_key VARCHAR(255) NOT NULL UNIQUE,
    metric_type VARCHAR(50) NOT NULL,
    
    -- Time Range
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    granularity VARCHAR(20) NOT NULL,
    
    -- Filters
    filters JSON,
    
    -- Cached Data
    data JSON NOT NULL,
    
    -- Metadata
    calculation_time INT, -- Milliseconds taken to calculate
    record_count INT, -- Number of records processed
    
    -- Cache Management
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    access_count INT DEFAULT 0,
    last_accessed TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_analytics_cache_key (cache_key),
    INDEX idx_analytics_cache_type (metric_type),
    INDEX idx_analytics_cache_dates (start_date, end_date),
    INDEX idx_analytics_cache_expires (expires_at),
    INDEX idx_analytics_cache_accessed (last_accessed)
);

-- User Sessions Table (for activity tracking)
CREATE TABLE IF NOT EXISTS user_sessions (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL,
    
    -- Session Information
    session_token VARCHAR(255) NOT NULL UNIQUE,
    device_info TEXT,
    browser_info TEXT,
    ip_address VARCHAR(45),
    location_info JSON,
    
    -- Session Tracking
    login_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    logout_at TIMESTAMP NULL,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Performance
    page_views INT DEFAULT 0,
    actions_performed INT DEFAULT 0,
    session_duration INT, -- Seconds
    
    INDEX idx_user_sessions_user (user_id),
    INDEX idx_user_sessions_token (session_token),
    INDEX idx_user_sessions_active (is_active),
    INDEX idx_user_sessions_last_activity (last_activity),
    
    FOREIGN KEY (user_id) REFERENCES admin_users(id) ON DELETE CASCADE
);

-- Performance Metrics Table
CREATE TABLE IF NOT EXISTS performance_metrics (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    
    -- Metric Information
    metric_name VARCHAR(100) NOT NULL,
    metric_type VARCHAR(50) NOT NULL, -- response_time, query_time, memory_usage, etc.
    entity_type VARCHAR(50), -- contacts, appointments, users, etc.
    entity_id INT UNSIGNED,
    
    -- Metric Values
    value DECIMAL(10,3) NOT NULL,
    unit VARCHAR(20) NOT NULL, -- ms, seconds, mb, count, percentage
    
    -- Context
    user_id INT UNSIGNED,
    request_id VARCHAR(100),
    endpoint VARCHAR(200),
    
    -- Metadata
    metadata JSON,
    tags JSON,
    
    -- Timestamp
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_performance_metrics_name (metric_name),
    INDEX idx_performance_metrics_type (metric_type),
    INDEX idx_performance_metrics_entity (entity_type, entity_id),
    INDEX idx_performance_metrics_user (user_id),
    INDEX idx_performance_metrics_recorded (recorded_at),
    INDEX idx_performance_metrics_endpoint (endpoint),
    
    FOREIGN KEY (user_id) REFERENCES admin_users(id) ON DELETE SET NULL
);

-- Business Metrics Table (for KPIs and business intelligence)
CREATE TABLE IF NOT EXISTS business_metrics (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    
    -- Metric Information
    metric_name VARCHAR(100) NOT NULL,
    metric_category VARCHAR(50) NOT NULL, -- revenue, conversion, productivity, satisfaction
    
    -- Time Period
    period_type VARCHAR(20) NOT NULL, -- daily, weekly, monthly, quarterly, yearly
    period_date DATE NOT NULL,
    
    -- Metric Values
    value DECIMAL(15,2) NOT NULL,
    target_value DECIMAL(15,2),
    previous_value DECIMAL(15,2),
    
    -- Calculations
    change_amount DECIMAL(15,2),
    change_percentage DECIMAL(5,2),
    is_target_met BOOLEAN DEFAULT FALSE,
    
    -- Dimensions
    department VARCHAR(50),
    user_id INT UNSIGNED,
    source VARCHAR(50),
    
    -- Metadata
    calculation_method TEXT,
    data_sources JSON,
    notes TEXT,
    
    -- Audit
    calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    calculated_by INT UNSIGNED,
    
    INDEX idx_business_metrics_name (metric_name),
    INDEX idx_business_metrics_category (metric_category),
    INDEX idx_business_metrics_period (period_type, period_date),
    INDEX idx_business_metrics_user (user_id),
    INDEX idx_business_metrics_department (department),
    
    UNIQUE KEY unique_metric_period (metric_name, period_type, period_date, department, user_id, source),
    
    FOREIGN KEY (user_id) REFERENCES admin_users(id) ON DELETE SET NULL,
    FOREIGN KEY (calculated_by) REFERENCES admin_users(id) ON DELETE SET NULL
);

-- Export Jobs Table (for analytics export tracking)
CREATE TABLE IF NOT EXISTS export_jobs (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    export_id VARCHAR(50) NOT NULL UNIQUE,
    
    -- Job Information
    job_type VARCHAR(50) NOT NULL, -- analytics, contacts, appointments, reports
    export_format VARCHAR(20) NOT NULL, -- csv, excel, pdf, json
    
    -- Parameters
    start_date DATE,
    end_date DATE,
    filters JSON,
    user_ids JSON,
    
    -- Status
    status VARCHAR(20) DEFAULT 'pending', -- pending, processing, completed, failed
    progress INT DEFAULT 0, -- 0-100 percentage
    
    -- Results
    file_path VARCHAR(500),
    file_size INT, -- Bytes
    record_count INT,
    download_url VARCHAR(500),
    
    -- Error Handling
    error_message TEXT,
    retry_count INT DEFAULT 0,
    max_retries INT DEFAULT 3,
    
    -- Timing
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    expires_at TIMESTAMP NULL,
    
    -- User Info
    created_by INT UNSIGNED NOT NULL,
    downloaded_at TIMESTAMP NULL,
    download_count INT DEFAULT 0,
    
    INDEX idx_export_jobs_export_id (export_id),
    INDEX idx_export_jobs_type (job_type),
    INDEX idx_export_jobs_status (status),
    INDEX idx_export_jobs_created_by (created_by),
    INDEX idx_export_jobs_created (created_at),
    INDEX idx_export_jobs_expires (expires_at),
    
    FOREIGN KEY (created_by) REFERENCES admin_users(id) ON DELETE CASCADE
);

-- Insert sample system alerts for demonstration
INSERT INTO system_alerts (type, priority, title, message, is_active, created_at) VALUES
('info', 'low', 'System Maintenance Scheduled', 'Scheduled maintenance window this Sunday 2 AM - 4 AM EST.', TRUE, NOW()),
('warning', 'medium', 'High Contact Volume', 'Contact submissions are 150% above normal. Consider increasing response capacity.', TRUE, NOW()),
('info', 'low', 'New Feature Available', 'Advanced analytics dashboard is now available in the Analytics section.', TRUE, NOW());

-- Create initial business metrics for current month
INSERT INTO business_metrics 
(metric_name, metric_category, period_type, period_date, value, target_value, department, calculation_method, calculated_at)
VALUES
('Total Contacts', 'productivity', 'monthly', DATE_FORMAT(NOW(), '%Y-%m-01'), 0, 500, 'sales', 'COUNT(*) FROM contacts WHERE MONTH(created_at) = MONTH(NOW())', NOW()),
('Conversion Rate', 'conversion', 'monthly', DATE_FORMAT(NOW(), '%Y-%m-01'), 0, 15.00, 'sales', 'Converted contacts / Total contacts * 100', NOW()),
('Average Response Time', 'productivity', 'monthly', DATE_FORMAT(NOW(), '%Y-%m-01'), 0, 2.00, 'support', 'AVG(TIMESTAMPDIFF(HOUR, created_at, first_response_date))', NOW()),
('Customer Satisfaction', 'satisfaction', 'monthly', DATE_FORMAT(NOW(), '%Y-%m-01'), 0, 4.50, 'support', 'AVG(rating) FROM appointments WHERE status = completed', NOW());