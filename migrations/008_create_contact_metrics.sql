-- Create contact_metrics table for analytics and reporting
CREATE TABLE IF NOT EXISTS contact_metrics (
    id INT PRIMARY KEY AUTO_INCREMENT,
    
    -- Time Period
    metric_date DATE NOT NULL,
    metric_period ENUM('daily', 'weekly', 'monthly', 'quarterly', 'yearly') NOT NULL,
    
    -- Contact Volume Metrics
    total_contacts INT DEFAULT 0,
    new_contacts INT DEFAULT 0,
    updated_contacts INT DEFAULT 0,
    deleted_contacts INT DEFAULT 0,
    
    -- Status Distribution
    contacts_new INT DEFAULT 0,
    contacts_contacted INT DEFAULT 0,
    contacts_qualified INT DEFAULT 0,
    contacts_proposal INT DEFAULT 0,
    contacts_negotiation INT DEFAULT 0,
    contacts_closed_won INT DEFAULT 0,
    contacts_closed_lost INT DEFAULT 0,
    contacts_on_hold INT DEFAULT 0,
    contacts_nurturing INT DEFAULT 0,
    
    -- Source Distribution
    source_website INT DEFAULT 0,
    source_referral INT DEFAULT 0,
    source_social_media INT DEFAULT 0,
    source_email_marketing INT DEFAULT 0,
    source_paid_ads INT DEFAULT 0,
    source_phone INT DEFAULT 0,
    source_other INT DEFAULT 0,
    
    -- Communication Metrics
    emails_sent INT DEFAULT 0,
    emails_opened INT DEFAULT 0,
    emails_clicked INT DEFAULT 0,
    emails_replied INT DEFAULT 0,
    sms_sent INT DEFAULT 0,
    sms_delivered INT DEFAULT 0,
    calls_made INT DEFAULT 0,
    calls_answered INT DEFAULT 0,
    
    -- Response Time Metrics
    avg_first_response_hours DECIMAL(8,2) DEFAULT 0.00,
    avg_resolution_hours DECIMAL(8,2) DEFAULT 0.00,
    contacts_responded_24h INT DEFAULT 0,
    contacts_responded_48h INT DEFAULT 0,
    contacts_overdue INT DEFAULT 0,
    
    -- Conversion Metrics
    leads_qualified INT DEFAULT 0,
    leads_converted INT DEFAULT 0,
    conversion_rate DECIMAL(5,2) DEFAULT 0.00,
    qualification_rate DECIMAL(5,2) DEFAULT 0.00,
    
    -- Value Metrics
    total_estimated_value DECIMAL(15,2) DEFAULT 0.00,
    total_actual_value DECIMAL(15,2) DEFAULT 0.00,
    avg_contact_value DECIMAL(12,2) DEFAULT 0.00,
    revenue_generated DECIMAL(15,2) DEFAULT 0.00,
    
    -- Activity Metrics
    total_activities INT DEFAULT 0,
    meetings_scheduled INT DEFAULT 0,
    meetings_held INT DEFAULT 0,
    meetings_completed INT DEFAULT 0,
    proposals_sent INT DEFAULT 0,
    
    -- Team Performance
    total_assigned_contacts INT DEFAULT 0,
    contacts_per_user DECIMAL(8,2) DEFAULT 0.00,
    active_users INT DEFAULT 0,
    
    -- Lead Scoring
    avg_lead_score DECIMAL(5,2) DEFAULT 0.00,
    high_score_contacts INT DEFAULT 0, -- Score > 70
    medium_score_contacts INT DEFAULT 0, -- Score 40-70
    low_score_contacts INT DEFAULT 0, -- Score < 40
    
    -- Geographic Distribution
    contacts_local INT DEFAULT 0,
    contacts_national INT DEFAULT 0,
    contacts_international INT DEFAULT 0,
    
    -- Industry Distribution
    contacts_technology INT DEFAULT 0,
    contacts_healthcare INT DEFAULT 0,
    contacts_finance INT DEFAULT 0,
    contacts_education INT DEFAULT 0,
    contacts_retail INT DEFAULT 0,
    contacts_other_industry INT DEFAULT 0,
    
    -- Audit Fields
    calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Indexes
    UNIQUE KEY unique_metric_date_period (metric_date, metric_period),
    INDEX idx_metric_date (metric_date),
    INDEX idx_metric_period (metric_period),
    INDEX idx_calculated_at (calculated_at)
);

-- Create contact_source_metrics for detailed source analysis
CREATE TABLE IF NOT EXISTS contact_source_metrics (
    id INT PRIMARY KEY AUTO_INCREMENT,
    contact_source_id INT NOT NULL,
    
    -- Time Period
    metric_date DATE NOT NULL,
    metric_period ENUM('daily', 'weekly', 'monthly', 'quarterly', 'yearly') NOT NULL,
    
    -- Volume Metrics
    total_contacts INT DEFAULT 0,
    new_contacts INT DEFAULT 0,
    qualified_contacts INT DEFAULT 0,
    converted_contacts INT DEFAULT 0,
    
    -- Performance Metrics
    conversion_rate DECIMAL(5,2) DEFAULT 0.00,
    qualification_rate DECIMAL(5,2) DEFAULT 0.00,
    avg_lead_score DECIMAL(5,2) DEFAULT 0.00,
    avg_contact_value DECIMAL(12,2) DEFAULT 0.00,
    
    -- Cost and ROI
    marketing_spend DECIMAL(12,2) DEFAULT 0.00,
    cost_per_contact DECIMAL(10,2) DEFAULT 0.00,
    cost_per_conversion DECIMAL(10,2) DEFAULT 0.00,
    roi_percentage DECIMAL(8,2) DEFAULT 0.00,
    
    -- Response Metrics
    avg_response_time_hours DECIMAL(8,2) DEFAULT 0.00,
    contacts_responded_24h INT DEFAULT 0,
    
    -- Revenue Metrics
    total_revenue DECIMAL(15,2) DEFAULT 0.00,
    avg_revenue_per_contact DECIMAL(12,2) DEFAULT 0.00,
    
    calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY unique_source_date_period (contact_source_id, metric_date, metric_period),
    INDEX idx_contact_source_id (contact_source_id),
    INDEX idx_metric_date (metric_date),
    INDEX idx_conversion_rate (conversion_rate),
    INDEX idx_roi_percentage (roi_percentage),
    
    FOREIGN KEY (contact_source_id) REFERENCES contact_sources(id) ON DELETE CASCADE
);

-- Create user_performance_metrics for team performance tracking
CREATE TABLE IF NOT EXISTS user_performance_metrics (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    
    -- Time Period
    metric_date DATE NOT NULL,
    metric_period ENUM('daily', 'weekly', 'monthly', 'quarterly', 'yearly') NOT NULL,
    
    -- Contact Management
    assigned_contacts INT DEFAULT 0,
    new_contacts_handled INT DEFAULT 0,
    contacts_qualified INT DEFAULT 0,
    contacts_converted INT DEFAULT 0,
    
    -- Activity Metrics
    total_activities INT DEFAULT 0,
    emails_sent INT DEFAULT 0,
    calls_made INT DEFAULT 0,
    meetings_held INT DEFAULT 0,
    
    -- Performance Metrics
    avg_response_time_hours DECIMAL(8,2) DEFAULT 0.00,
    conversion_rate DECIMAL(5,2) DEFAULT 0.00,
    qualification_rate DECIMAL(5,2) DEFAULT 0.00,
    
    -- Revenue Metrics
    revenue_generated DECIMAL(15,2) DEFAULT 0.00,
    avg_deal_size DECIMAL(12,2) DEFAULT 0.00,
    
    -- Quality Metrics
    customer_satisfaction_score DECIMAL(3,2) DEFAULT 0.00,
    response_quality_score DECIMAL(3,2) DEFAULT 0.00,
    
    calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY unique_user_date_period (user_id, metric_date, metric_period),
    INDEX idx_user_id (user_id),
    INDEX idx_metric_date (metric_date),
    INDEX idx_conversion_rate (conversion_rate),
    INDEX idx_revenue_generated (revenue_generated)
);

-- Create contact_funnel_metrics for conversion funnel analysis
CREATE TABLE IF NOT EXISTS contact_funnel_metrics (
    id INT PRIMARY KEY AUTO_INCREMENT,
    
    -- Time Period
    metric_date DATE NOT NULL,
    metric_period ENUM('daily', 'weekly', 'monthly', 'quarterly', 'yearly') NOT NULL,
    
    -- Funnel Stages (Lead to Customer Journey)
    stage_1_new INT DEFAULT 0,
    stage_2_contacted INT DEFAULT 0,
    stage_3_qualified INT DEFAULT 0,
    stage_4_proposal INT DEFAULT 0,
    stage_5_negotiation INT DEFAULT 0,
    stage_6_closed_won INT DEFAULT 0,
    stage_7_closed_lost INT DEFAULT 0,
    
    -- Conversion Rates Between Stages
    new_to_contacted_rate DECIMAL(5,2) DEFAULT 0.00,
    contacted_to_qualified_rate DECIMAL(5,2) DEFAULT 0.00,
    qualified_to_proposal_rate DECIMAL(5,2) DEFAULT 0.00,
    proposal_to_negotiation_rate DECIMAL(5,2) DEFAULT 0.00,
    negotiation_to_won_rate DECIMAL(5,2) DEFAULT 0.00,
    
    -- Overall Funnel Performance
    overall_conversion_rate DECIMAL(5,2) DEFAULT 0.00,
    avg_days_to_conversion DECIMAL(8,2) DEFAULT 0.00,
    
    -- Drop-off Analysis
    dropped_after_contacted INT DEFAULT 0,
    dropped_after_qualified INT DEFAULT 0,
    dropped_after_proposal INT DEFAULT 0,
    dropped_after_negotiation INT DEFAULT 0,
    
    calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY unique_funnel_date_period (metric_date, metric_period),
    INDEX idx_metric_date (metric_date),
    INDEX idx_metric_period (metric_period),
    INDEX idx_overall_conversion_rate (overall_conversion_rate)
);