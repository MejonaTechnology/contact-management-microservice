-- Create contact_sources table for tracking where contacts come from
CREATE TABLE IF NOT EXISTS contact_sources (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    utm_source VARCHAR(100), -- UTM parameter mapping
    utm_medium VARCHAR(100), -- UTM parameter mapping
    utm_campaign VARCHAR(100), -- UTM parameter mapping
    conversion_rate DECIMAL(5,2) DEFAULT 0.00, -- Track conversion performance
    cost_per_lead DECIMAL(10,2) DEFAULT 0.00, -- Marketing cost tracking
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_active (is_active),
    INDEX idx_utm_source (utm_source),
    INDEX idx_conversion_rate (conversion_rate)
);

-- Insert default contact sources
INSERT INTO contact_sources (name, description, utm_source, utm_medium, sort_order) VALUES
('Website Contact Form', 'Main website contact form submissions', 'website', 'form', 1),
('Homepage Quick Contact', 'Quick contact widget on homepage', 'homepage', 'widget', 2),
('Google Ads', 'Google advertising campaigns', 'google', 'cpc', 3),
('Facebook Ads', 'Facebook and Instagram advertising', 'facebook', 'social-paid', 4),
('LinkedIn', 'LinkedIn organic and paid traffic', 'linkedin', 'social', 5),
('Email Marketing', 'Email campaign responses', 'email', 'newsletter', 6),
('Referral', 'Word-of-mouth and referral traffic', 'referral', 'word-of-mouth', 7),
('Direct Traffic', 'Direct website visits', 'direct', 'none', 8),
('SEO Organic', 'Organic search engine traffic', 'google', 'organic', 9),
('Social Media', 'Social media organic traffic', 'social', 'organic', 10),
('Phone Call', 'Direct phone inquiries', 'phone', 'call', 11),
('Walk-in', 'Physical office visits', 'office', 'walk-in', 12),
('Trade Show', 'Conference and trade show leads', 'event', 'trade-show', 13),
('Partnership', 'Partner referrals', 'partner', 'referral', 14),
('Other', 'Other sources not listed', 'other', 'unknown', 15);