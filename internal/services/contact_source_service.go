package services

import (
	"contact-service/internal/models"
	"contact-service/pkg/database"
	"contact-service/pkg/logger"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// ContactSourceService handles business logic for contact source management
type ContactSourceService struct {
	db *gorm.DB
}

// NewContactSourceService creates a new contact source service instance
func NewContactSourceService() *ContactSourceService {
	return &ContactSourceService{
		db: database.DB,
	}
}

// CreateContactSource creates a new contact source
func (s *ContactSourceService) CreateContactSource(req *models.ContactSourceRequest, createdBy *uint) (*models.ContactSource, error) {
	// Check if name already exists
	var existing models.ContactSource
	if err := s.db.Where("name = ?", req.Name).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("contact source with name '%s' already exists", req.Name)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing contact source: %v", err)
	}

	// Create contact source
	contactSource := &models.ContactSource{
		Name:           req.Name,
		Description:    req.Description,
		UTMSource:      req.UTMSource,
		UTMMedium:      req.UTMMedium,
		UTMCampaign:    req.UTMCampaign,
		ConversionRate: 0.0,
		CostPerLead:    0.0,
		IsActive:       true,
		SortOrder:      0,
	}

	// Set optional fields
	if req.ConversionRate != nil {
		contactSource.ConversionRate = *req.ConversionRate
	}
	if req.CostPerLead != nil {
		contactSource.CostPerLead = *req.CostPerLead
	}
	if req.IsActive != nil {
		contactSource.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		contactSource.SortOrder = *req.SortOrder
	}

	// Save to database
	if err := s.db.Create(contactSource).Error; err != nil {
		logger.Error("Failed to create contact source", err, map[string]interface{}{
			"name": req.Name,
		})
		return nil, fmt.Errorf("failed to create contact source: %v", err)
	}

	logger.LogBusinessEvent("contact_source_created", "contact_source", contactSource.ID, map[string]interface{}{
		"name":       contactSource.Name,
		"created_by": createdBy,
	})

	return contactSource, nil
}

// GetContactSource retrieves a contact source by ID
func (s *ContactSourceService) GetContactSource(id uint) (*models.ContactSource, error) {
	var contactSource models.ContactSource
	
	if err := s.db.First(&contactSource, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("contact source not found")
		}
		return nil, fmt.Errorf("failed to get contact source: %v", err)
	}

	return &contactSource, nil
}

// UpdateContactSource updates an existing contact source
func (s *ContactSourceService) UpdateContactSource(id uint, req *models.ContactSourceRequest, updatedBy *uint) (*models.ContactSource, error) {
	// Get existing contact source
	contactSource, err := s.GetContactSource(id)
	if err != nil {
		return nil, err
	}

	// Check if name already exists (excluding current record)
	if req.Name != contactSource.Name {
		var existing models.ContactSource
		if err := s.db.Where("name = ? AND id != ?", req.Name, id).First(&existing).Error; err == nil {
			return nil, fmt.Errorf("contact source with name '%s' already exists", req.Name)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to check existing contact source: %v", err)
		}
	}

	// Store original values
	originalName := contactSource.Name
	originalActive := contactSource.IsActive

	// Update fields
	contactSource.Name = req.Name
	contactSource.Description = req.Description
	contactSource.UTMSource = req.UTMSource
	contactSource.UTMMedium = req.UTMMedium
	contactSource.UTMCampaign = req.UTMCampaign

	// Set optional fields
	if req.ConversionRate != nil {
		contactSource.ConversionRate = *req.ConversionRate
	}
	if req.CostPerLead != nil {
		contactSource.CostPerLead = *req.CostPerLead
	}
	if req.IsActive != nil {
		contactSource.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		contactSource.SortOrder = *req.SortOrder
	}

	// Save to database
	if err := s.db.Save(contactSource).Error; err != nil {
		logger.Error("Failed to update contact source", err, map[string]interface{}{
			"id":   id,
			"name": req.Name,
		})
		return nil, fmt.Errorf("failed to update contact source: %v", err)
	}

	// Log significant changes
	changes := make(map[string]interface{})
	if originalName != contactSource.Name {
		changes["name_changed"] = map[string]string{"from": originalName, "to": contactSource.Name}
	}
	if originalActive != contactSource.IsActive {
		changes["status_changed"] = map[string]bool{"from": originalActive, "to": contactSource.IsActive}
	}
	changes["updated_by"] = updatedBy

	logger.LogBusinessEvent("contact_source_updated", "contact_source", contactSource.ID, changes)

	return contactSource, nil
}

// DeleteContactSource soft deletes a contact source
func (s *ContactSourceService) DeleteContactSource(id uint, deletedBy *uint) error {
	// Check if contact source exists
	contactSource, err := s.GetContactSource(id)
	if err != nil {
		return err
	}

	// Check if contact source is in use
	var contactCount int64
	if err := s.db.Model(&models.Contact{}).
		Where("contact_source_id = ? AND deleted_at IS NULL", id).
		Count(&contactCount).Error; err != nil {
		return fmt.Errorf("failed to check contact usage: %v", err)
	}

	if contactCount > 0 {
		return fmt.Errorf("cannot delete contact source: %d contacts are using this source", contactCount)
	}

	// Soft delete by setting is_active to false
	contactSource.IsActive = false
	if err := s.db.Save(contactSource).Error; err != nil {
		logger.Error("Failed to delete contact source", err, map[string]interface{}{
			"id": id,
		})
		return fmt.Errorf("failed to delete contact source: %v", err)
	}

	logger.LogBusinessEvent("contact_source_deleted", "contact_source", contactSource.ID, map[string]interface{}{
		"name":       contactSource.Name,
		"deleted_by": deletedBy,
	})

	return nil
}

// ListContactSources retrieves all contact sources
func (s *ContactSourceService) ListContactSources(activeOnly bool) ([]*models.ContactSource, error) {
	query := s.db.Model(&models.ContactSource{})
	
	if activeOnly {
		query = query.Where("is_active = true")
	}
	
	var contactSources []*models.ContactSource
	if err := query.Order("sort_order ASC, name ASC").Find(&contactSources).Error; err != nil {
		return nil, fmt.Errorf("failed to list contact sources: %v", err)
	}

	return contactSources, nil
}

// GetContactSourceUsageStats returns usage statistics for contact sources
func (s *ContactSourceService) GetContactSourceUsageStats() (map[uint]int64, error) {
	var results []struct {
		ContactSourceID uint  `json:"contact_source_id"`
		Count           int64 `json:"count"`
	}

	if err := s.db.Model(&models.Contact{}).
		Select("contact_source_id, COUNT(*) as count").
		Where("deleted_at IS NULL").
		Group("contact_source_id").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get usage stats: %v", err)
	}

	stats := make(map[uint]int64)
	for _, result := range results {
		stats[result.ContactSourceID] = result.Count
	}

	return stats, nil
}

// GetContactSourcePerformanceMetrics returns performance metrics for contact sources
func (s *ContactSourceService) GetContactSourcePerformanceMetrics() (map[uint]map[string]interface{}, error) {
	var results []struct {
		ContactSourceID   uint    `json:"contact_source_id"`
		TotalContacts     int64   `json:"total_contacts"`
		QualifiedContacts int64   `json:"qualified_contacts"`
		ConvertedContacts int64   `json:"converted_contacts"`
		AvgLeadScore      float64 `json:"avg_lead_score"`
		TotalValue        float64 `json:"total_value"`
	}

	query := `
		SELECT 
			contact_source_id,
			COUNT(*) as total_contacts,
			COUNT(CASE WHEN status IN ('qualified', 'proposal', 'negotiation', 'closed_won') THEN 1 END) as qualified_contacts,
			COUNT(CASE WHEN status = 'closed_won' THEN 1 END) as converted_contacts,
			AVG(lead_score) as avg_lead_score,
			SUM(CASE WHEN status = 'closed_won' THEN estimated_value ELSE 0 END) as total_value
		FROM contacts 
		WHERE deleted_at IS NULL 
		GROUP BY contact_source_id
	`

	if err := s.db.Raw(query).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get performance metrics: %v", err)
	}

	metrics := make(map[uint]map[string]interface{})
	for _, result := range results {
		conversionRate := 0.0
		if result.TotalContacts > 0 {
			conversionRate = (float64(result.ConvertedContacts) / float64(result.TotalContacts)) * 100
		}

		qualificationRate := 0.0
		if result.TotalContacts > 0 {
			qualificationRate = (float64(result.QualifiedContacts) / float64(result.TotalContacts)) * 100
		}

		avgValue := 0.0
		if result.ConvertedContacts > 0 {
			avgValue = result.TotalValue / float64(result.ConvertedContacts)
		}

		metrics[result.ContactSourceID] = map[string]interface{}{
			"total_contacts":      result.TotalContacts,
			"qualified_contacts":  result.QualifiedContacts,
			"converted_contacts":  result.ConvertedContacts,
			"conversion_rate":     conversionRate,
			"qualification_rate":  qualificationRate,
			"avg_lead_score":      result.AvgLeadScore,
			"total_revenue":       result.TotalValue,
			"avg_deal_value":      avgValue,
		}
	}

	return metrics, nil
}

// UpdateContactSourceMetrics updates performance metrics for a contact source
func (s *ContactSourceService) UpdateContactSourceMetrics(id uint, conversionRate, costPerLead float64) error {
	contactSource, err := s.GetContactSource(id)
	if err != nil {
		return err
	}

	contactSource.ConversionRate = conversionRate
	contactSource.CostPerLead = costPerLead

	if err := s.db.Save(contactSource).Error; err != nil {
		return fmt.Errorf("failed to update contact source metrics: %v", err)
	}

	logger.LogBusinessEvent("contact_source_metrics_updated", "contact_source", contactSource.ID, map[string]interface{}{
		"conversion_rate": conversionRate,
		"cost_per_lead":   costPerLead,
	})

	return nil
}

// ReorderContactSources updates the sort order of contact sources
func (s *ContactSourceService) ReorderContactSources(orderMap map[uint]int, updatedBy *uint) error {
	return database.Transaction(func(tx *gorm.DB) error {
		for id, sortOrder := range orderMap {
			if err := tx.Model(&models.ContactSource{}).
				Where("id = ?", id).
				Update("sort_order", sortOrder).Error; err != nil {
				return fmt.Errorf("failed to update sort order for contact source %d: %v", id, err)
			}
		}

		logger.LogBusinessEvent("contact_sources_reordered", "contact_source", nil, map[string]interface{}{
			"order_map":  orderMap,
			"updated_by": updatedBy,
		})

		return nil
	})
}

// GetContactSourceByUTM finds a contact source by UTM parameters
func (s *ContactSourceService) GetContactSourceByUTM(utmSource, utmMedium, utmCampaign string) (*models.ContactSource, error) {
	var contactSource models.ContactSource
	
	query := s.db.Where("is_active = true")
	
	if utmSource != "" {
		query = query.Where("utm_source = ?", utmSource)
	}
	if utmMedium != "" {
		query = query.Where("utm_medium = ?", utmMedium)
	}
	if utmCampaign != "" {
		query = query.Where("utm_campaign = ?", utmCampaign)
	}
	
	if err := query.First(&contactSource).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No matching source found
		}
		return nil, fmt.Errorf("failed to find contact source by UTM: %v", err)
	}

	return &contactSource, nil
}