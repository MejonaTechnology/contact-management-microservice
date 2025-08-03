package handlers

import (
	"contact-service/internal/models"
	"contact-service/internal/services"
	"contact-service/pkg/database"
	"contact-service/pkg/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// LifecycleHandler handles contact lifecycle management requests
type LifecycleHandler struct {
	lifecycleService *services.LifecycleService
}

// NewLifecycleHandler creates a new lifecycle handler
func NewLifecycleHandler() *LifecycleHandler {
	return &LifecycleHandler{
		lifecycleService: services.NewLifecycleService(database.DB),
	}
}

// ScoreContact godoc
// @Summary Score a contact
// @Description Calculate and update lead score for a contact
// @Tags lifecycle
// @Accept json
// @Produce json
// @Param request body models.ContactScoringRequest true "Scoring request"
// @Success 200 {object} APIResponse{data=models.ContactLifecycleResponse}
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /lifecycle/score [post]
func (h *LifecycleHandler) ScoreContact(c *gin.Context) {
	var req models.ContactScoringRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request data", err.Error()))
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := getUserIDFromContext(c)

	result, err := h.lifecycleService.ScoreContact(req.ContactID, req.ForceRescore, req.Reason, userID)
	if err != nil {
		logger.Error("Failed to score contact", err, map[string]interface{}{
			"contact_id": req.ContactID,
			"user_id":    userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to score contact", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Contact scored successfully", result))
}

// ScoreContactByID godoc
// @Summary Score a contact by ID
// @Description Calculate and update lead score for a contact using URL parameter
// @Tags lifecycle
// @Accept json
// @Produce json
// @Param id path int true "Contact ID"
// @Param force query bool false "Force rescore even if recently scored"
// @Param reason query string false "Reason for scoring"
// @Success 200 {object} APIResponse{data=models.ContactLifecycleResponse}
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /lifecycle/score/{id} [post]
func (h *LifecycleHandler) ScoreContactByID(c *gin.Context) {
	contactID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid contact ID", err.Error()))
		return
	}

	forceRescore := c.Query("force") == "true"
	reason := c.Query("reason")
	if reason == "" {
		reason = "Manual scoring request"
	}

	// Get user ID from context (set by auth middleware)
	userID := getUserIDFromContext(c)

	result, err := h.lifecycleService.ScoreContact(uint(contactID), forceRescore, reason, userID)
	if err != nil {
		logger.Error("Failed to score contact", err, map[string]interface{}{
			"contact_id": contactID,
			"user_id":    userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to score contact", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Contact scored successfully", result))
}

// ChangeContactStatus godoc
// @Summary Change contact status
// @Description Manually change the status of a contact
// @Tags lifecycle
// @Accept json
// @Produce json
// @Param request body models.StatusChangeRequest true "Status change request"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /lifecycle/status/change [post]
func (h *LifecycleHandler) ChangeContactStatus(c *gin.Context) {
	var req models.StatusChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request data", err.Error()))
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	err := h.lifecycleService.ChangeContactStatus(&req, *userID)
	if err != nil {
		logger.Error("Failed to change contact status", err, map[string]interface{}{
			"contact_id": req.ContactID,
			"new_status": req.NewStatus,
			"user_id":    *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to change contact status", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Contact status changed successfully", nil))
}

// ChangeContactStatusByID godoc
// @Summary Change contact status by ID
// @Description Manually change the status of a contact using URL parameter
// @Tags lifecycle
// @Accept json
// @Produce json
// @Param id path int true "Contact ID"
// @Param request body map[string]interface{} true "Status change data"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /lifecycle/status/{id}/change [post]
func (h *LifecycleHandler) ChangeContactStatusByID(c *gin.Context) {
	contactID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid contact ID", err.Error()))
		return
	}

	var requestData map[string]interface{}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request data", err.Error()))
		return
	}

	// Extract required fields
	newStatusStr, ok := requestData["new_status"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, NewErrorResponse("new_status is required", ""))
		return
	}

	reason, ok := requestData["reason"].(string)
	if !ok || reason == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse("reason is required", ""))
		return
	}

	forceChange := false
	if force, exists := requestData["force_change"]; exists {
		if forceBool, ok := force.(bool); ok {
			forceChange = forceBool
		}
	}

	req := &models.StatusChangeRequest{
		ContactID:   uint(contactID),
		NewStatus:   models.ContactStatus(newStatusStr),
		Reason:      reason,
		ForceChange: forceChange,
	}

	// Get user ID from context (set by auth middleware)
	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	err = h.lifecycleService.ChangeContactStatus(req, *userID)
	if err != nil {
		logger.Error("Failed to change contact status", err, map[string]interface{}{
			"contact_id": contactID,
			"new_status": newStatusStr,
			"user_id":    *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to change contact status", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Contact status changed successfully", nil))
}

// BulkChangeStatus godoc
// @Summary Bulk change contact status
// @Description Change status for multiple contacts at once
// @Tags lifecycle
// @Accept json
// @Produce json
// @Param request body models.BulkStatusChangeRequest true "Bulk status change request"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /lifecycle/status/bulk-change [post]
func (h *LifecycleHandler) BulkChangeStatus(c *gin.Context) {
	var req models.BulkStatusChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request data", err.Error()))
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	err := h.lifecycleService.BulkChangeContactStatus(&req, *userID)
	if err != nil {
		logger.Error("Failed to bulk change contact status", err, map[string]interface{}{
			"contact_count": len(req.ContactIDs),
			"new_status":    req.NewStatus,
			"user_id":       *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to bulk change contact status", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Contact statuses changed successfully", nil))
}

// GetContactLifecycle godoc
// @Summary Get contact lifecycle
// @Description Get lifecycle information for a specific contact
// @Tags lifecycle
// @Accept json
// @Produce json
// @Param id path int true "Contact ID"
// @Success 200 {object} APIResponse{data=models.ContactLifecycleResponse}
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /lifecycle/{id} [get]
func (h *LifecycleHandler) GetContactLifecycle(c *gin.Context) {
	contactID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid contact ID", err.Error()))
		return
	}

	result, err := h.lifecycleService.GetContactLifecycle(uint(contactID))
	if err != nil {
		logger.Error("Failed to get contact lifecycle", err, map[string]interface{}{
			"contact_id": contactID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get contact lifecycle", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Contact lifecycle retrieved successfully", result))
}

// GetLifecycleEvents godoc
// @Summary Get lifecycle events
// @Description Get lifecycle events for a specific contact
// @Tags lifecycle
// @Accept json
// @Produce json
// @Param id path int true "Contact ID"
// @Param limit query int false "Maximum number of events to return" default(50)
// @Success 200 {object} APIResponse{data=[]models.LifecycleEvent}
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /lifecycle/{id}/events [get]
func (h *LifecycleHandler) GetLifecycleEvents(c *gin.Context) {
	contactID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid contact ID", err.Error()))
		return
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
			limit = l
		}
	}

	events, err := h.lifecycleService.GetLifecycleEvents(uint(contactID), limit)
	if err != nil {
		logger.Error("Failed to get lifecycle events", err, map[string]interface{}{
			"contact_id": contactID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get lifecycle events", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Lifecycle events retrieved successfully", events))
}

// AnalyzeScoring godoc
// @Summary Analyze contact scoring
// @Description Get detailed scoring analysis for a contact
// @Tags lifecycle
// @Accept json
// @Produce json
// @Param id path int true "Contact ID"
// @Success 200 {object} APIResponse{data=models.ScoringAnalysisResponse}
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /lifecycle/{id}/analyze [get]
func (h *LifecycleHandler) AnalyzeScoring(c *gin.Context) {
	contactID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid contact ID", err.Error()))
		return
	}

	analysis, err := h.lifecycleService.AnalyzeContactScoring(uint(contactID))
	if err != nil {
		logger.Error("Failed to analyze contact scoring", err, map[string]interface{}{
			"contact_id": contactID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to analyze contact scoring", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Scoring analysis completed successfully", analysis))
}