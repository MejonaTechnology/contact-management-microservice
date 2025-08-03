package services

import (
	"contact-service/internal/models"
	"contact-service/pkg/database"
	"contact-service/pkg/logger"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// ContactTypeService handles business logic for contact type management
type ContactTypeService struct {
	db *gorm.DB
}

// NewContactTypeService creates a new contact type service instance
func NewContactTypeService() *ContactTypeService {
	return &ContactTypeService{
		db: database.DB,
	}
}

// CreateContactType creates a new contact type
func (s *ContactTypeService) CreateContactType(req *models.ContactTypeRequest, createdBy *uint) (*models.ContactType, error) {
	// Check if name already exists
	var existing models.ContactType
	if err := s.db.Where("name = ?", req.Name).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("contact type with name '%s' already exists", req.Name)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing contact type: %v", err)
	}

	// Create contact type
	contactType := &models.ContactType{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
		IsActive:    true,
		SortOrder:   0,
	}

	// Set optional fields
	if req.IsActive != nil {
		contactType.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		contactType.SortOrder = *req.SortOrder
	}

	// Set defaults
	if contactType.Color == "" {
		contactType.Color = "#007bff"
	}
	if contactType.Icon == "" {
		contactType.Icon = "contact"
	}

	// Save to database
	if err := s.db.Create(contactType).Error; err != nil {
		logger.Error("Failed to create contact type", err, map[string]interface{}{
			"name": req.Name,
		})
		return nil, fmt.Errorf("failed to create contact type: %v", err)
	}

	logger.LogBusinessEvent("contact_type_created", "contact_type", contactType.ID, map[string]interface{}{
		"name":       contactType.Name,
		"created_by": createdBy,
	})

	return contactType, nil
}

// GetContactType retrieves a contact type by ID
func (s *ContactTypeService) GetContactType(id uint) (*models.ContactType, error) {
	var contactType models.ContactType
	
	if err := s.db.First(&contactType, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("contact type not found")
		}
		return nil, fmt.Errorf("failed to get contact type: %v", err)
	}

	return &contactType, nil
}

// UpdateContactType updates an existing contact type
func (s *ContactTypeService) UpdateContactType(id uint, req *models.ContactTypeRequest, updatedBy *uint) (*models.ContactType, error) {
	// Get existing contact type
	contactType, err := s.GetContactType(id)
	if err != nil {
		return nil, err
	}

	// Check if name already exists (excluding current record)
	if req.Name != contactType.Name {
		var existing models.ContactType
		if err := s.db.Where("name = ? AND id != ?", req.Name, id).First(&existing).Error; err == nil {
			return nil, fmt.Errorf("contact type with name '%s' already exists", req.Name)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to check existing contact type: %v", err)
		}
	}

	// Store original values
	originalName := contactType.Name
	originalActive := contactType.IsActive

	// Update fields
	contactType.Name = req.Name
	contactType.Description = req.Description
	contactType.Color = req.Color
	contactType.Icon = req.Icon

	// Set optional fields
	if req.IsActive != nil {
		contactType.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		contactType.SortOrder = *req.SortOrder
	}

	// Set defaults
	if contactType.Color == "" {
		contactType.Color = "#007bff"
	}
	if contactType.Icon == "" {
		contactType.Icon = "contact"
	}

	// Save to database
	if err := s.db.Save(contactType).Error; err != nil {
		logger.Error("Failed to update contact type", err, map[string]interface{}{
			"id":   id,
			"name": req.Name,
		})
		return nil, fmt.Errorf("failed to update contact type: %v", err)
	}

	// Log significant changes
	changes := make(map[string]interface{})
	if originalName != contactType.Name {
		changes["name_changed"] = map[string]string{"from": originalName, "to": contactType.Name}
	}
	if originalActive != contactType.IsActive {
		changes["status_changed"] = map[string]bool{"from": originalActive, "to": contactType.IsActive}
	}
	changes["updated_by"] = updatedBy

	logger.LogBusinessEvent("contact_type_updated", "contact_type", contactType.ID, changes)

	return contactType, nil
}

// DeleteContactType soft deletes a contact type
func (s *ContactTypeService) DeleteContactType(id uint, deletedBy *uint) error {
	// Check if contact type exists
	contactType, err := s.GetContactType(id)
	if err != nil {
		return err
	}

	// Check if contact type is in use
	var contactCount int64
	if err := s.db.Model(&models.Contact{}).
		Where("contact_type_id = ? AND deleted_at IS NULL", id).
		Count(&contactCount).Error; err != nil {
		return fmt.Errorf("failed to check contact usage: %v", err)
	}

	if contactCount > 0 {
		return fmt.Errorf("cannot delete contact type: %d contacts are using this type", contactCount)
	}

	// Soft delete by setting is_active to false
	contactType.IsActive = false
	if err := s.db.Save(contactType).Error; err != nil {
		logger.Error("Failed to delete contact type", err, map[string]interface{}{
			"id": id,
		})
		return fmt.Errorf("failed to delete contact type: %v", err)
	}

	logger.LogBusinessEvent("contact_type_deleted", "contact_type", contactType.ID, map[string]interface{}{
		"name":       contactType.Name,
		"deleted_by": deletedBy,
	})

	return nil
}

// ListContactTypes retrieves all contact types
func (s *ContactTypeService) ListContactTypes(activeOnly bool) ([]*models.ContactType, error) {
	query := s.db.Model(&models.ContactType{})
	
	if activeOnly {
		query = query.Where("is_active = true")
	}
	
	var contactTypes []*models.ContactType
	if err := query.Order("sort_order ASC, name ASC").Find(&contactTypes).Error; err != nil {
		return nil, fmt.Errorf("failed to list contact types: %v", err)
	}

	return contactTypes, nil
}

// GetContactTypeUsageStats returns usage statistics for contact types
func (s *ContactTypeService) GetContactTypeUsageStats() (map[uint]int64, error) {
	var results []struct {
		ContactTypeID uint  `json:"contact_type_id"`
		Count         int64 `json:"count"`
	}

	if err := s.db.Model(&models.Contact{}).
		Select("contact_type_id, COUNT(*) as count").
		Where("deleted_at IS NULL").
		Group("contact_type_id").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get usage stats: %v", err)
	}

	stats := make(map[uint]int64)
	for _, result := range results {
		stats[result.ContactTypeID] = result.Count
	}

	return stats, nil
}

// ReorderContactTypes updates the sort order of contact types
func (s *ContactTypeService) ReorderContactTypes(orderMap map[uint]int, updatedBy *uint) error {
	return database.Transaction(func(tx *gorm.DB) error {
		for id, sortOrder := range orderMap {
			if err := tx.Model(&models.ContactType{}).
				Where("id = ?", id).
				Update("sort_order", sortOrder).Error; err != nil {
				return fmt.Errorf("failed to update sort order for contact type %d: %v", id, err)
			}
		}

		logger.LogBusinessEvent("contact_types_reordered", "contact_type", nil, map[string]interface{}{
			"order_map":  orderMap,
			"updated_by": updatedBy,
		})

		return nil
	})
}