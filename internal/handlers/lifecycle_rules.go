package handlers

import (
	"contact-service/internal/models"
	"contact-service/pkg/database"
	"contact-service/pkg/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LifecycleRulesHandler handles lifecycle rules management requests
type LifecycleRulesHandler struct {
	db *gorm.DB
}

// NewLifecycleRulesHandler creates a new lifecycle rules handler
func NewLifecycleRulesHandler() *LifecycleRulesHandler {
	return &LifecycleRulesHandler{
		db: database.DB,
	}
}

// Lead Scoring Rules

// CreateScoringRule godoc
// @Summary Create lead scoring rule
// @Description Create a new lead scoring rule
// @Tags lifecycle-rules
// @Accept json
// @Produce json
// @Param rule body models.LeadScoringRuleRequest true "Lead scoring rule data"
// @Success 201 {object} APIResponse{data=models.LeadScoringRule}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /lifecycle/scoring-rules [post]
func (h *LifecycleRulesHandler) CreateScoringRule(c *gin.Context) {
	var req models.LeadScoringRuleRequest
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

	// Create scoring rule
	rule := &models.LeadScoringRule{
		Name:           req.Name,
		Description:    req.Description,
		IsActive:       true,
		Priority:       0,
		Category:       req.Category,
		BaseScore:      0,
		MaxScore:       100,
		Criteria:       req.Criteria,
		ApplicableWhen: req.ApplicableWhen,
		CreatedBy:      userID,
	}

	// Set optional fields
	if req.IsActive != nil {
		rule.IsActive = *req.IsActive
	}
	if req.Priority != nil {
		rule.Priority = *req.Priority
	}
	if req.BaseScore != nil {
		rule.BaseScore = *req.BaseScore
	}
	if req.MaxScore != nil {
		rule.MaxScore = *req.MaxScore
	}

	if err := h.db.Create(rule).Error; err != nil {
		logger.Error("Failed to create lead scoring rule", err, map[string]interface{}{
			"rule_name": req.Name,
			"user_id":   *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to create scoring rule", err.Error()))
		return
	}

	logger.Info("Lead scoring rule created", map[string]interface{}{
		"rule_id":   rule.ID,
		"rule_name": rule.Name,
		"user_id":   *userID,
	})

	c.JSON(http.StatusCreated, NewSuccessResponse("Lead scoring rule created successfully", rule))
}

// GetScoringRules godoc
// @Summary Get lead scoring rules
// @Description Get all lead scoring rules with pagination and filtering
// @Tags lifecycle-rules
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param active query bool false "Filter by active status"
// @Param category query string false "Filter by category"
// @Success 200 {object} APIResponse{data=PaginatedResponse{items=[]models.LeadScoringRule}}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /lifecycle/scoring-rules [get]
func (h *LifecycleRulesHandler) GetScoringRules(c *gin.Context) {
	// Parse pagination parameters
	page, limit := parsePaginationParams(c)
	offset := (page - 1) * limit

	// Parse filters
	activeFilter := c.Query("active")
	category := c.Query("category")

	// Build query
	query := h.db.Where("deleted_at IS NULL")
	if activeFilter != "" {
		if activeFilter == "true" {
			query = query.Where("is_active = ?", true)
		} else if activeFilter == "false" {
			query = query.Where("is_active = ?", false)
		}
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// Get total count
	var total int64
	if err := query.Model(&models.LeadScoringRule{}).Count(&total).Error; err != nil {
		logger.Error("Failed to count lead scoring rules", err, nil)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get scoring rules", err.Error()))
		return
	}

	// Get rules
	var rules []models.LeadScoringRule
	if err := query.Order("priority DESC, created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&rules).Error; err != nil {
		logger.Error("Failed to get lead scoring rules", err, nil)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get scoring rules", err.Error()))
		return
	}

	response := NewPaginatedResponseWithItems(rules, int(total), page, limit)
	c.JSON(http.StatusOK, NewSuccessResponse("Lead scoring rules retrieved successfully", response))
}

// GetScoringRule godoc
// @Summary Get lead scoring rule
// @Description Get a specific lead scoring rule by ID
// @Tags lifecycle-rules
// @Accept json
// @Produce json
// @Param id path int true "Rule ID"
// @Success 200 {object} APIResponse{data=models.LeadScoringRule}
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /lifecycle/scoring-rules/{id} [get]
func (h *LifecycleRulesHandler) GetScoringRule(c *gin.Context) {
	ruleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid rule ID", err.Error()))
		return
	}

	var rule models.LeadScoringRule
	if err := h.db.Where("deleted_at IS NULL").First(&rule, ruleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, NewErrorResponse("Lead scoring rule not found", ""))
			return
		}
		logger.Error("Failed to get lead scoring rule", err, map[string]interface{}{
			"rule_id": ruleID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get scoring rule", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Lead scoring rule retrieved successfully", rule))
}

// UpdateScoringRule godoc
// @Summary Update lead scoring rule
// @Description Update an existing lead scoring rule
// @Tags lifecycle-rules
// @Accept json
// @Produce json
// @Param id path int true "Rule ID"
// @Param rule body models.LeadScoringRuleRequest true "Updated scoring rule data"
// @Success 200 {object} APIResponse{data=models.LeadScoringRule}
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /lifecycle/scoring-rules/{id} [put]
func (h *LifecycleRulesHandler) UpdateScoringRule(c *gin.Context) {
	ruleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid rule ID", err.Error()))
		return
	}

	var req models.LeadScoringRuleRequest
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

	// Get existing rule
	var rule models.LeadScoringRule
	if err := h.db.Where("deleted_at IS NULL").First(&rule, ruleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, NewErrorResponse("Lead scoring rule not found", ""))
			return
		}
		logger.Error("Failed to get lead scoring rule for update", err, map[string]interface{}{
			"rule_id": ruleID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get scoring rule", err.Error()))
		return
	}

	// Update fields
	updates := map[string]interface{}{
		"name":             req.Name,
		"description":      req.Description,
		"category":         req.Category,
		"criteria":         req.Criteria,
		"applicable_when":  req.ApplicableWhen,
		"updated_by":       *userID,
	}

	// Set optional fields
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.BaseScore != nil {
		updates["base_score"] = *req.BaseScore
	}
	if req.MaxScore != nil {
		updates["max_score"] = *req.MaxScore
	}

	if err := h.db.Model(&rule).Updates(updates).Error; err != nil {
		logger.Error("Failed to update lead scoring rule", err, map[string]interface{}{
			"rule_id": ruleID,
			"user_id": *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to update scoring rule", err.Error()))
		return
	}

	// Reload updated rule
	h.db.First(&rule, ruleID)

	logger.Info("Lead scoring rule updated", map[string]interface{}{
		"rule_id":   rule.ID,
		"rule_name": rule.Name,
		"user_id":   *userID,
	})

	c.JSON(http.StatusOK, NewSuccessResponse("Lead scoring rule updated successfully", rule))
}

// DeleteScoringRule godoc
// @Summary Delete lead scoring rule
// @Description Soft delete a lead scoring rule
// @Tags lifecycle-rules
// @Accept json
// @Produce json
// @Param id path int true "Rule ID"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /lifecycle/scoring-rules/{id} [delete]
func (h *LifecycleRulesHandler) DeleteScoringRule(c *gin.Context) {
	ruleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid rule ID", err.Error()))
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	// Check if rule exists
	var rule models.LeadScoringRule
	if err := h.db.Where("deleted_at IS NULL").First(&rule, ruleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, NewErrorResponse("Lead scoring rule not found", ""))
			return
		}
		logger.Error("Failed to get lead scoring rule for deletion", err, map[string]interface{}{
			"rule_id": ruleID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get scoring rule", err.Error()))
		return
	}

	// Soft delete the rule
	if err := h.db.Model(&rule).Updates(map[string]interface{}{
		"updated_by": *userID,
	}).Delete(&rule).Error; err != nil {
		logger.Error("Failed to delete lead scoring rule", err, map[string]interface{}{
			"rule_id": ruleID,
			"user_id": *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to delete scoring rule", err.Error()))
		return
	}

	logger.Info("Lead scoring rule deleted", map[string]interface{}{
		"rule_id":   rule.ID,
		"rule_name": rule.Name,
		"user_id":   *userID,
	})

	c.JSON(http.StatusOK, NewSuccessResponse("Lead scoring rule deleted successfully", nil))
}

// Status Transition Rules

// CreateTransitionRule godoc
// @Summary Create status transition rule
// @Description Create a new status transition rule
// @Tags lifecycle-rules
// @Accept json
// @Produce json
// @Param rule body models.StatusTransitionRuleRequest true "Status transition rule data"
// @Success 201 {object} APIResponse{data=models.StatusTransitionRule}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /lifecycle/transition-rules [post]
func (h *LifecycleRulesHandler) CreateTransitionRule(c *gin.Context) {
	var req models.StatusTransitionRuleRequest
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

	// Create transition rule
	rule := &models.StatusTransitionRule{
		Name:           req.Name,
		Description:    req.Description,
		IsActive:       true,
		Priority:       0,
		FromStatus:     req.FromStatus,
		ToStatus:       req.ToStatus,
		TransitionType: models.TransitionAutomatic,
		Conditions:     req.Conditions,
		RequiredScore:  0,
		DaysInStatus:   0,
		Actions:        req.Actions,
		NotifyUsers:    req.NotifyUsers,
		CreatedBy:      userID,
	}

	// Set optional fields
	if req.IsActive != nil {
		rule.IsActive = *req.IsActive
	}
	if req.Priority != nil {
		rule.Priority = *req.Priority
	}
	if req.TransitionType != nil {
		rule.TransitionType = *req.TransitionType
	}
	if req.RequiredScore != nil {
		rule.RequiredScore = *req.RequiredScore
	}
	if req.DaysInStatus != nil {
		rule.DaysInStatus = *req.DaysInStatus
	}

	if err := h.db.Create(rule).Error; err != nil {
		logger.Error("Failed to create status transition rule", err, map[string]interface{}{
			"rule_name": req.Name,
			"user_id":   *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to create transition rule", err.Error()))
		return
	}

	logger.Info("Status transition rule created", map[string]interface{}{
		"rule_id":   rule.ID,
		"rule_name": rule.Name,
		"user_id":   *userID,
	})

	c.JSON(http.StatusCreated, NewSuccessResponse("Status transition rule created successfully", rule))
}

// GetTransitionRules godoc
// @Summary Get status transition rules
// @Description Get all status transition rules with pagination and filtering
// @Tags lifecycle-rules
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param active query bool false "Filter by active status"
// @Param from_status query string false "Filter by from status"
// @Param to_status query string false "Filter by to status"
// @Success 200 {object} APIResponse{data=PaginatedResponse{items=[]models.StatusTransitionRule}}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /lifecycle/transition-rules [get]
func (h *LifecycleRulesHandler) GetTransitionRules(c *gin.Context) {
	// Parse pagination parameters
	page, limit := parsePaginationParams(c)
	offset := (page - 1) * limit

	// Parse filters
	activeFilter := c.Query("active")
	fromStatus := c.Query("from_status")
	toStatus := c.Query("to_status")

	// Build query
	query := h.db.Where("deleted_at IS NULL")
	if activeFilter != "" {
		if activeFilter == "true" {
			query = query.Where("is_active = ?", true)
		} else if activeFilter == "false" {
			query = query.Where("is_active = ?", false)
		}
	}
	if fromStatus != "" {
		query = query.Where("from_status = ?", fromStatus)
	}
	if toStatus != "" {
		query = query.Where("to_status = ?", toStatus)
	}

	// Get total count
	var total int64
	if err := query.Model(&models.StatusTransitionRule{}).Count(&total).Error; err != nil {
		logger.Error("Failed to count status transition rules", err, nil)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get transition rules", err.Error()))
		return
	}

	// Get rules
	var rules []models.StatusTransitionRule
	if err := query.Order("priority DESC, created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&rules).Error; err != nil {
		logger.Error("Failed to get status transition rules", err, nil)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get transition rules", err.Error()))
		return
	}

	response := NewPaginatedResponseWithItems(rules, int(total), page, limit)
	c.JSON(http.StatusOK, NewSuccessResponse("Status transition rules retrieved successfully", response))
}

// GetTransitionRule godoc
// @Summary Get status transition rule
// @Description Get a specific status transition rule by ID
// @Tags lifecycle-rules
// @Accept json
// @Produce json
// @Param id path int true "Rule ID"
// @Success 200 {object} APIResponse{data=models.StatusTransitionRule}
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /lifecycle/transition-rules/{id} [get]
func (h *LifecycleRulesHandler) GetTransitionRule(c *gin.Context) {
	ruleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid rule ID", err.Error()))
		return
	}

	var rule models.StatusTransitionRule
	if err := h.db.Where("deleted_at IS NULL").First(&rule, ruleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, NewErrorResponse("Status transition rule not found", ""))
			return
		}
		logger.Error("Failed to get status transition rule", err, map[string]interface{}{
			"rule_id": ruleID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get transition rule", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Status transition rule retrieved successfully", rule))
}

// UpdateTransitionRule godoc
// @Summary Update status transition rule
// @Description Update an existing status transition rule
// @Tags lifecycle-rules
// @Accept json
// @Produce json
// @Param id path int true "Rule ID"
// @Param rule body models.StatusTransitionRuleRequest true "Updated transition rule data"
// @Success 200 {object} APIResponse{data=models.StatusTransitionRule}
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /lifecycle/transition-rules/{id} [put]
func (h *LifecycleRulesHandler) UpdateTransitionRule(c *gin.Context) {
	ruleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid rule ID", err.Error()))
		return
	}

	var req models.StatusTransitionRuleRequest
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

	// Get existing rule
	var rule models.StatusTransitionRule
	if err := h.db.Where("deleted_at IS NULL").First(&rule, ruleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, NewErrorResponse("Status transition rule not found", ""))
			return
		}
		logger.Error("Failed to get status transition rule for update", err, map[string]interface{}{
			"rule_id": ruleID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get transition rule", err.Error()))
		return
	}

	// Update fields
	updates := map[string]interface{}{
		"name":         req.Name,
		"description":  req.Description,
		"from_status":  req.FromStatus,
		"to_status":    req.ToStatus,
		"conditions":   req.Conditions,
		"actions":      req.Actions,
		"notify_users": req.NotifyUsers,
		"updated_by":   *userID,
	}

	// Set optional fields
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.TransitionType != nil {
		updates["transition_type"] = *req.TransitionType
	}
	if req.RequiredScore != nil {
		updates["required_score"] = *req.RequiredScore
	}
	if req.DaysInStatus != nil {
		updates["days_in_status"] = *req.DaysInStatus
	}

	if err := h.db.Model(&rule).Updates(updates).Error; err != nil {
		logger.Error("Failed to update status transition rule", err, map[string]interface{}{
			"rule_id": ruleID,
			"user_id": *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to update transition rule", err.Error()))
		return
	}

	// Reload updated rule
	h.db.First(&rule, ruleID)

	logger.Info("Status transition rule updated", map[string]interface{}{
		"rule_id":   rule.ID,
		"rule_name": rule.Name,
		"user_id":   *userID,
	})

	c.JSON(http.StatusOK, NewSuccessResponse("Status transition rule updated successfully", rule))
}

// DeleteTransitionRule godoc
// @Summary Delete status transition rule
// @Description Soft delete a status transition rule
// @Tags lifecycle-rules
// @Accept json
// @Produce json
// @Param id path int true "Rule ID"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /lifecycle/transition-rules/{id} [delete]
func (h *LifecycleRulesHandler) DeleteTransitionRule(c *gin.Context) {
	ruleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid rule ID", err.Error()))
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	// Check if rule exists
	var rule models.StatusTransitionRule
	if err := h.db.Where("deleted_at IS NULL").First(&rule, ruleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, NewErrorResponse("Status transition rule not found", ""))
			return
		}
		logger.Error("Failed to get status transition rule for deletion", err, map[string]interface{}{
			"rule_id": ruleID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get transition rule", err.Error()))
		return
	}

	// Soft delete the rule
	if err := h.db.Model(&rule).Updates(map[string]interface{}{
		"updated_by": *userID,
	}).Delete(&rule).Error; err != nil {
		logger.Error("Failed to delete status transition rule", err, map[string]interface{}{
			"rule_id": ruleID,
			"user_id": *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to delete transition rule", err.Error()))
		return
	}

	logger.Info("Status transition rule deleted", map[string]interface{}{
		"rule_id":   rule.ID,
		"rule_name": rule.Name,
		"user_id":   *userID,
	})

	c.JSON(http.StatusOK, NewSuccessResponse("Status transition rule deleted successfully", nil))
}