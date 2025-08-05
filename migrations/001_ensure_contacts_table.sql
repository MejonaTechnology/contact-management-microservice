-- Ensure contacts table exists with proper structure for CRM
-- This migration creates the table if it doesn't exist or adds missing columns

-- Create contacts table if it doesn't exist
CREATE TABLE IF NOT EXISTS `contacts` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  
  -- Basic Information
  `first_name` varchar(100) NOT NULL,
  `last_name` varchar(100) DEFAULT NULL,
  `email` varchar(255) NOT NULL,
  `phone` varchar(20) DEFAULT NULL,
  `company` varchar(200) DEFAULT NULL,
  `job_title` varchar(100) DEFAULT NULL,
  `website` varchar(255) DEFAULT NULL,
  
  -- Address Information
  `address_line1` varchar(255) DEFAULT NULL,
  `address_line2` varchar(255) DEFAULT NULL,
  `city` varchar(100) DEFAULT NULL,
  `state` varchar(100) DEFAULT NULL,
  `postal_code` varchar(20) DEFAULT NULL,
  `country` varchar(100) DEFAULT 'India',
  
  -- Contact Details
  `contact_type_id` bigint(20) UNSIGNED NOT NULL DEFAULT 1,
  `contact_source_id` bigint(20) UNSIGNED NOT NULL DEFAULT 1,
  `subject` varchar(500) DEFAULT NULL,
  `message` text DEFAULT NULL,
  `preferred_contact_method` enum('email','phone','sms','whatsapp') DEFAULT 'email',
  
  -- Lead Management
  `status` enum('new','contacted','qualified','proposal','negotiation','closed_won','closed_lost','on_hold','nurturing') DEFAULT 'new',
  `priority` enum('low','medium','high','urgent') DEFAULT 'medium',
  `lead_score` int(11) DEFAULT 0,
  `estimated_value` decimal(12,2) DEFAULT 0.00,
  `probability` int(11) DEFAULT 0,
  
  -- Assignment and Ownership
  `assigned_to` bigint(20) UNSIGNED DEFAULT NULL,
  `assigned_at` timestamp NULL DEFAULT NULL,
  `assigned_by` bigint(20) UNSIGNED DEFAULT NULL,
  
  -- Communication Tracking
  `last_contact_date` timestamp NULL DEFAULT NULL,
  `next_followup_date` timestamp NULL DEFAULT NULL,
  `response_time_hours` int(11) DEFAULT 0,
  `total_interactions` int(11) DEFAULT 0,
  `email_opened` tinyint(1) DEFAULT 0,
  `email_clicked` tinyint(1) DEFAULT 0,
  
  -- Lifecycle Tracking
  `first_contact_date` timestamp DEFAULT CURRENT_TIMESTAMP,
  `last_activity_date` timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `conversion_date` timestamp NULL DEFAULT NULL,
  `closed_date` timestamp NULL DEFAULT NULL,
  
  -- Technical Fields
  `ip_address` varchar(45) DEFAULT NULL,
  `user_agent` text DEFAULT NULL,
  `referrer_url` varchar(500) DEFAULT NULL,
  `landing_page` varchar(500) DEFAULT NULL,
  `utm_source` varchar(100) DEFAULT NULL,
  `utm_medium` varchar(100) DEFAULT NULL,
  `utm_campaign` varchar(100) DEFAULT NULL,
  `utm_term` varchar(100) DEFAULT NULL,
  `utm_content` varchar(100) DEFAULT NULL,
  
  -- Data Management
  `is_verified` tinyint(1) DEFAULT 0,
  `is_duplicate` tinyint(1) DEFAULT 0,
  `original_contact_id` bigint(20) UNSIGNED DEFAULT NULL,
  `data_source` varchar(100) DEFAULT 'form',
  
  -- Privacy and Compliance
  `marketing_consent` tinyint(1) DEFAULT 0,
  `data_processing_consent` tinyint(1) DEFAULT 1,
  `gdpr_consent` tinyint(1) DEFAULT 0,
  `unsubscribed` tinyint(1) DEFAULT 0,
  `do_not_call` tinyint(1) DEFAULT 0,
  
  -- Metadata
  `tags` json DEFAULT NULL,
  `custom_fields` json DEFAULT NULL,
  `notes` text DEFAULT NULL,
  
  -- Audit Fields
  `created_at` timestamp DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `created_by` bigint(20) UNSIGNED DEFAULT NULL,
  `updated_by` bigint(20) UNSIGNED DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_email_active` (`email`, `deleted_at`),
  KEY `idx_contacts_email` (`email`),
  KEY `idx_contacts_phone` (`phone`),
  KEY `idx_contacts_status` (`status`),
  KEY `idx_contacts_priority` (`priority`),
  KEY `idx_contacts_lead_score` (`lead_score`),
  KEY `idx_contacts_assigned_to` (`assigned_to`),
  KEY `idx_contacts_next_followup` (`next_followup_date`),
  KEY `idx_contacts_first_contact` (`first_contact_date`),
  KEY `idx_contacts_last_activity` (`last_activity_date`),
  KEY `idx_contacts_utm_source` (`utm_source`),
  KEY `idx_contacts_is_duplicate` (`is_duplicate`),
  KEY `idx_contacts_created_at` (`created_at`),
  KEY `idx_contacts_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create contact_types table if it doesn't exist
CREATE TABLE IF NOT EXISTS `contact_types` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL,
  `description` text DEFAULT NULL,
  `color` varchar(7) DEFAULT '#3B82F6',
  `icon` varchar(50) DEFAULT 'user',
  `is_active` tinyint(1) DEFAULT 1,
  `sort_order` int(11) DEFAULT 0,
  `created_at` timestamp DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_contact_type_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create contact_sources table if it doesn't exist
CREATE TABLE IF NOT EXISTS `contact_sources` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL,
  `description` text DEFAULT NULL,
  `utm_source` varchar(100) DEFAULT NULL,
  `utm_medium` varchar(100) DEFAULT NULL,
  `utm_campaign` varchar(100) DEFAULT NULL,
  `conversion_rate` decimal(5,2) DEFAULT 0.00,
  `cost_per_lead` decimal(10,2) DEFAULT 0.00,
  `is_active` tinyint(1) DEFAULT 1,
  `sort_order` int(11) DEFAULT 0,
  `created_at` timestamp DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_contact_source_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default contact types
INSERT IGNORE INTO `contact_types` (`id`, `name`, `description`, `color`, `icon`) VALUES
(1, 'General Inquiry', 'General inquiries and questions', '#3B82F6', 'mail'),
(2, 'Sales Lead', 'Potential sales opportunities', '#10B981', 'trending-up'),
(3, 'Support Request', 'Technical support requests', '#F59E0B', 'help-circle'),
(4, 'Partnership', 'Business partnership inquiries', '#8B5CF6', 'handshake'),
(5, 'Career Inquiry', 'Job applications and career questions', '#EF4444', 'briefcase');

-- Insert default contact sources
INSERT IGNORE INTO `contact_sources` (`id`, `name`, `description`, `utm_source`) VALUES
(1, 'Website Contact Form', 'Main contact form on website', 'website'),
(2, 'Google Ads', 'Google advertising campaigns', 'google'),
(3, 'Social Media', 'Social media platforms', 'social'),
(4, 'Email Campaign', 'Email marketing campaigns', 'email'),
(5, 'Referral', 'Word of mouth referrals', 'referral'),
(6, 'Direct', 'Direct website visits', 'direct');