package handlers

import (
	"contact-service/internal/models"
	"contact-service/internal/services"
	"contact-service/pkg/logger"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SearchHandler handles advanced search and filtering requests
type SearchHandler struct {
	contactService *services.ContactService
}

// NewSearchHandler creates a new search handler
func NewSearchHandler() *SearchHandler {
	return &SearchHandler{
		contactService: services.NewContactService(),
	}
}

// AdvancedSearch godoc
// @Summary Advanced contact search with multiple criteria
// @Description Perform advanced search across all contact fields with flexible filtering
// @Tags search
// @Accept json
// @Produce json
// @Param q query string false "Full-text search query"
// @Param first_name query string false "Search by first name"
// @Param last_name query string false "Search by last name"
// @Param email query string false "Search by email"
// @Param phone query string false "Search by phone"
// @Param company query string false "Search by company"
// @Param job_title query string false "Search by job title"
// @Param country query string false "Filter by country"
// @Param status query string false "Filter by status"
// @Param priority query string false "Filter by priority"
// @Param lead_score_min query int false "Minimum lead score"
// @Param lead_score_max query int false "Maximum lead score"
// @Param estimated_value_min query float64 false "Minimum estimated value"
// @Param estimated_value_max query float64 false "Maximum estimated value"
// @Param assigned_to query int false "Filter by assigned user"
// @Param source_id query int false "Filter by contact source"
// @Param type_id query int false "Filter by contact type"
// @Param tags query string false "Filter by tags (comma-separated)"
// @Param created_from query string false "Filter from creation date (YYYY-MM-DD)"
// @Param created_to query string false "Filter to creation date (YYYY-MM-DD)"
// @Param last_contact_from query string false "Filter from last contact date (YYYY-MM-DD)"
// @Param last_contact_to query string false "Filter to last contact date (YYYY-MM-DD)"
// @Param has_activities query bool false "Filter contacts with activities"
// @Param is_hot_lead query bool false "Filter hot leads only"
// @Param is_high_priority query bool false "Filter high priority contacts only"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param sort_by query string false "Sort field" default(created_at)
// @Param sort_order query string false "Sort order (ASC/DESC)" default(DESC)
// @Success 200 {object} APIResponse{data=[]models.ContactResponse,meta=PaginationMeta}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /search/contacts/advanced [get]
func (h *SearchHandler) AdvancedSearch(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Build search criteria
	criteria := &services.AdvancedSearchCriteria{
		Page:     page,
		PageSize: pageSize,
		SortBy:   c.DefaultQuery("sort_by", "created_at"),
		SortOrder: c.DefaultQuery("sort_order", "DESC"),
	}

	// Text search fields
	if q := c.Query("q"); q != "" {
		criteria.FullTextSearch = &q
	}
	if firstName := c.Query("first_name"); firstName != "" {
		criteria.FirstName = &firstName
	}
	if lastName := c.Query("last_name"); lastName != "" {
		criteria.LastName = &lastName
	}
	if email := c.Query("email"); email != "" {
		criteria.Email = &email
	}
	if phone := c.Query("phone"); phone != "" {
		criteria.Phone = &phone
	}
	if company := c.Query("company"); company != "" {
		criteria.Company = &company
	}
	if jobTitle := c.Query("job_title"); jobTitle != "" {
		criteria.JobTitle = &jobTitle
	}
	if country := c.Query("country"); country != "" {
		criteria.Country = &country
	}

	// Status and priority filters
	if status := c.Query("status"); status != "" {
		criteria.Status = &status
	}
	if priority := c.Query("priority"); priority != "" {
		criteria.Priority = &priority
	}

	// Numeric filters
	if scoreMin := c.Query("lead_score_min"); scoreMin != "" {
		if score, err := strconv.Atoi(scoreMin); err == nil {
			criteria.LeadScoreMin = &score
		}
	}
	if scoreMax := c.Query("lead_score_max"); scoreMax != "" {
		if score, err := strconv.Atoi(scoreMax); err == nil {
			criteria.LeadScoreMax = &score
		}
	}
	if valueMin := c.Query("estimated_value_min"); valueMin != "" {
		if value, err := strconv.ParseFloat(valueMin, 64); err == nil {
			criteria.EstimatedValueMin = &value
		}
	}
	if valueMax := c.Query("estimated_value_max"); valueMax != "" {
		if value, err := strconv.ParseFloat(valueMax, 64); err == nil {
			criteria.EstimatedValueMax = &value
		}
	}

	// Reference filters
	if assignedTo := c.Query("assigned_to"); assignedTo != "" {
		if id, err := strconv.ParseUint(assignedTo, 10, 32); err == nil {
			assignedToUint := uint(id)
			criteria.AssignedTo = &assignedToUint
		}
	}
	if sourceID := c.Query("source_id"); sourceID != "" {
		if id, err := strconv.ParseUint(sourceID, 10, 32); err == nil {
			sourceIDUint := uint(id)
			criteria.SourceID = &sourceIDUint
		}
	}
	if typeID := c.Query("type_id"); typeID != "" {
		if id, err := strconv.ParseUint(typeID, 10, 32); err == nil {
			typeIDUint := uint(id)
			criteria.TypeID = &typeIDUint
		}
	}

	// Date filters
	if createdFrom := c.Query("created_from"); createdFrom != "" {
		if date, err := time.Parse("2006-01-02", createdFrom); err == nil {
			criteria.CreatedFrom = &date
		}
	}
	if createdTo := c.Query("created_to"); createdTo != "" {
		if date, err := time.Parse("2006-01-02", createdTo); err == nil {
			endOfDay := date.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			criteria.CreatedTo = &endOfDay
		}
	}
	if lastContactFrom := c.Query("last_contact_from"); lastContactFrom != "" {
		if date, err := time.Parse("2006-01-02", lastContactFrom); err == nil {
			criteria.LastContactFrom = &date
		}
	}
	if lastContactTo := c.Query("last_contact_to"); lastContactTo != "" {
		if date, err := time.Parse("2006-01-02", lastContactTo); err == nil {
			endOfDay := date.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			criteria.LastContactTo = &endOfDay
		}
	}

	// Boolean filters
	if hasActivities := c.Query("has_activities"); hasActivities != "" {
		if value, err := strconv.ParseBool(hasActivities); err == nil {
			criteria.HasActivities = &value
		}
	}
	if isHotLead := c.Query("is_hot_lead"); isHotLead != "" {
		if value, err := strconv.ParseBool(isHotLead); err == nil {
			criteria.IsHotLead = &value
		}
	}
	if isHighPriority := c.Query("is_high_priority"); isHighPriority != "" {
		if value, err := strconv.ParseBool(isHighPriority); err == nil {
			criteria.IsHighPriority = &value
		}
	}

	// Tags filter
	if tags := c.Query("tags"); tags != "" {
		criteria.Tags = strings.Split(tags, ",")
	}

	userID := getUserIDFromContext(c)
	start := time.Now()
	contacts, total, err := h.contactService.AdvancedSearch(criteria)
	duration := time.Since(start)

	logger.LogAPIRequest(c.Request.Method, c.Request.URL.Path, userID, duration, http.StatusOK)

	if err != nil {
		logger.Error("Advanced search failed", err, map[string]interface{}{
			"criteria": criteria,
			"user_id":  userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Search failed", ""))
		return
	}

	// Map to response format
	contactHandler := NewContactHandler()
	responses := make([]*models.ContactResponse, len(contacts))
	for i, contact := range contacts {
		responses[i] = contactHandler.mapContactToResponse(contact)
	}

	// Create pagination metadata
	meta := NewPaginationMeta(page, pageSize, total)

	c.JSON(http.StatusOK, NewPaginatedResponse("Advanced search completed", responses, meta))
}

// SearchSuggestions godoc
// @Summary Get search suggestions for autocomplete
// @Description Get suggestions for contact fields to assist with search autocomplete
// @Tags search
// @Accept json
// @Produce json
// @Param field query string true "Field to get suggestions for (email, company, job_title, country)"
// @Param q query string true "Search query for suggestions"
// @Param limit query int false "Maximum number of suggestions" default(10)
// @Success 200 {object} APIResponse{data=[]string}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /search/suggestions [get]
func (h *SearchHandler) SearchSuggestions(c *gin.Context) {
	field := c.Query("field")
	query := c.Query("q")
	limitStr := c.DefaultQuery("limit", "10")

	if field == "" || query == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Field and query parameters are required", ""))
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}

	// Validate field
	allowedFields := map[string]bool{
		"email":     true,
		"company":   true,
		"job_title": true,
		"country":   true,
	}

	if !allowedFields[field] {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid field", "Allowed fields: email, company, job_title, country"))
		return
	}

	userID := getUserIDFromContext(c)
	start := time.Now()
	suggestions, err := h.contactService.GetSearchSuggestions(field, query, limit)
	duration := time.Since(start)

	logger.LogAPIRequest(c.Request.Method, c.Request.URL.Path, userID, duration, http.StatusOK)

	if err != nil {
		logger.Error("Failed to get search suggestions", err, map[string]interface{}{
			"field":   field,
			"query":   query,
			"user_id": userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get suggestions", ""))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Suggestions retrieved", suggestions))
}

// SavedSearches godoc
// @Summary Get user's saved searches
// @Description Retrieve all saved searches for the current user
// @Tags search
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=[]models.SavedSearchResponse}
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /search/saved [get]
func (h *SearchHandler) SavedSearches(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	start := time.Now()
	searches, err := h.contactService.GetSavedSearches(*userID)
	duration := time.Since(start)

	logger.LogAPIRequest(c.Request.Method, c.Request.URL.Path, userID, duration, http.StatusOK)

	if err != nil {
		logger.Error("Failed to get saved searches", err, map[string]interface{}{
			"user_id": *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to retrieve saved searches", ""))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Saved searches retrieved", searches))
}

// SaveSearch godoc
// @Summary Save a search query
// @Description Save a search query for later use
// @Tags search
// @Accept json
// @Produce json
// @Param search body models.SavedSearchRequest true "Search to save"
// @Success 201 {object} APIResponse{data=models.SavedSearchResponse}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /search/saved [post]
func (h *SearchHandler) SaveSearch(c *gin.Context) {
	var req models.SavedSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request data", err.Error()))
		return
	}

	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	start := time.Now()
	savedSearch, err := h.contactService.SaveSearch(*userID, &req)
	duration := time.Since(start)

	logger.LogAPIRequest(c.Request.Method, c.Request.URL.Path, userID, duration, http.StatusCreated)

	if err != nil {
		logger.Error("Failed to save search", err, map[string]interface{}{
			"name":    req.Name,
			"user_id": *userID,
		})

		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, NewConflictResponse("Search name already exists"))
			return
		}

		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to save search", ""))
		return
	}

	c.JSON(http.StatusCreated, NewSuccessResponse("Search saved successfully", savedSearch))
}

// DeleteSavedSearch godoc
// @Summary Delete a saved search
// @Description Delete a saved search by ID
// @Tags search
// @Accept json
// @Produce json
// @Param id path int true "Saved search ID"
// @Success 200 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /search/saved/{id} [delete]
func (h *SearchHandler) DeleteSavedSearch(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid search ID", ""))
		return
	}

	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	start := time.Now()
	err = h.contactService.DeleteSavedSearch(uint(id), *userID)
	duration := time.Since(start)

	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}

		logger.LogAPIRequest(c.Request.Method, c.Request.URL.Path, userID, duration, status)
		c.JSON(status, NewErrorResponse("Failed to delete saved search", ""))
		return
	}

	logger.LogAPIRequest(c.Request.Method, c.Request.URL.Path, userID, duration, http.StatusOK)
	c.JSON(http.StatusOK, NewSuccessResponse("Saved search deleted successfully", nil))
}

// ExecuteSavedSearch godoc
// @Summary Execute a saved search
// @Description Execute a previously saved search query
// @Tags search
// @Accept json
// @Produce json
// @Param id path int true "Saved search ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} APIResponse{data=[]models.ContactResponse,meta=PaginationMeta}
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /search/saved/{id}/execute [get]
func (h *SearchHandler) ExecuteSavedSearch(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid search ID", ""))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	start := time.Now()
	contacts, total, err := h.contactService.ExecuteSavedSearch(uint(id), *userID, page, pageSize)
	duration := time.Since(start)

	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}

		logger.LogAPIRequest(c.Request.Method, c.Request.URL.Path, userID, duration, status)
		c.JSON(status, NewErrorResponse("Failed to execute saved search", ""))
		return
	}

	logger.LogAPIRequest(c.Request.Method, c.Request.URL.Path, userID, duration, http.StatusOK)

	// Map to response format
	contactHandler := NewContactHandler()
	responses := make([]*models.ContactResponse, len(contacts))
	for i, contact := range contacts {
		responses[i] = contactHandler.mapContactToResponse(contact)
	}

	// Create pagination metadata
	meta := NewPaginationMeta(page, pageSize, total)

	c.JSON(http.StatusOK, NewPaginatedResponse("Saved search executed successfully", responses, meta))
}

// Request/Response types are now in the models package