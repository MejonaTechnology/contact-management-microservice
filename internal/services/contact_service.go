package services

import (
	"contact-service/internal/models"
	"contact-service/pkg/database"
	"contact-service/pkg/logger"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ContactService handles business logic for contact management
type ContactService struct {
	db *gorm.DB
}

// NewContactService creates a new contact service instance
func NewContactService() *ContactService {
	return &ContactService{
		db: database.DB,
	}
}

// ContactListOptions represents options for listing contacts
type ContactListOptions struct {
	Page        int
	PageSize    int
	Status      string
	Priority    string
	AssignedTo  *uint
	SourceID    *uint
	TypeID      *uint
	Search      string
	Tags        []string
	SortBy      string
	SortOrder   string
	DateFrom    *time.Time
	DateTo      *time.Time
}

// AdvancedSearchCriteria represents advanced search criteria
type AdvancedSearchCriteria struct {
	Page                int
	PageSize            int
	SortBy              string
	SortOrder           string
	FullTextSearch      *string
	FirstName           *string
	LastName            *string
	Email               *string
	Phone               *string
	Company             *string
	JobTitle            *string
	Country             *string
	Status              *string
	Priority            *string
	LeadScoreMin        *int
	LeadScoreMax        *int
	EstimatedValueMin   *float64
	EstimatedValueMax   *float64
	AssignedTo          *uint
	SourceID            *uint
	TypeID              *uint
	Tags                []string
	CreatedFrom         *time.Time
	CreatedTo           *time.Time
	LastContactFrom     *time.Time
	LastContactTo       *time.Time
	HasActivities       *bool
	IsHotLead           *bool
	IsHighPriority      *bool
}

// CreateContact creates a new contact
func (s *ContactService) CreateContact(req *models.ContactRequest, createdBy *uint) (*models.Contact, error) {
	// Validate contact type and source exist
	if err := s.validateContactTypeAndSource(req.ContactTypeID, req.ContactSourceID); err != nil {
		return nil, err
	}

	// Check for duplicates if enabled
	if duplicate, err := s.checkForDuplicates(req.Email, req.Phone); err != nil {
		return nil, fmt.Errorf("duplicate check failed: %v", err)
	} else if duplicate != nil {
		return nil, fmt.Errorf("contact already exists with ID: %d", duplicate.ID)
	}

	// Create contact entity
	contact := &models.Contact{
		FirstName:             req.FirstName,
		LastName:              req.LastName,
		Email:                 req.Email,
		Phone:                 req.Phone,
		Company:               req.Company,
		JobTitle:              req.JobTitle,
		Website:               req.Website,
		AddressLine1:          req.AddressLine1,
		AddressLine2:          req.AddressLine2,
		City:                  req.City,
		State:                 req.State,
		PostalCode:            req.PostalCode,
		ContactTypeID:         req.ContactTypeID,
		ContactSourceID:       req.ContactSourceID,
		Subject:               req.Subject,
		Message:               req.Message,
		Status:                models.StatusNew,
		Priority:              models.PriorityMedium,
		LeadScore:             0,
		Country:               "India",
		MarketingConsent:      false,
		DataProcessingConsent: true,
		Tags:                  req.Tags,
		CustomFields:          req.CustomFields,
		Notes:                 req.Notes,
		CreatedBy:             createdBy,
	}

	// Set optional fields
	if req.Country != nil {
		contact.Country = *req.Country
	}
	if req.PreferredContactMethod != nil {
		contact.PreferredContactMethod = *req.PreferredContactMethod
	}
	if req.Priority != nil {
		contact.Priority = *req.Priority
	}
	if req.EstimatedValue != nil {
		contact.EstimatedValue = *req.EstimatedValue
	}
	if req.AssignedTo != nil {
		contact.AssignedTo = req.AssignedTo
		now := time.Now()
		contact.AssignedAt = &now
		contact.AssignedBy = createdBy
	}
	if req.NextFollowupDate != nil {
		contact.NextFollowupDate = req.NextFollowupDate
	}
	if req.MarketingConsent != nil {
		contact.MarketingConsent = *req.MarketingConsent
	}
	if req.DataProcessingConsent != nil {
		contact.DataProcessingConsent = *req.DataProcessingConsent
	}
	if req.GDPRConsent != nil {
		contact.GDPRConsent = *req.GDPRConsent
	}

	// Save to database
	if err := s.db.Create(contact).Error; err != nil {
		logger.Error("Failed to create contact", err, map[string]interface{}{
			"email": req.Email,
		})
		return nil, fmt.Errorf("failed to create contact: %v", err)
	}

	// Log activity
	s.logContactActivity(contact.ID, "contact_created", map[string]interface{}{
		"source":      "api",
		"created_by":  createdBy,
		"contact_type": req.ContactTypeID,
		"source_id":   req.ContactSourceID,
	})

	// Calculate initial lead score
	if err := s.calculateLeadScore(contact.ID); err != nil {
		logger.Warn("Failed to calculate initial lead score", map[string]interface{}{
			"contact_id": contact.ID,
			"error":      err.Error(),
		})
	}

	return contact, nil
}

// GetContact retrieves a contact by ID
func (s *ContactService) GetContact(id uint) (*models.Contact, error) {
	var contact models.Contact
	
	if err := s.db.Preload("ContactType").Preload("ContactSource").
		First(&contact, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("contact not found")
		}
		return nil, fmt.Errorf("failed to get contact: %v", err)
	}

	return &contact, nil
}

// UpdateContact updates an existing contact
func (s *ContactService) UpdateContact(id uint, req *models.ContactRequest, updatedBy *uint) (*models.Contact, error) {
	// Get existing contact
	contact, err := s.GetContact(id)
	if err != nil {
		return nil, err
	}

	// Store original values for activity logging
	_ = contact.Status // originalStatus
	originalAssignedTo := contact.AssignedTo

	// Validate contact type and source if changed
	if req.ContactTypeID != contact.ContactTypeID || req.ContactSourceID != contact.ContactSourceID {
		if err := s.validateContactTypeAndSource(req.ContactTypeID, req.ContactSourceID); err != nil {
			return nil, err
		}
	}

	// Check for duplicates if email changed
	if req.Email != contact.Email {
		if duplicate, err := s.checkForDuplicates(req.Email, req.Phone); err != nil {
			return nil, fmt.Errorf("duplicate check failed: %v", err)
		} else if duplicate != nil && duplicate.ID != id {
			return nil, fmt.Errorf("contact already exists with email: %s", req.Email)
		}
	}

	// Update fields
	contact.FirstName = req.FirstName
	contact.LastName = req.LastName
	contact.Email = req.Email
	contact.Phone = req.Phone
	contact.Company = req.Company
	contact.JobTitle = req.JobTitle
	contact.Website = req.Website
	contact.AddressLine1 = req.AddressLine1
	contact.AddressLine2 = req.AddressLine2
	contact.City = req.City
	contact.State = req.State
	contact.PostalCode = req.PostalCode
	contact.ContactTypeID = req.ContactTypeID
	contact.ContactSourceID = req.ContactSourceID
	contact.Subject = req.Subject
	contact.Message = req.Message
	contact.Tags = req.Tags
	contact.CustomFields = req.CustomFields
	contact.Notes = req.Notes
	contact.UpdatedBy = updatedBy

	// Set optional fields
	if req.Country != nil {
		contact.Country = *req.Country
	}
	if req.PreferredContactMethod != nil {
		contact.PreferredContactMethod = *req.PreferredContactMethod
	}
	if req.Priority != nil {
		contact.Priority = *req.Priority
	}
	if req.EstimatedValue != nil {
		contact.EstimatedValue = *req.EstimatedValue
	}
	if req.AssignedTo != nil && *req.AssignedTo != *contact.AssignedTo {
		contact.AssignedTo = req.AssignedTo
		now := time.Now()
		contact.AssignedAt = &now
		contact.AssignedBy = updatedBy
	}
	if req.NextFollowupDate != nil {
		contact.NextFollowupDate = req.NextFollowupDate
	}
	if req.MarketingConsent != nil {
		contact.MarketingConsent = *req.MarketingConsent
	}
	if req.DataProcessingConsent != nil {
		contact.DataProcessingConsent = *req.DataProcessingConsent
	}
	if req.GDPRConsent != nil {
		contact.GDPRConsent = *req.GDPRConsent
	}

	// Save to database
	if err := s.db.Save(contact).Error; err != nil {
		logger.Error("Failed to update contact", err, map[string]interface{}{
			"contact_id": id,
			"email":      req.Email,
		})
		return nil, fmt.Errorf("failed to update contact: %v", err)
	}

	// Log activities for significant changes
	if originalAssignedTo != contact.AssignedTo {
		s.logContactActivity(contact.ID, "assignment_changed", map[string]interface{}{
			"old_assigned_to": originalAssignedTo,
			"new_assigned_to": contact.AssignedTo,
			"changed_by":      updatedBy,
		})
	}

	// Recalculate lead score if relevant fields changed
	if err := s.calculateLeadScore(contact.ID); err != nil {
		logger.Warn("Failed to recalculate lead score", map[string]interface{}{
			"contact_id": contact.ID,
			"error":      err.Error(),
		})
	}

	return contact, nil
}

// DeleteContact soft deletes a contact
func (s *ContactService) DeleteContact(id uint, deletedBy *uint) error {
	contact, err := s.GetContact(id)
	if err != nil {
		return err
	}

	now := time.Now()
	contact.DeletedAt = &now
	contact.UpdatedBy = deletedBy

	if err := s.db.Save(contact).Error; err != nil {
		logger.Error("Failed to delete contact", err, map[string]interface{}{
			"contact_id": id,
		})
		return fmt.Errorf("failed to delete contact: %v", err)
	}

	// Log activity
	s.logContactActivity(contact.ID, "contact_deleted", map[string]interface{}{
		"deleted_by": deletedBy,
	})

	return nil
}

// ListContacts retrieves contacts with filtering and pagination
func (s *ContactService) ListContacts(opts *ContactListOptions) ([]*models.Contact, int64, error) {
	query := s.db.Model(&models.Contact{}).
		Preload("ContactType").
		Preload("ContactSource").
		Where("deleted_at IS NULL")

	// Apply filters
	if opts.Status != "" {
		query = query.Where("status = ?", opts.Status)
	}
	if opts.Priority != "" {
		query = query.Where("priority = ?", opts.Priority)
	}
	if opts.AssignedTo != nil {
		query = query.Where("assigned_to = ?", *opts.AssignedTo)
	}
	if opts.SourceID != nil {
		query = query.Where("contact_source_id = ?", *opts.SourceID)
	}
	if opts.TypeID != nil {
		query = query.Where("contact_type_id = ?", *opts.TypeID)
	}
	if opts.Search != "" {
		searchTerm := "%" + strings.ToLower(opts.Search) + "%"
		query = query.Where(
			"LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(company) LIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm,
		)
	}
	if opts.DateFrom != nil {
		query = query.Where("created_at >= ?", *opts.DateFrom)
	}
	if opts.DateTo != nil {
		query = query.Where("created_at <= ?", *opts.DateTo)
	}

	// Apply tag filters
	if len(opts.Tags) > 0 {
		query = query.Joins("JOIN contact_tag_assignments cta ON contacts.id = cta.contact_id").
			Joins("JOIN contact_tags ct ON cta.tag_id = ct.id").
			Where("ct.name IN ?", opts.Tags).
			Group("contacts.id")
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count contacts: %v", err)
	}

	// Apply sorting
	sortBy := "created_at"
	sortOrder := "DESC"
	if opts.SortBy != "" {
		sortBy = opts.SortBy
	}
	if opts.SortOrder != "" {
		sortOrder = strings.ToUpper(opts.SortOrder)
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.PageSize < 1 || opts.PageSize > 100 {
		opts.PageSize = 10
	}
	offset := (opts.Page - 1) * opts.PageSize
	query = query.Offset(offset).Limit(opts.PageSize)

	// Execute query
	var contacts []*models.Contact
	if err := query.Find(&contacts).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list contacts: %v", err)
	}

	return contacts, total, nil
}

// UpdateContactStatus updates the status of a contact
func (s *ContactService) UpdateContactStatus(id uint, status models.ContactStatus, updatedBy *uint) error {
	contact, err := s.GetContact(id)
	if err != nil {
		return err
	}

	oldStatus := contact.Status
	contact.Status = status
	contact.UpdatedBy = updatedBy

	// Set conversion date if converting to won
	if status == models.StatusClosedWon && contact.ConversionDate == nil {
		now := time.Now()
		contact.ConversionDate = &now
	}

	// Set closed date if closing
	if (status == models.StatusClosedWon || status == models.StatusClosedLost) && contact.ClosedDate == nil {
		now := time.Now()
		contact.ClosedDate = &now
	}

	if err := s.db.Save(contact).Error; err != nil {
		return fmt.Errorf("failed to update contact status: %v", err)
	}

	// Log status change activity
	s.logContactActivity(contact.ID, "status_change", map[string]interface{}{
		"old_status": oldStatus,
		"new_status": status,
		"updated_by": updatedBy,
	})

	return nil
}

// SearchContacts performs advanced search on contacts
func (s *ContactService) SearchContacts(query string, filters map[string]interface{}) ([]*models.Contact, error) {
	dbQuery := s.db.Model(&models.Contact{}).
		Preload("ContactType").
		Preload("ContactSource").
		Where("deleted_at IS NULL")

	// Full-text search if supported
	if query != "" {
		searchTerm := "%" + strings.ToLower(query) + "%"
		dbQuery = dbQuery.Where(
			"LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(company) LIKE ? OR LOWER(subject) LIKE ? OR LOWER(message) LIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm,
		)
	}

	// Apply additional filters
	for key, value := range filters {
		switch key {
		case "status":
			dbQuery = dbQuery.Where("status = ?", value)
		case "priority":
			dbQuery = dbQuery.Where("priority = ?", value)
		case "assigned_to":
			dbQuery = dbQuery.Where("assigned_to = ?", value)
		case "lead_score_min":
			dbQuery = dbQuery.Where("lead_score >= ?", value)
		case "lead_score_max":
			dbQuery = dbQuery.Where("lead_score <= ?", value)
		case "estimated_value_min":
			dbQuery = dbQuery.Where("estimated_value >= ?", value)
		case "estimated_value_max":
			dbQuery = dbQuery.Where("estimated_value <= ?", value)
		}
	}

	var contacts []*models.Contact
	if err := dbQuery.Order("lead_score DESC, created_at DESC").Find(&contacts).Error; err != nil {
		return nil, fmt.Errorf("failed to search contacts: %v", err)
	}

	return contacts, nil
}

// Helper methods

func (s *ContactService) validateContactTypeAndSource(typeID, sourceID uint) error {
	var count int64
	
	// Check contact type exists and is active
	if err := s.db.Model(&models.ContactType{}).
		Where("id = ? AND is_active = true", typeID).
		Count(&count).Error; err != nil {
		return fmt.Errorf("failed to validate contact type: %v", err)
	}
	if count == 0 {
		return fmt.Errorf("contact type not found or inactive")
	}

	// Check contact source exists and is active
	if err := s.db.Model(&models.ContactSource{}).
		Where("id = ? AND is_active = true", sourceID).
		Count(&count).Error; err != nil {
		return fmt.Errorf("failed to validate contact source: %v", err)
	}
	if count == 0 {
		return fmt.Errorf("contact source not found or inactive")
	}

	return nil
}

func (s *ContactService) checkForDuplicates(email string, phone *string) (*models.Contact, error) {
	var contact models.Contact
	
	// Check for duplicate email only (primary key for contacts)
	// Phone numbers can be shared across family members or businesses
	query := s.db.Where("email = ? AND deleted_at IS NULL", email)
	
	if err := query.First(&contact).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No duplicates found
		}
		return nil, err
	}
	
	return &contact, nil
}

func (s *ContactService) calculateLeadScore(contactID uint) error {
	// This is a placeholder for lead scoring logic
	// In a real implementation, this would apply scoring rules
	// For now, we'll set a basic score based on available data
	
	var contact models.Contact
	if err := s.db.First(&contact, contactID).Error; err != nil {
		return err
	}

	score := 0
	
	// Basic scoring rules
	if contact.Company != nil && *contact.Company != "" {
		score += 10
	}
	if contact.Phone != nil && *contact.Phone != "" {
		score += 15
	}
	if contact.JobTitle != nil && *contact.JobTitle != "" {
		score += 12
	}
	if !strings.Contains(strings.ToLower(contact.Email), "gmail") &&
	   !strings.Contains(strings.ToLower(contact.Email), "yahoo") &&
	   !strings.Contains(strings.ToLower(contact.Email), "hotmail") {
		score += 10 // Business email
	}

	// Update lead score
	if err := s.db.Model(&contact).Update("lead_score", score).Error; err != nil {
		return err
	}

	return nil
}

func (s *ContactService) logContactActivity(contactID uint, activityType string, details map[string]interface{}) {
	logger.LogContactActivity(contactID, activityType, details)
}

// AdvancedSearch performs advanced search with multiple criteria
func (s *ContactService) AdvancedSearch(criteria *AdvancedSearchCriteria) ([]*models.Contact, int64, error) {
	query := s.db.Model(&models.Contact{}).
		Preload("ContactType").
		Preload("ContactSource").
		Where("deleted_at IS NULL")

	// Apply text search filters
	if criteria.FullTextSearch != nil && *criteria.FullTextSearch != "" {
		searchTerm := "%" + strings.ToLower(*criteria.FullTextSearch) + "%"
		query = query.Where(
			"LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(company) LIKE ? OR LOWER(job_title) LIKE ? OR LOWER(subject) LIKE ? OR LOWER(message) LIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm,
		)
	}

	if criteria.FirstName != nil && *criteria.FirstName != "" {
		query = query.Where("LOWER(first_name) LIKE ?", "%"+strings.ToLower(*criteria.FirstName)+"%")
	}
	if criteria.LastName != nil && *criteria.LastName != "" {
		query = query.Where("LOWER(last_name) LIKE ?", "%"+strings.ToLower(*criteria.LastName)+"%")
	}
	if criteria.Email != nil && *criteria.Email != "" {
		query = query.Where("LOWER(email) LIKE ?", "%"+strings.ToLower(*criteria.Email)+"%")
	}
	if criteria.Phone != nil && *criteria.Phone != "" {
		query = query.Where("phone LIKE ?", "%"+*criteria.Phone+"%")
	}
	if criteria.Company != nil && *criteria.Company != "" {
		query = query.Where("LOWER(company) LIKE ?", "%"+strings.ToLower(*criteria.Company)+"%")
	}
	if criteria.JobTitle != nil && *criteria.JobTitle != "" {
		query = query.Where("LOWER(job_title) LIKE ?", "%"+strings.ToLower(*criteria.JobTitle)+"%")
	}
	if criteria.Country != nil && *criteria.Country != "" {
		query = query.Where("LOWER(country) LIKE ?", "%"+strings.ToLower(*criteria.Country)+"%")
	}

	// Apply status and priority filters
	if criteria.Status != nil && *criteria.Status != "" {
		query = query.Where("status = ?", *criteria.Status)
	}
	if criteria.Priority != nil && *criteria.Priority != "" {
		query = query.Where("priority = ?", *criteria.Priority)
	}

	// Apply numeric range filters
	if criteria.LeadScoreMin != nil {
		query = query.Where("lead_score >= ?", *criteria.LeadScoreMin)
	}
	if criteria.LeadScoreMax != nil {
		query = query.Where("lead_score <= ?", *criteria.LeadScoreMax)
	}
	if criteria.EstimatedValueMin != nil {
		query = query.Where("estimated_value >= ?", *criteria.EstimatedValueMin)
	}
	if criteria.EstimatedValueMax != nil {
		query = query.Where("estimated_value <= ?", *criteria.EstimatedValueMax)
	}

	// Apply reference filters
	if criteria.AssignedTo != nil {
		query = query.Where("assigned_to = ?", *criteria.AssignedTo)
	}
	if criteria.SourceID != nil {
		query = query.Where("contact_source_id = ?", *criteria.SourceID)
	}
	if criteria.TypeID != nil {
		query = query.Where("contact_type_id = ?", *criteria.TypeID)
	}

	// Apply date filters
	if criteria.CreatedFrom != nil {
		query = query.Where("created_at >= ?", *criteria.CreatedFrom)
	}
	if criteria.CreatedTo != nil {
		query = query.Where("created_at <= ?", *criteria.CreatedTo)
	}
	if criteria.LastContactFrom != nil {
		query = query.Where("last_contact_date >= ?", *criteria.LastContactFrom)
	}
	if criteria.LastContactTo != nil {
		query = query.Where("last_contact_date <= ?", *criteria.LastContactTo)
	}

	// Apply boolean filters
	if criteria.HasActivities != nil && *criteria.HasActivities {
		query = query.Where("total_interactions > 0")
	}
	if criteria.IsHotLead != nil && *criteria.IsHotLead {
		query = query.Where("lead_score >= 80")
	}
	if criteria.IsHighPriority != nil && *criteria.IsHighPriority {
		query = query.Where("priority = 'high'")
	}

	// Apply tag filters
	if len(criteria.Tags) > 0 {
		tagConditions := make([]string, len(criteria.Tags))
		tagValues := make([]interface{}, len(criteria.Tags))
		for i, tag := range criteria.Tags {
			tagConditions[i] = "JSON_CONTAINS(tags, ?)"
			tagBytes, _ := json.Marshal(tag)
			tagValues[i] = string(tagBytes)
		}
		query = query.Where(strings.Join(tagConditions, " OR "), tagValues...)
	}

	// Get total count
	var total int64
	countQuery := *query // Create a copy for counting
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %v", err)
	}

	// Apply sorting
	sortBy := "created_at"
	sortOrder := "DESC"
	if criteria.SortBy != "" {
		sortBy = criteria.SortBy
	}
	if criteria.SortOrder != "" {
		sortOrder = strings.ToUpper(criteria.SortOrder)
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination
	if criteria.Page < 1 {
		criteria.Page = 1
	}
	if criteria.PageSize < 1 || criteria.PageSize > 100 {
		criteria.PageSize = 20
	}
	offset := (criteria.Page - 1) * criteria.PageSize
	query = query.Offset(offset).Limit(criteria.PageSize)

	// Execute query
	var contacts []*models.Contact
	if err := query.Find(&contacts).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to execute advanced search: %v", err)
	}

	return contacts, total, nil
}

// GetSearchSuggestions returns suggestions for autocomplete
func (s *ContactService) GetSearchSuggestions(field, query string, limit int) ([]string, error) {
	var results []string
	
	dbQuery := s.db.Model(&models.Contact{}).
		Where("deleted_at IS NULL").
		Where(fmt.Sprintf("LOWER(%s) LIKE ?", field), "%"+strings.ToLower(query)+"%").
		Group(field).
		Order(fmt.Sprintf("COUNT(%s) DESC", field)).
		Limit(limit)

	rows, err := dbQuery.Select(field).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get suggestions: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			continue
		}
		if value != "" {
			results = append(results, value)
		}
	}

	return results, nil
}

// GetSavedSearches retrieves saved searches for a user
func (s *ContactService) GetSavedSearches(userID uint) ([]*models.SavedSearch, error) {
	var searches []*models.SavedSearch
	
	if err := s.db.Where("user_id = ? OR is_public = true", userID).
		Order("name ASC").
		Find(&searches).Error; err != nil {
		return nil, fmt.Errorf("failed to get saved searches: %v", err)
	}

	return searches, nil
}

// SaveSearch saves a search query for later use
func (s *ContactService) SaveSearch(userID uint, req *models.SavedSearchRequest) (*models.SavedSearch, error) {
	// Check if name already exists for this user
	var count int64
	if err := s.db.Model(&models.SavedSearch{}).
		Where("user_id = ? AND name = ?", userID, req.Name).
		Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to check for duplicate search name: %v", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("search with name '%s' already exists", req.Name)
	}

	// Convert criteria to JSON
	criteriaJSON, err := json.Marshal(req.Criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize search criteria: %v", err)
	}

	savedSearch := &models.SavedSearch{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Criteria:    criteriaJSON,
		IsPublic:    req.IsPublic,
	}

	if err := s.db.Create(savedSearch).Error; err != nil {
		return nil, fmt.Errorf("failed to save search: %v", err)
	}

	return savedSearch, nil
}

// DeleteSavedSearch deletes a saved search
func (s *ContactService) DeleteSavedSearch(searchID, userID uint) error {
	result := s.db.Where("id = ? AND user_id = ?", searchID, userID).
		Delete(&models.SavedSearch{})
	
	if result.Error != nil {
		return fmt.Errorf("failed to delete saved search: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("saved search not found")
	}

	return nil
}

// ExecuteSavedSearch executes a previously saved search
func (s *ContactService) ExecuteSavedSearch(searchID, userID uint, page, pageSize int) ([]*models.Contact, int64, error) {
	// Get saved search
	var savedSearch models.SavedSearch
	if err := s.db.Where("id = ? AND (user_id = ? OR is_public = true)", searchID, userID).
		First(&savedSearch).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, fmt.Errorf("saved search not found")
		}
		return nil, 0, fmt.Errorf("failed to get saved search: %v", err)
	}

	// Parse criteria
	var criteriaMap map[string]interface{}
	if err := json.Unmarshal(savedSearch.Criteria, &criteriaMap); err != nil {
		return nil, 0, fmt.Errorf("failed to parse search criteria: %v", err)
	}

	// Convert to AdvancedSearchCriteria
	criteria := &AdvancedSearchCriteria{
		Page:     page,
		PageSize: pageSize,
		SortBy:   "created_at",
		SortOrder: "DESC",
	}

	// Map the criteria fields (simplified version)
	if val, ok := criteriaMap["full_text_search"].(string); ok && val != "" {
		criteria.FullTextSearch = &val
	}
	if val, ok := criteriaMap["status"].(string); ok && val != "" {
		criteria.Status = &val
	}
	if val, ok := criteriaMap["priority"].(string); ok && val != "" {
		criteria.Priority = &val
	}
	// Add more field mappings as needed...

	// Execute the search
	return s.AdvancedSearch(criteria)
}

