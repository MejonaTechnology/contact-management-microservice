package handlers

import (
	"contact-service/internal/models"
	"contact-service/internal/services"
	"contact-service/pkg/logger"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ContactHandler handles HTTP requests for contact management
type ContactHandler struct {
	contactService *services.ContactService
}

// NewContactHandler creates a new contact handler
func NewContactHandler() *ContactHandler {
	return &ContactHandler{
		contactService: services.NewContactService(),
	}
}

// CreateContact godoc
// @Summary Create a new contact
// @Description Create a new contact with the provided information
// @Tags contacts
// @Accept json
// @Produce json
// @Param contact body models.ContactRequest true "Contact information"
// @Success 201 {object} APIResponse{data=models.ContactResponse}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /contacts [post]
func (h *ContactHandler) CreateContact(c *gin.Context) {
	var req models.ContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid contact creation request", map[string]interface{}{
			"error": err.Error(),
			"ip":    c.ClientIP(),
		})
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request data", err.Error()))
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := getUserIDFromContext(c)

	start := time.Now()
	contact, err := h.contactService.CreateContact(&req, userID)
	duration := time.Since(start)

	logger.LogAPIRequest(c.Request.Method, c.Request.URL.Path, userID, duration, http.StatusCreated)

	if err != nil {
		logger.Error("Failed to create contact", err, map[string]interface{}{
			"email":   req.Email,
			"user_id": userID,
		})
		
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, NewErrorResponse("Contact already exists", err.Error()))
			return
		}
		
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to create contact", ""))
		return
	}

	response := h.mapContactToResponse(contact)
	c.JSON(http.StatusCreated, NewSuccessResponse("Contact created successfully", response))
}

// GetContact godoc
// @Summary Get a contact by ID
// @Description Retrieve a contact by its ID
// @Tags contacts
// @Accept json
// @Produce json
// @Param id path int true "Contact ID"
// @Success 200 {object} APIResponse{data=models.ContactResponse}
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /contacts/{id} [get]
func (h *ContactHandler) GetContact(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid contact ID", ""))
		return
	}

	userID := getUserIDFromContext(c)
	start := time.Now()
	contact, err := h.contactService.GetContact(uint(id))
	duration := time.Since(start)

	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		
		logger.LogAPIRequest(c.Request.Method, c.Request.URL.Path, userID, duration, status)
		c.JSON(status, NewErrorResponse("Contact not found", ""))
		return
	}

	logger.LogAPIRequest(c.Request.Method, c.Request.URL.Path, userID, duration, http.StatusOK)

	response := h.mapContactToResponse(contact)
	c.JSON(http.StatusOK, NewSuccessResponse("Contact retrieved successfully", response))
}

// UpdateContact godoc
// @Summary Update a contact
// @Description Update an existing contact with new information
// @Tags contacts
// @Accept json
// @Produce json
// @Param id path int true "Contact ID"
// @Param contact body models.ContactRequest true "Updated contact information"
// @Success 200 {object} APIResponse{data=models.ContactResponse}
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /contacts/{id} [put]
func (h *ContactHandler) UpdateContact(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid contact ID", ""))
		return
	}

	var req models.ContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid contact update request", map[string]interface{}{
			"error":      err.Error(),
			"contact_id": id,
			"ip":         c.ClientIP(),
		})
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request data", err.Error()))
		return
	}

	userID := getUserIDFromContext(c)
	start := time.Now()
	contact, err := h.contactService.UpdateContact(uint(id), &req, userID)
	duration := time.Since(start)

	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		} else if strings.Contains(err.Error(), "already exists") {
			status = http.StatusConflict
		}
		
		logger.LogAPIRequest(c.Request.Method, c.Request.URL.Path, userID, duration, status)
		c.JSON(status, NewErrorResponse("Failed to update contact", err.Error()))
		return
	}

	logger.LogAPIRequest(c.Request.Method, c.Request.URL.Path, userID, duration, http.StatusOK)

	response := h.mapContactToResponse(contact)
	c.JSON(http.StatusOK, NewSuccessResponse("Contact updated successfully", response))
}

// DeleteContact godoc
// @Summary Delete a contact
// @Description Soft delete a contact by ID
// @Tags contacts
// @Accept json
// @Produce json
// @Param id path int true "Contact ID"
// @Success 200 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /contacts/{id} [delete]
func (h *ContactHandler) DeleteContact(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid contact ID", ""))
		return
	}

	userID := getUserIDFromContext(c)
	start := time.Now()
	err = h.contactService.DeleteContact(uint(id), userID)
	duration := time.Since(start)

	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		
		logger.LogAPIRequest(c.Request.Method, c.Request.URL.Path, userID, duration, status)
		c.JSON(status, NewErrorResponse("Failed to delete contact", ""))
		return
	}

	logger.LogAPIRequest(c.Request.Method, c.Request.URL.Path, userID, duration, http.StatusOK)
	c.JSON(http.StatusOK, NewSuccessResponse("Contact deleted successfully", nil))
}

// ListContacts godoc
// @Summary List contacts with pagination and filtering
// @Description Retrieve a paginated list of contacts with optional filtering
// @Tags contacts
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param status query string false "Filter by status"
// @Param priority query string false "Filter by priority"
// @Param assigned_to query int false "Filter by assigned user"
// @Param source_id query int false "Filter by contact source"
// @Param type_id query int false "Filter by contact type"
// @Param search query string false "Search term"
// @Param tags query string false "Filter by tags (comma-separated)"
// @Param sort_by query string false "Sort field" default(created_at)
// @Param sort_order query string false "Sort order (ASC/DESC)" default(DESC)
// @Param date_from query string false "Filter from date (YYYY-MM-DD)"
// @Param date_to query string false "Filter to date (YYYY-MM-DD)"
// @Success 200 {object} APIResponse{data=[]models.ContactResponse,meta=PaginationMeta}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /contacts [get]
func (h *ContactHandler) ListContacts(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	
	opts := &services.ContactListOptions{
		Page:      page,
		PageSize:  pageSize,
		Status:    c.Query("status"),
		Priority:  c.Query("priority"),
		Search:    c.Query("search"),
		SortBy:    c.DefaultQuery("sort_by", "created_at"),
		SortOrder: c.DefaultQuery("sort_order", "DESC"),
	}

	// Parse optional filters
	if assignedTo := c.Query("assigned_to"); assignedTo != "" {
		if id, err := strconv.ParseUint(assignedTo, 10, 32); err == nil {
			assignedToUint := uint(id)
			opts.AssignedTo = &assignedToUint
		}
	}
	
	if sourceID := c.Query("source_id"); sourceID != "" {
		if id, err := strconv.ParseUint(sourceID, 10, 32); err == nil {
			sourceIDUint := uint(id)
			opts.SourceID = &sourceIDUint
		}
	}
	
	if typeID := c.Query("type_id"); typeID != "" {
		if id, err := strconv.ParseUint(typeID, 10, 32); err == nil {
			typeIDUint := uint(id)
			opts.TypeID = &typeIDUint
		}
	}

	// Parse date filters
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if date, err := time.Parse("2006-01-02", dateFrom); err == nil {
			opts.DateFrom = &date
		}
	}
	
	if dateTo := c.Query("date_to"); dateTo != "" {
		if date, err := time.Parse("2006-01-02", dateTo); err == nil {
			// Set to end of day
			endOfDay := date.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			opts.DateTo = &endOfDay
		}
	}

	// Parse tags
	if tags := c.Query("tags"); tags != "" {
		opts.Tags = strings.Split(tags, ",")
	}

	userID := getUserIDFromContext(c)
	start := time.Now()
	contacts, total, err := h.contactService.ListContacts(opts)
	duration := time.Since(start)

	logger.LogAPIRequest(c.Request.Method, c.Request.URL.Path, userID, duration, http.StatusOK)

	if err != nil {
		logger.Error("Failed to list contacts", err, map[string]interface{}{
			"user_id": userID,
			"filters": opts,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to retrieve contacts", ""))
		return
	}

	// Map to response format
	responses := make([]*models.ContactResponse, len(contacts))
	for i, contact := range contacts {
		responses[i] = h.mapContactToResponse(contact)
	}

	// Create pagination metadata
	meta := NewPaginationMeta(page, pageSize, total)

	c.JSON(http.StatusOK, NewPaginatedResponse("Contacts retrieved successfully", responses, meta))
}

// UpdateContactStatus godoc
// @Summary Update contact status
// @Description Update the status of a contact
// @Tags contacts
// @Accept json
// @Produce json
// @Param id path int true "Contact ID"
// @Param status body StatusUpdateRequest true "Status update information"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /contacts/{id}/status [put]
func (h *ContactHandler) UpdateContactStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid contact ID", ""))
		return
	}

	var req StatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request data", err.Error()))
		return
	}

	userID := getUserIDFromContext(c)
	start := time.Now()
	err = h.contactService.UpdateContactStatus(uint(id), models.ContactStatus(req.Status), userID)
	duration := time.Since(start)

	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		
		logger.LogAPIRequest(c.Request.Method, c.Request.URL.Path, userID, duration, status)
		c.JSON(status, NewErrorResponse("Failed to update contact status", ""))
		return
	}

	logger.LogAPIRequest(c.Request.Method, c.Request.URL.Path, userID, duration, http.StatusOK)
	c.JSON(http.StatusOK, NewSuccessResponse("Contact status updated successfully", gin.H{
		"contact_id": id,
		"status":     req.Status,
	}))
}

// SearchContacts godoc
// @Summary Search contacts
// @Description Search contacts using advanced search functionality
// @Tags contacts
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param status query string false "Filter by status"
// @Param priority query string false "Filter by priority"
// @Param assigned_to query int false "Filter by assigned user"
// @Param lead_score_min query int false "Minimum lead score"
// @Param lead_score_max query int false "Maximum lead score"
// @Param estimated_value_min query float64 false "Minimum estimated value"
// @Param estimated_value_max query float64 false "Maximum estimated value"
// @Success 200 {object} APIResponse{data=[]models.ContactResponse}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /contacts/search [get]
func (h *ContactHandler) SearchContacts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Search query is required", ""))
		return
	}

	// Build filters
	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if priority := c.Query("priority"); priority != "" {
		filters["priority"] = priority
	}
	if assignedTo := c.Query("assigned_to"); assignedTo != "" {
		if id, err := strconv.ParseUint(assignedTo, 10, 32); err == nil {
			filters["assigned_to"] = uint(id)
		}
	}
	if scoreMin := c.Query("lead_score_min"); scoreMin != "" {
		if score, err := strconv.Atoi(scoreMin); err == nil {
			filters["lead_score_min"] = score
		}
	}
	if scoreMax := c.Query("lead_score_max"); scoreMax != "" {
		if score, err := strconv.Atoi(scoreMax); err == nil {
			filters["lead_score_max"] = score
		}
	}
	if valueMin := c.Query("estimated_value_min"); valueMin != "" {
		if value, err := strconv.ParseFloat(valueMin, 64); err == nil {
			filters["estimated_value_min"] = value
		}
	}
	if valueMax := c.Query("estimated_value_max"); valueMax != "" {
		if value, err := strconv.ParseFloat(valueMax, 64); err == nil {
			filters["estimated_value_max"] = value
		}
	}

	userID := getUserIDFromContext(c)
	start := time.Now()
	contacts, err := h.contactService.SearchContacts(query, filters)
	duration := time.Since(start)

	logger.LogAPIRequest(c.Request.Method, c.Request.URL.Path, userID, duration, http.StatusOK)

	if err != nil {
		logger.Error("Failed to search contacts", err, map[string]interface{}{
			"query":   query,
			"filters": filters,
			"user_id": userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Search failed", ""))
		return
	}

	// Map to response format
	responses := make([]*models.ContactResponse, len(contacts))
	for i, contact := range contacts {
		responses[i] = h.mapContactToResponse(contact)
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Search completed successfully", responses))
}

// Helper methods

func (h *ContactHandler) mapContactToResponse(contact *models.Contact) *models.ContactResponse {
	response := &models.ContactResponse{
		ID:                    contact.ID,
		FirstName:             contact.FirstName,
		LastName:              contact.LastName,
		FullName:              contact.GetFullName(),
		DisplayName:           contact.GetDisplayName(),
		Email:                 contact.Email,
		Phone:                 contact.Phone,
		Company:               contact.Company,
		JobTitle:              contact.JobTitle,
		Website:               contact.Website,
		Country:               contact.Country,
		Subject:               contact.Subject,
		Status:                contact.Status,
		Priority:              contact.Priority,
		LeadScore:             contact.LeadScore,
		EstimatedValue:        contact.EstimatedValue,
		AssignedTo:            contact.AssignedTo,
		AssignedAt:            contact.AssignedAt,
		LastContactDate:       contact.LastContactDate,
		NextFollowupDate:      contact.NextFollowupDate,
		ResponseTimeHours:     contact.ResponseTimeHours,
		TotalInteractions:     contact.TotalInteractions,
		FirstContactDate:      contact.FirstContactDate,
		LastActivityDate:      contact.LastActivityDate,
		Tags:                  contact.Tags,
		CustomFields:          contact.CustomFields,
		Notes:                 contact.Notes,
		CreatedAt:             contact.CreatedAt,
		UpdatedAt:             contact.UpdatedAt,
		DaysInStatus:          contact.DaysInStatus(),
		IsHighPriority:        contact.IsHighPriority(),
		IsHotLead:             contact.IsHotLead(),
	}

	// Include contact type and source if loaded
	if contact.ContactType != nil {
		response.ContactType = &models.ContactTypeResponse{
			ID:          contact.ContactType.ID,
			Name:        contact.ContactType.Name,
			Description: contact.ContactType.Description,
			Color:       contact.ContactType.Color,
			Icon:        contact.ContactType.Icon,
			IsActive:    contact.ContactType.IsActive,
			SortOrder:   contact.ContactType.SortOrder,
			CreatedAt:   contact.ContactType.CreatedAt,
			UpdatedAt:   contact.ContactType.UpdatedAt,
		}
	}

	if contact.ContactSource != nil {
		response.ContactSource = &models.ContactSourceResponse{
			ID:             contact.ContactSource.ID,
			Name:           contact.ContactSource.Name,
			Description:    contact.ContactSource.Description,
			UTMSource:      contact.ContactSource.UTMSource,
			UTMMedium:      contact.ContactSource.UTMMedium,
			UTMCampaign:    contact.ContactSource.UTMCampaign,
			ConversionRate: contact.ContactSource.ConversionRate,
			CostPerLead:    contact.ContactSource.CostPerLead,
			IsActive:       contact.ContactSource.IsActive,
			SortOrder:      contact.ContactSource.SortOrder,
			CreatedAt:      contact.ContactSource.CreatedAt,
			UpdatedAt:      contact.ContactSource.UpdatedAt,
		}
	}

	return response
}

// Request/Response types

type StatusUpdateRequest struct {
	Status string `json:"status" binding:"required,oneof=new contacted qualified proposal negotiation closed_won closed_lost on_hold nurturing"`
}

// Public contact submission (no authentication required)

// SubmitContact godoc
// @Summary Submit a contact form (public endpoint)
// @Description Submit a contact form from the public website
// @Tags public
// @Accept json
// @Produce json
// @Param contact body models.PublicContactRequest true "Contact form data"
// @Success 201 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /public/contact [post]
func (h *ContactHandler) SubmitContact(c *gin.Context) {
	// Get raw JSON data
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Failed to read request body", err.Error()))
		return
	}
	
	// Try to parse as new PublicContactRequest format first
	var pubReq models.PublicContactRequest
	if err := json.Unmarshal(body, &pubReq); err == nil {
		// Check if required fields for new format are present
		if pubReq.FirstName != "" && pubReq.Email != "" && pubReq.Message != "" {
			h.submitContactNewFormat(c, &pubReq)
			return
		}
	}
	
	// Fall back to old ContactSubmissionRequest format for backward compatibility
	var oldReq models.ContactSubmissionRequest
	if err := json.Unmarshal(body, &oldReq); err != nil {
		logger.Warn("Invalid contact submission - neither new nor old format", map[string]interface{}{
			"error": err.Error(),
			"ip":    c.ClientIP(),
		})
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid contact form data", err.Error()))
		return
	}
	
	// Convert old format to new format
	firstName := oldReq.Name
	var lastName *string
	
	// Parse name into first_name and last_name
	if strings.Contains(oldReq.Name, " ") {
		parts := strings.Fields(oldReq.Name)
		if len(parts) > 1 {
			firstName = parts[0]
			lastNameStr := strings.Join(parts[1:], " ")
			lastName = &lastNameStr
		}
	}
	
	// Convert to PublicContactRequest format
	convertedReq := models.PublicContactRequest{
		FirstName:        firstName,
		LastName:         lastName,
		Email:            oldReq.Email,
		Phone:            oldReq.Phone,
		Subject:          oldReq.Subject,
		Message:          oldReq.Message,
		ContactTypeID:    nil, // Will use default
		ContactSourceID:  nil, // Will use default  
		MarketingConsent: nil, // Will use default
		Website:          oldReq.Website, // Honeypot field
	}
	
	h.submitContactNewFormat(c, &convertedReq)
}

// submitContactNewFormat handles the actual contact submission logic
func (h *ContactHandler) submitContactNewFormat(c *gin.Context, req *models.PublicContactRequest) {
	// Honeypot spam detection
	if req.Website != "" {
		logger.LogSecurityEvent("spam_detected", nil, c.ClientIP(), map[string]interface{}{
			"honeypot": "website_field_filled",
			"email":    req.Email,
		})
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid form submission", ""))
		return
	}

	// Convert to internal request format
	contactReq := &models.ContactRequest{
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		Email:            req.Email,
		Phone:            req.Phone,
		Company:          req.Company,
		Subject:          req.Subject,
		Message:          &req.Message,
		ContactTypeID:    getDefaultContactTypeID(req.ContactTypeID),
		ContactSourceID:  getDefaultContactSourceID(req.ContactSourceID),
		MarketingConsent: req.MarketingConsent,
	}

	start := time.Now()
	contact, err := h.contactService.CreateContact(contactReq, nil) // No authenticated user
	duration := time.Since(start)

	logger.LogAPIRequest(c.Request.Method, c.Request.URL.Path, nil, duration, http.StatusCreated)

	if err != nil {
		// Check if this is a duplicate email error
		if strings.Contains(err.Error(), "already exists") {
			logger.Warn("Duplicate contact submission", map[string]interface{}{
				"email": req.Email,
				"ip":    c.ClientIP(),
				"error": err.Error(),
			})
			// Return success response for user experience, but log the duplicate
			c.JSON(http.StatusCreated, NewSuccessResponse("Contact form submitted successfully", gin.H{
				"contact_id": 0, // Indicate this was a duplicate
				"message":    "Thank you for contacting us. We already have your information and will get back to you soon!",
				"duplicate":  true,
			}))
			return
		}
		
		logger.Error("Failed to create public contact", err, map[string]interface{}{
			"email": req.Email,
			"ip":    c.ClientIP(),
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to submit contact form", ""))
		return
	}

	logger.LogBusinessEvent("public_contact_submitted", "contact", contact.ID, map[string]interface{}{
		"email":    req.Email,
		"company":  req.Company,
		"source":   "public_form",
		"ip":       c.ClientIP(),
	})

	c.JSON(http.StatusCreated, NewSuccessResponse("Contact form submitted successfully", gin.H{
		"contact_id": contact.ID,
		"message":    "Thank you for contacting us. We'll get back to you soon!",
	}))
}

// Helper functions for public endpoints

func getDefaultContactTypeID(provided *uint) uint {
	if provided != nil && *provided > 0 {
		return *provided
	}
	return 1 // General Inquiry
}

func getDefaultContactSourceID(provided *uint) uint {
	if provided != nil && *provided > 0 {
		return *provided
	}
	return 1 // Website Contact Form
}

func getUserIDFromContext(c *gin.Context) *uint {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uint); ok {
			return &id
		}
	}
	return nil
}