package services

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"contact-service/internal/models"
	"contact-service/internal/repository"
)

// BulkService handles bulk operations for contacts
type BulkService struct {
	contactRepo repository.ContactRepository
	userRepo    repository.UserRepository
}

// NewBulkService creates a new bulk service instance
func NewBulkService(contactRepo repository.ContactRepository, userRepo repository.UserRepository) *BulkService {
	return &BulkService{
		contactRepo: contactRepo,
		userRepo:    userRepo,
	}
}

// BulkImportResult represents the result of a bulk import operation
type BulkImportResult struct {
	TotalRecords    int                    `json:"total_records"`
	SuccessCount    int                    `json:"success_count"`
	ErrorCount      int                    `json:"error_count"`
	Errors          []BulkImportError      `json:"errors,omitempty"`
	ImportedIDs     []uint                 `json:"imported_ids"`
	ProcessingTime  time.Duration          `json:"processing_time"`
	Summary         map[string]interface{} `json:"summary"`
}

// BulkImportError represents an error during bulk import
type BulkImportError struct {
	Row     int    `json:"row"`
	Field   string `json:"field,omitempty"`
	Value   string `json:"value,omitempty"`
	Message string `json:"message"`
}

// BulkUpdateRequest represents a bulk update request
type BulkUpdateRequest struct {
	ContactIDs []uint                 `json:"contact_ids"`
	Updates    map[string]interface{} `json:"updates"`
	Conditions map[string]interface{} `json:"conditions,omitempty"`
}

// BulkUpdateResult represents the result of a bulk update operation
type BulkUpdateResult struct {
	UpdatedCount   int           `json:"updated_count"`
	SkippedCount   int           `json:"skipped_count"`
	ErrorCount     int           `json:"error_count"`
	Errors         []string      `json:"errors,omitempty"`
	UpdatedIDs     []uint        `json:"updated_ids"`
	ProcessingTime time.Duration `json:"processing_time"`
}

// ExportFormat represents supported export formats
type ExportFormat string

const (
	ExportFormatCSV  ExportFormat = "csv"
	ExportFormatJSON ExportFormat = "json"
	ExportFormatXLSX ExportFormat = "xlsx"
)

// ExportRequest represents an export request
type ExportRequest struct {
	Format      ExportFormat               `json:"format"`
	Filters     map[string]interface{}     `json:"filters,omitempty"`
	Fields      []string                   `json:"fields,omitempty"`
	SortBy      string                     `json:"sort_by,omitempty"`
	SortOrder   string                     `json:"sort_order,omitempty"`
	Limit       int                        `json:"limit,omitempty"`
	IncludeMeta bool                       `json:"include_meta,omitempty"`
}

// CSV column headers and their corresponding model fields
var csvHeaders = map[string]string{
	"Name":         "name",
	"Email":        "email",
	"Phone":        "phone",
	"Company":      "company",
	"Position":     "position",
	"Status":       "status",
	"Type":         "type",
	"Source":       "source",
	"Notes":        "notes",
	"AssignedTo":   "assigned_to",
	"CreatedAt":    "created_at",
	"UpdatedAt":    "updated_at",
}

// ImportContactsFromCSV imports contacts from CSV data
func (s *BulkService) ImportContactsFromCSV(data io.Reader, skipHeader bool) (*BulkImportResult, error) {
	startTime := time.Now()
	
	result := &BulkImportResult{
		Errors:      make([]BulkImportError, 0),
		ImportedIDs: make([]uint, 0),
		Summary:     make(map[string]interface{}),
	}

	reader := csv.NewReader(data)
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV data: %w", err)
	}

	if len(records) == 0 {
		return result, nil
	}

	result.TotalRecords = len(records)
	
	// Skip header row if requested
	startRow := 0
	var headers []string
	if skipHeader && len(records) > 0 {
		headers = records[0]
		startRow = 1
		result.TotalRecords--
	} else {
		// Use default headers if no header row
		headers = []string{"Name", "Email", "Phone", "Company", "Position", "Status", "Type", "Source", "Notes"}
	}

	// Process each record
	for i := startRow; i < len(records); i++ {
		row := records[i]
		rowNum := i + 1

		contact, validationErrors := s.parseCSVRow(row, headers, rowNum)
		
		// Add validation errors
		for _, validationError := range validationErrors {
			result.Errors = append(result.Errors, validationError)
			result.ErrorCount++
		}

		// Skip if there were validation errors
		if len(validationErrors) > 0 {
			continue
		}

		// Check for duplicate email
		if existing, _ := s.contactRepo.GetByEmail(contact.Email); existing != nil {
			result.Errors = append(result.Errors, BulkImportError{
				Row:     rowNum,
				Field:   "Email",
				Value:   contact.Email,
				Message: "Contact with this email already exists",
			})
			result.ErrorCount++
			continue
		}

		// Create contact
		if err := s.contactRepo.Create(contact); err != nil {
			result.Errors = append(result.Errors, BulkImportError{
				Row:     rowNum,
				Message: fmt.Sprintf("Failed to create contact: %v", err),
			})
			result.ErrorCount++
			continue
		}

		result.ImportedIDs = append(result.ImportedIDs, contact.ID)
		result.SuccessCount++
	}

	result.ProcessingTime = time.Since(startTime)
	
	// Generate summary
	result.Summary = map[string]interface{}{
		"success_rate":     float64(result.SuccessCount) / float64(result.TotalRecords) * 100,
		"error_rate":       float64(result.ErrorCount) / float64(result.TotalRecords) * 100,
		"processing_speed": fmt.Sprintf("%.2f records/second", float64(result.TotalRecords)/result.ProcessingTime.Seconds()),
		"duplicate_emails": countErrorsByType(result.Errors, "already exists"),
		"validation_errors": countErrorsByType(result.Errors, "validation"),
	}

	return result, nil
}

// parseCSVRow parses a CSV row into a Contact model
func (s *BulkService) parseCSVRow(row []string, headers []string, rowNum int) (*models.Contact, []BulkImportError) {
	contact := &models.Contact{
		Status: "new", // Default status
	}
	
	var errors []BulkImportError

	for i, value := range row {
		if i >= len(headers) {
			break
		}

		header := headers[i]
		value = strings.TrimSpace(value)

		if value == "" {
			continue
		}

		switch strings.ToLower(header) {
		case "name", "first_name":
			if len(value) < 2 {
				errors = append(errors, BulkImportError{
					Row:     rowNum,
					Field:   "FirstName",
					Value:   value,
					Message: "First name must be at least 2 characters",
				})
			} else {
				contact.FirstName = value
			}

		case "last_name":
			if value != "" {
				contact.LastName = &value
			}

		case "email":
			if !isValidEmail(value) {
				errors = append(errors, BulkImportError{
					Row:     rowNum,
					Field:   "Email",
					Value:   value,
					Message: "Invalid email format",
				})
			} else {
				contact.Email = value
			}

		case "phone":
			if value != "" {
				contact.Phone = &value
			}

		case "company":
			if value != "" {
				contact.Company = &value
			}

		case "position", "job_title":
			if value != "" {
				contact.JobTitle = &value
			}

		case "status":
			if isValidStatus(value) {
				contact.Status = models.ContactStatus(value)
			} else {
				errors = append(errors, BulkImportError{
					Row:     rowNum,
					Field:   "Status",
					Value:   value,
					Message: "Invalid status. Must be one of: new, contacted, qualified, customer, inactive",
				})
			}

		case "notes":
			if value != "" {
				contact.Notes = &value
			}

		case "type":
			// Handle type lookup if needed
			contact.ContactTypeID = 1 // Default type

		case "source":
			// Handle source lookup if needed
			contact.ContactSourceID = 1 // Default source
		}
	}

	// Validate required fields
	if contact.FirstName == "" {
		errors = append(errors, BulkImportError{
			Row:     rowNum,
			Field:   "FirstName",
			Message: "First name is required",
		})
	}

	if contact.Email == "" {
		errors = append(errors, BulkImportError{
			Row:     rowNum,
			Field:   "Email",
			Message: "Email is required",
		})
	}

	return contact, errors
}

// ExportContactsToCSV exports contacts to CSV format
func (s *BulkService) ExportContactsToCSV(request ExportRequest) ([]byte, error) {
	// Build query parameters
	params := repository.ContactListParams{
		Page:  1,
		Limit: request.Limit,
		Sort:  request.SortBy,
		Order: request.SortOrder,
	}

	if params.Limit == 0 {
		params.Limit = 10000 // Default large limit for export
	}

	// Apply filters
	if request.Filters != nil {
		if status, ok := request.Filters["status"].(string); ok {
			params.Status = status
		}
		if typeID, ok := request.Filters["type_id"].(float64); ok {
			params.TypeID = uint(typeID)
		}
		if sourceID, ok := request.Filters["source_id"].(float64); ok {
			params.SourceID = uint(sourceID)
		}
	}

	// Get contacts
	contacts, _, err := s.contactRepo.List(params)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve contacts: %w", err)
	}

	// Create CSV buffer
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Determine fields to export
	fields := request.Fields
	if len(fields) == 0 {
		fields = []string{"FirstName", "LastName", "Email", "Phone", "Company", "JobTitle", "Status", "Notes", "CreatedAt", "UpdatedAt"}
	}

	// Write header
	if err := writer.Write(fields); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, contact := range contacts {
		row := make([]string, len(fields))
		
		for i, field := range fields {
			switch strings.ToLower(field) {
			case "firstname":
				row[i] = contact.FirstName
			case "lastname":
				if contact.LastName != nil {
					row[i] = *contact.LastName
				}
			case "email":
				row[i] = contact.Email
			case "phone":
				if contact.Phone != nil {
					row[i] = *contact.Phone
				}
			case "company":
				if contact.Company != nil {
					row[i] = *contact.Company
				}
			case "jobtitle":
				if contact.JobTitle != nil {
					row[i] = *contact.JobTitle
				}
			case "status":
				row[i] = string(contact.Status)
			case "notes":
				if contact.Notes != nil {
					row[i] = *contact.Notes
				}
			case "createdat":
				row[i] = contact.CreatedAt.Format("2006-01-02 15:04:05")
			case "updatedat":
				row[i] = contact.UpdatedAt.Format("2006-01-02 15:04:05")
			default:
				row[i] = ""
			}
		}

		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
}

// ExportContactsToJSON exports contacts to JSON format
func (s *BulkService) ExportContactsToJSON(request ExportRequest) ([]byte, error) {
	// Build query parameters
	params := repository.ContactListParams{
		Page:  1,
		Limit: request.Limit,
		Sort:  request.SortBy,
		Order: request.SortOrder,
	}

	if params.Limit == 0 {
		params.Limit = 10000 // Default large limit for export
	}

	// Apply filters
	if request.Filters != nil {
		if status, ok := request.Filters["status"].(string); ok {
			params.Status = status
		}
	}

	// Get contacts
	contacts, total, err := s.contactRepo.List(params)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve contacts: %w", err)
	}

	// Create export structure
	exportData := map[string]interface{}{
		"contacts": contacts,
		"total":    total,
	}

	if request.IncludeMeta {
		exportData["meta"] = map[string]interface{}{
			"exported_at": time.Now(),
			"format":      "json",
			"filters":     request.Filters,
			"fields":      request.Fields,
		}
	}

	return json.MarshalIndent(exportData, "", "  ")
}

// BulkUpdateContacts performs bulk updates on contacts
func (s *BulkService) BulkUpdateContacts(request BulkUpdateRequest) (*BulkUpdateResult, error) {
	startTime := time.Now()
	
	result := &BulkUpdateResult{
		Errors:     make([]string, 0),
		UpdatedIDs: make([]uint, 0),
	}

	// Validate request
	if len(request.ContactIDs) == 0 {
		return nil, fmt.Errorf("no contact IDs provided")
	}

	if len(request.Updates) == 0 {
		return nil, fmt.Errorf("no updates provided")
	}

	// Process each contact
	for _, contactID := range request.ContactIDs {
		// Get existing contact
		contact, err := s.contactRepo.GetByID(contactID)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Contact ID %d: %v", contactID, err))
			result.ErrorCount++
			continue
		}

		// Check conditions if provided
		if !s.checkUpdateConditions(contact, request.Conditions) {
			result.SkippedCount++
			continue
		}

		// Apply updates
		updated := false
		if status, ok := request.Updates["status"].(string); ok && isValidStatus(status) {
			contact.Status = models.ContactStatus(status)
			updated = true
		}

		if assignedTo, ok := request.Updates["assigned_to"].(float64); ok {
			userID := uint(assignedTo)
			contact.AssignedTo = &userID
			updated = true
		}

		if notes, ok := request.Updates["notes"].(string); ok {
			contact.Notes = &notes
			updated = true
		}

		if company, ok := request.Updates["company"].(string); ok {
			contact.Company = &company
			updated = true
		}

		if jobTitle, ok := request.Updates["job_title"].(string); ok {
			contact.JobTitle = &jobTitle
			updated = true
		}

		if !updated {
			result.Errors = append(result.Errors, fmt.Sprintf("Contact ID %d: No valid updates provided", contactID))
			result.ErrorCount++
			continue
		}

		// Save updated contact
		if err := s.contactRepo.Update(contact); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Contact ID %d: Failed to update - %v", contactID, err))
			result.ErrorCount++
			continue
		}

		result.UpdatedIDs = append(result.UpdatedIDs, contactID)
		result.UpdatedCount++
	}

	result.ProcessingTime = time.Since(startTime)
	return result, nil
}

// BulkDeleteContacts performs bulk deletion of contacts
func (s *BulkService) BulkDeleteContacts(contactIDs []uint) (*BulkUpdateResult, error) {
	startTime := time.Now()
	
	result := &BulkUpdateResult{
		Errors:     make([]string, 0),
		UpdatedIDs: make([]uint, 0),
	}

	if len(contactIDs) == 0 {
		return result, nil
	}

	// Process each contact
	for _, contactID := range contactIDs {
		if err := s.contactRepo.Delete(contactID); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Contact ID %d: %v", contactID, err))
			result.ErrorCount++
			continue
		}

		result.UpdatedIDs = append(result.UpdatedIDs, contactID)
		result.UpdatedCount++
	}

	result.ProcessingTime = time.Since(startTime)
	return result, nil
}

// checkUpdateConditions checks if contact meets update conditions
func (s *BulkService) checkUpdateConditions(contact *models.Contact, conditions map[string]interface{}) bool {
	if conditions == nil {
		return true
	}

	for field, expectedValue := range conditions {
		switch field {
		case "status":
			if string(contact.Status) != expectedValue.(string) {
				return false
			}
		case "company":
			if contact.Company == nil || *contact.Company != expectedValue.(string) {
				return false
			}
		case "assigned_to":
			expectedID := uint(expectedValue.(float64))
			if contact.AssignedTo == nil || *contact.AssignedTo != expectedID {
				return false
			}
		}
	}

	return true
}

// Helper functions
func countErrorsByType(errors []BulkImportError, errorType string) int {
	count := 0
	for _, err := range errors {
		if strings.Contains(strings.ToLower(err.Message), strings.ToLower(errorType)) {
			count++
		}
	}
	return count
}

func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".") && len(email) > 5
}

func isValidStatus(status string) bool {
	validStatuses := []string{"new", "contacted", "qualified", "customer", "inactive"}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}