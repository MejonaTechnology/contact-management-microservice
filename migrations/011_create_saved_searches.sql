-- Migration: Create saved searches table
-- This table stores user-saved search queries for contacts

CREATE TABLE saved_searches (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    criteria JSON NOT NULL,
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_saved_searches_user_id (user_id),
    INDEX idx_saved_searches_name (name),
    INDEX idx_saved_searches_is_public (is_public),
    UNIQUE KEY unique_user_search_name (user_id, name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert some default public saved searches for common use cases
INSERT INTO saved_searches (user_id, name, description, criteria, is_public) VALUES
(1, 'Hot Leads', 'Contacts with high lead scores (80+)', 
 '{"lead_score_min": 80, "status": "new", "sort_by": "lead_score", "sort_order": "DESC"}', true),

(1, 'High Priority Contacts', 'All high priority contacts requiring immediate attention', 
 '{"priority": "high", "sort_by": "created_at", "sort_order": "DESC"}', true),

(1, 'Recent Contacts', 'Contacts created in the last 7 days', 
 '{"created_from_days": 7, "sort_by": "created_at", "sort_order": "DESC"}', true),

(1, 'Qualified Leads', 'Contacts that have been qualified and are in negotiation', 
 '{"status": "qualified", "sort_by": "estimated_value", "sort_order": "DESC"}', true),

(1, 'Unassigned Contacts', 'Contacts that haven\'t been assigned to anyone yet', 
 '{"assigned_to": null, "status": "new", "sort_by": "created_at", "sort_order": "ASC"}', true),

(1, 'Enterprise Prospects', 'High-value prospects with significant estimated value', 
 '{"estimated_value_min": 50000, "sort_by": "estimated_value", "sort_order": "DESC"}', true);