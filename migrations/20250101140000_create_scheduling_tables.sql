-- Migration: Create scheduling and appointment management tables
-- Created: 2025-01-01 14:00:00
-- Description: Creates tables for appointment scheduling, user availability, and calendar management

-- Appointments Table
CREATE TABLE IF NOT EXISTS appointments (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    contact_id INT UNSIGNED NOT NULL,
    assigned_to INT UNSIGNED NOT NULL,
    
    -- Basic Information
    title VARCHAR(255) NOT NULL,
    description TEXT,
    type ENUM('call', 'meeting', 'demo', 'consultation', 'followup', 'presentation', 'negotiation', 'other') NOT NULL,
    priority ENUM('low', 'medium', 'high', 'urgent') DEFAULT 'medium',
    status ENUM('scheduled', 'confirmed', 'in_progress', 'completed', 'cancelled', 'rescheduled', 'no_show') DEFAULT 'scheduled',
    
    -- Scheduling Information
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    duration INT NOT NULL, -- Minutes
    timezone VARCHAR(50) DEFAULT 'UTC',
    
    -- Location Information
    location VARCHAR(500),
    is_virtual BOOLEAN DEFAULT FALSE,
    meeting_url VARCHAR(500),
    meeting_id VARCHAR(100),
    meeting_password VARCHAR(100),
    
    -- Recurrence
    is_recurring BOOLEAN DEFAULT FALSE,
    recurrence JSON,
    parent_id INT UNSIGNED,
    series_id VARCHAR(100), -- UUID for recurring series
    
    -- Participants
    attendees JSON, -- Additional attendees
    
    -- Notifications
    notifications JSON,
    reminders_sent JSON,
    
    -- Status Tracking
    confirmed_at TIMESTAMP NULL,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    cancelled_at TIMESTAMP NULL,
    cancel_reason TEXT,
    
    -- Follow-up
    follow_up_required BOOLEAN DEFAULT FALSE,
    follow_up_date TIMESTAMP NULL,
    follow_up_notes TEXT,
    
    -- Outcome
    outcome TEXT,
    rating INT, -- 1-5 scale
    next_steps TEXT,
    
    -- Metadata
    external_id VARCHAR(255), -- For calendar integration
    calendar_id VARCHAR(255),
    tags JSON,
    custom_fields JSON,
    
    -- Audit Fields
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by INT UNSIGNED,
    updated_by INT UNSIGNED,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_appointments_contact (contact_id),
    INDEX idx_appointments_assigned_to (assigned_to),
    INDEX idx_appointments_type (type),
    INDEX idx_appointments_priority (priority),
    INDEX idx_appointments_status (status),
    INDEX idx_appointments_start_time (start_time),
    INDEX idx_appointments_end_time (end_time),
    INDEX idx_appointments_series (series_id),
    INDEX idx_appointments_external (external_id),
    INDEX idx_appointments_deleted (deleted_at),
    INDEX idx_appointments_parent (parent_id),
    
    FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE,
    FOREIGN KEY (assigned_to) REFERENCES admin_users(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES admin_users(id) ON DELETE SET NULL,
    FOREIGN KEY (updated_by) REFERENCES admin_users(id) ON DELETE SET NULL,
    FOREIGN KEY (parent_id) REFERENCES appointments(id) ON DELETE CASCADE
);

-- User Availabilities Table
CREATE TABLE IF NOT EXISTS user_availabilities (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL,
    
    -- Schedule Information
    day_of_week INT NOT NULL, -- 0=Sunday, 1=Monday, etc.
    start_time TIME NOT NULL, -- "09:00:00"
    end_time TIME NOT NULL,   -- "17:00:00"
    is_available BOOLEAN DEFAULT TRUE,
    
    -- Break Times
    break_times JSON, -- Array of {start, end, title}
    
    -- Timezone
    timezone VARCHAR(50) DEFAULT 'UTC',
    
    -- Metadata
    notes TEXT,
    
    -- Audit Fields
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by INT UNSIGNED,
    updated_by INT UNSIGNED,
    
    INDEX idx_user_availabilities_user (user_id),
    INDEX idx_user_availabilities_day (day_of_week),
    INDEX idx_user_availabilities_available (is_available),
    UNIQUE KEY unique_user_day (user_id, day_of_week),
    
    FOREIGN KEY (user_id) REFERENCES admin_users(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES admin_users(id) ON DELETE SET NULL,
    FOREIGN KEY (updated_by) REFERENCES admin_users(id) ON DELETE SET NULL
);

-- Availability Exceptions Table
CREATE TABLE IF NOT EXISTS availability_exceptions (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL,
    
    -- Exception Information
    date DATE NOT NULL,
    is_available BOOLEAN DEFAULT FALSE,
    start_time TIME, -- Override start time if available
    end_time TIME,   -- Override end time if available
    
    -- Details
    reason VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Recurrence (for holidays, etc.)
    is_recurring BOOLEAN DEFAULT FALSE,
    recurrence JSON,
    
    -- Audit Fields
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by INT UNSIGNED,
    updated_by INT UNSIGNED,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_availability_exceptions_user (user_id),
    INDEX idx_availability_exceptions_date (date),
    INDEX idx_availability_exceptions_available (is_available),
    INDEX idx_availability_exceptions_deleted (deleted_at),
    
    FOREIGN KEY (user_id) REFERENCES admin_users(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES admin_users(id) ON DELETE SET NULL,
    FOREIGN KEY (updated_by) REFERENCES admin_users(id) ON DELETE SET NULL
);

-- Appointment Reminders Table
CREATE TABLE IF NOT EXISTS appointment_reminders (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    appointment_id INT UNSIGNED NOT NULL,
    
    -- Reminder Configuration
    method ENUM('email', 'sms', 'push', 'slack') NOT NULL,
    minutes_before INT NOT NULL,
    message TEXT NOT NULL,
    
    -- Status
    is_sent BOOLEAN DEFAULT FALSE,
    sent_at TIMESTAMP NULL,
    scheduled_for TIMESTAMP NOT NULL,
    
    -- Error Tracking
    attempts INT DEFAULT 0,
    last_error TEXT,
    
    -- Audit Fields
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_appointment_reminders_appointment (appointment_id),
    INDEX idx_appointment_reminders_sent (is_sent),
    INDEX idx_appointment_reminders_scheduled (scheduled_for),
    INDEX idx_appointment_reminders_method (method),
    
    FOREIGN KEY (appointment_id) REFERENCES appointments(id) ON DELETE CASCADE
);

-- Calendar Integrations Table
CREATE TABLE IF NOT EXISTS calendar_integrations (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL,
    
    -- Integration Information
    provider VARCHAR(50) NOT NULL, -- google, outlook, apple, etc.
    external_id VARCHAR(255) NOT NULL,
    calendar_name VARCHAR(255) NOT NULL,
    
    -- Configuration
    is_enabled BOOLEAN DEFAULT TRUE,
    sync_direction VARCHAR(20) DEFAULT 'bidirectional', -- read, write, bidirectional
    auto_sync BOOLEAN DEFAULT TRUE,
    
    -- Authentication (encrypted in production)
    access_token TEXT,
    refresh_token TEXT,
    token_expires_at TIMESTAMP NULL,
    
    -- Sync Status
    last_sync_at TIMESTAMP NULL,
    last_sync_status VARCHAR(50) DEFAULT 'pending',
    sync_errors JSON,
    
    -- Settings
    default_reminders JSON,
    settings JSON,
    
    -- Audit Fields
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by INT UNSIGNED,
    updated_by INT UNSIGNED,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_calendar_integrations_user (user_id),
    INDEX idx_calendar_integrations_provider (provider),
    INDEX idx_calendar_integrations_enabled (is_enabled),
    INDEX idx_calendar_integrations_deleted (deleted_at),
    
    FOREIGN KEY (user_id) REFERENCES admin_users(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES admin_users(id) ON DELETE SET NULL,
    FOREIGN KEY (updated_by) REFERENCES admin_users(id) ON DELETE SET NULL
);

-- Insert default availability for existing users (Monday-Friday 9 AM - 5 PM)
INSERT INTO user_availabilities (user_id, day_of_week, start_time, end_time, is_available, timezone, created_at)
SELECT 
    id as user_id,
    day_num as day_of_week,
    '09:00:00' as start_time,
    '17:00:00' as end_time,
    CASE WHEN day_num IN (1,2,3,4,5) THEN TRUE ELSE FALSE END as is_available,
    'UTC' as timezone,
    NOW() as created_at
FROM admin_users
CROSS JOIN (
    SELECT 0 as day_num UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 
    UNION SELECT 4 UNION SELECT 5 UNION SELECT 6
) days
WHERE is_active = TRUE
ON DUPLICATE KEY UPDATE updated_at = NOW();