-- Create simplified contact_submissions table for dashboard compatibility
CREATE TABLE IF NOT EXISTS contact_submissions (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(200) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    subject VARCHAR(500),
    message TEXT NOT NULL,
    source VARCHAR(100),
    status VARCHAR(50) DEFAULT 'new',
    assigned_to INT,
    response_sent BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Indexes for performance
    INDEX idx_email (email),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    INDEX idx_assigned_to (assigned_to)
);

-- Insert sample data for testing
INSERT INTO contact_submissions (name, email, phone, subject, message, source, status) VALUES
('John Doe', 'john.doe@example.com', '+91-9876543210', 'Website Development Inquiry', 'I am interested in getting a professional website developed for my business.', 'website', 'new'),
('Jane Smith', 'jane.smith@example.com', '+91-9876543211', 'Mobile App Development', 'Looking for a mobile app development service for my startup.', 'referral', 'in_progress'),
('Mike Johnson', 'mike.johnson@example.com', '+91-9876543212', 'Digital Marketing Services', 'Need help with SEO and digital marketing for my online store.', 'google', 'resolved'),
('Sarah Wilson', 'sarah.wilson@example.com', '+91-9876543213', 'Partnership Opportunity', 'Interested in exploring partnership opportunities with Mejona Technology.', 'linkedin', 'new'),
('David Brown', 'david.brown@example.com', '+91-9876543214', 'Technical Support', 'Need technical support for the application you developed for us.', 'website', 'in_progress'),
('Emily Davis', 'emily.davis@example.com', '+91-9876543215', 'Consultation Request', 'Would like to schedule a consultation for AI integration in our business.', 'website', 'new');
