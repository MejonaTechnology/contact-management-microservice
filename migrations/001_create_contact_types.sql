-- Create contact_types table for categorizing different types of contacts
CREATE TABLE IF NOT EXISTS contact_types (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    color VARCHAR(7) DEFAULT '#007bff', -- Hex color for UI display
    icon VARCHAR(50) DEFAULT 'contact', -- Icon identifier for UI
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_active (is_active),
    INDEX idx_sort_order (sort_order)
);

-- Insert default contact types
INSERT INTO contact_types (name, description, color, icon, sort_order) VALUES
('General Inquiry', 'General questions and information requests', '#007bff', 'info', 1),
('Sales Inquiry', 'Product/service sales related inquiries', '#28a745', 'shopping-cart', 2),
('Support Request', 'Technical support and help requests', '#dc3545', 'life-ring', 3),
('Partnership', 'Business partnership and collaboration inquiries', '#fd7e14', 'handshake', 4),
('Consultation', 'Professional consultation requests', '#6f42c1', 'user-md', 5),
('Job Application', 'Career and employment related contacts', '#17a2b8', 'briefcase', 6),
('Media Inquiry', 'Press and media related contacts', '#e83e8c', 'newspaper', 7),
('Complaint', 'Customer complaints and issues', '#dc3545', 'exclamation-triangle', 8),
('Feedback', 'General feedback and suggestions', '#20c997', 'comment', 9),
('Other', 'Other types of inquiries', '#6c757d', 'question', 10);