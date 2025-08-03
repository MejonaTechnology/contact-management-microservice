-- Create contact_tags table for flexible contact categorization
CREATE TABLE IF NOT EXISTS contact_tags (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    color VARCHAR(7) DEFAULT '#007bff', -- Hex color for UI display
    category VARCHAR(50), -- Group tags by category
    is_system BOOLEAN DEFAULT FALSE, -- System-generated vs user-created
    usage_count INT DEFAULT 0, -- Track tag usage frequency
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by INT,
    
    INDEX idx_name (name),
    INDEX idx_category (category),
    INDEX idx_usage_count (usage_count)
);

-- Create junction table for many-to-many relationship
CREATE TABLE IF NOT EXISTS contact_tag_assignments (
    id INT PRIMARY KEY AUTO_INCREMENT,
    contact_id INT NOT NULL,
    tag_id INT NOT NULL,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    assigned_by INT,
    
    UNIQUE KEY unique_contact_tag (contact_id, tag_id),
    INDEX idx_contact_id (contact_id),
    INDEX idx_tag_id (tag_id),
    INDEX idx_assigned_at (assigned_at),
    
    FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES contact_tags(id) ON DELETE CASCADE
);

-- Insert default tags
INSERT INTO contact_tags (name, description, color, category, is_system) VALUES
-- Lead Quality Tags
('Hot Lead', 'High-priority, high-conversion potential lead', '#dc3545', 'quality', TRUE),
('Warm Lead', 'Engaged lead with medium conversion potential', '#fd7e14', 'quality', TRUE),
('Cold Lead', 'Low engagement, requires nurturing', '#6c757d', 'quality', TRUE),
('Qualified', 'Lead has been qualified by sales team', '#28a745', 'quality', TRUE),

-- Industry Tags
('Technology', 'Technology sector contacts', '#007bff', 'industry', FALSE),
('Healthcare', 'Healthcare and medical contacts', '#dc3545', 'industry', FALSE),
('Finance', 'Financial services contacts', '#28a745', 'industry', FALSE),
('Education', 'Educational institutions', '#6f42c1', 'industry', FALSE),
('Retail', 'Retail and e-commerce contacts', '#fd7e14', 'industry', FALSE),
('Manufacturing', 'Manufacturing sector contacts', '#6c757d', 'industry', FALSE),

-- Engagement Level
('Highly Engaged', 'Very active and responsive contact', '#28a745', 'engagement', TRUE),
('Moderately Engaged', 'Some engagement and responses', '#ffc107', 'engagement', TRUE),
('Low Engagement', 'Minimal interaction and responses', '#dc3545', 'engagement', TRUE),

-- Business Size
('Enterprise', 'Large enterprise clients (500+ employees)', '#6f42c1', 'size', FALSE),
('Mid-Market', 'Medium-sized businesses (50-500 employees)', '#007bff', 'size', FALSE),
('Small Business', 'Small businesses (1-50 employees)', '#28a745', 'size', FALSE),
('Startup', 'Startup companies', '#fd7e14', 'size', FALSE),

-- Geographic Tags
('Local', 'Local area contacts', '#17a2b8', 'geography', FALSE),
('National', 'National level contacts', '#007bff', 'geography', FALSE),
('International', 'International contacts', '#6f42c1', 'geography', FALSE),

-- Service Interest
('Web Development', 'Interested in web development services', '#007bff', 'service', FALSE),
('Mobile Development', 'Interested in mobile app development', '#28a745', 'service', FALSE),
('Digital Marketing', 'Interested in marketing services', '#fd7e14', 'service', FALSE),
('Consulting', 'Interested in consulting services', '#6f42c1', 'service', FALSE),
('Custom Software', 'Interested in custom software development', '#dc3545', 'service', FALSE),

-- Behavioral Tags
('Quick Decision Maker', 'Makes decisions quickly', '#28a745', 'behavior', FALSE),
('Budget Conscious', 'Very price-sensitive', '#ffc107', 'behavior', FALSE),
('Quality Focused', 'Prioritizes quality over price', '#6f42c1', 'behavior', FALSE),
('Comparison Shopper', 'Compares multiple vendors', '#17a2b8', 'behavior', FALSE),

-- Campaign Tags
('Email Campaign', 'From email marketing campaigns', '#007bff', 'campaign', TRUE),
('Social Media', 'From social media campaigns', '#e83e8c', 'campaign', TRUE),
('Google Ads', 'From Google advertising', '#28a745', 'campaign', TRUE),
('Referral Program', 'From referral campaigns', '#fd7e14', 'campaign', TRUE);