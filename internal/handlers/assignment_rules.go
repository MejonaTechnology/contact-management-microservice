package handlers

import (
	"contact-service/internal/models"
	"contact-service/internal/services"
	"contact-service/pkg/database"
	"contact-service/pkg/logger"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AssignmentRuleHandler handles assignment rule requests
type AssignmentRuleHandler struct {
	db *gorm.DB
}

// NewAssignmentRuleHandler creates a new assignment rule handler
func NewAssignmentRuleHandler() *AssignmentRuleHandler {
	return &AssignmentRuleHandler{
		db: database.DB,
	}
}

// CreateAssignmentRule godoc
// @Summary Create assignment rule
// @Description Create a new contact assignment rule
// @Tags assignment-rules
// @Accept json
// @Produce json
// @Param rule body models.AssignmentRuleRequest true "Assignment rule data"
// @Success 201 {object} APIResponse{data=models.AssignmentRule}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /assignment-rules [post]
func (h *AssignmentRuleHandler) CreateAssignmentRule(c *gin.Context) {
	var req models.AssignmentRuleRequest
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

	// Create assignment rule
	rule := &models.AssignmentRule{
		Name:                 req.Name,
		Description:          req.Description,
		Type:                 req.Type,
		Status:               models.AssignmentRuleActive,
		Priority:             0,
		Conditions:           req.Conditions,
		Settings:             req.Settings,
		AssigneeIDs:          req.AssigneeIDs,
		FallbackUserID:       req.FallbackUserID,
		BusinessHoursEnabled: false,
		BusinessHoursStart:   req.BusinessHoursStart,
		BusinessHoursEnd:     req.BusinessHoursEnd,
		WorkingDays:          req.WorkingDays,
		Timezone:             "UTC",
		MaxAssignmentsPerHour: req.MaxAssignmentsPerHour,
		MaxAssignmentsPerDay:  req.MaxAssignmentsPerDay,
		CreatedBy:            userID,
	}

	// Set optional fields
	if req.Status != nil {
		rule.Status = *req.Status
	}
	if req.Priority != nil {
		rule.Priority = *req.Priority
	}
	if req.BusinessHoursEnabled != nil {
		rule.BusinessHoursEnabled = *req.BusinessHoursEnabled
	}
	if req.Timezone != nil {
		rule.Timezone = *req.Timezone
	}

	if err := h.db.Create(rule).Error; err != nil {
		logger.Error("Failed to create assignment rule", err, map[string]interface{}{
			"rule_name": req.Name,
			"user_id":   *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to create assignment rule", err.Error()))
		return
	}

	logger.Info("Assignment rule created", map[string]interface{}{
		"rule_id":   rule.ID,
		"rule_name": rule.Name,
		"user_id":   *userID,
	})

	c.JSON(http.StatusCreated, NewSuccessResponse("Assignment rule created successfully", rule))
}

// GetAssignmentRules godoc
// @Summary Get assignment rules
// @Description Get all assignment rules with pagination and filtering
// @Tags assignment-rules
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param status query string false "Status filter (active, inactive, paused)"
// @Param type query string false "Type filter"
// @Success 200 {object} APIResponse{data=PaginatedResponse{items=[]models.AssignmentRule}}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /assignment-rules [get]
func (h *AssignmentRuleHandler) GetAssignmentRules(c *gin.Context) {
	// Parse pagination parameters
	page, limit := parsePaginationParams(c)
	offset := (page - 1) * limit

	// Parse filters
	status := c.Query("status")
	ruleType := c.Query("type")

	// Build query
	query := h.db.Where("deleted_at IS NULL")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if ruleType != "" {
		query = query.Where("type = ?", ruleType)
	}

	// Get total count
	var total int64
	if err := query.Model(&models.AssignmentRule{}).Count(&total).Error; err != nil {
		logger.Error("Failed to count assignment rules", err, nil)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get assignment rules", err.Error()))
		return
	}

	// Get rules
	var rules []models.AssignmentRule
	if err := query.Order("priority DESC, created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&rules).Error; err != nil {
		logger.Error("Failed to get assignment rules", err, nil)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get assignment rules", err.Error()))
		return
	}

	response := NewPaginatedResponseWithItems(rules, int(total), page, limit)
	c.JSON(http.StatusOK, NewSuccessResponse("Assignment rules retrieved successfully", response))
}

// GetAssignmentRule godoc
// @Summary Get assignment rule
// @Description Get a specific assignment rule by ID
// @Tags assignment-rules
// @Accept json
// @Produce json
// @Param id path int true "Rule ID"
// @Success 200 {object} APIResponse{data=models.AssignmentRule}
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /assignment-rules/{id} [get]
func (h *AssignmentRuleHandler) GetAssignmentRule(c *gin.Context) {
	ruleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid rule ID", err.Error()))
		return
	}

	var rule models.AssignmentRule
	if err := h.db.Where("deleted_at IS NULL").First(&rule, ruleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, NewErrorResponse("Assignment rule not found", ""))
			return
		}
		logger.Error("Failed to get assignment rule", err, map[string]interface{}{
			"rule_id": ruleID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get assignment rule", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Assignment rule retrieved successfully", rule))
}

// UpdateAssignmentRule godoc
// @Summary Update assignment rule
// @Description Update an existing assignment rule
// @Tags assignment-rules
// @Accept json
// @Produce json
// @Param id path int true "Rule ID"
// @Param rule body models.AssignmentRuleRequest true "Updated assignment rule data"
// @Success 200 {object} APIResponse{data=models.AssignmentRule}
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /assignment-rules/{id} [put]
func (h *AssignmentRuleHandler) UpdateAssignmentRule(c *gin.Context) {
	ruleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid rule ID", err.Error()))
		return
	}

	var req models.AssignmentRuleRequest
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
	var rule models.AssignmentRule
	if err := h.db.Where("deleted_at IS NULL").First(&rule, ruleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, NewErrorResponse("Assignment rule not found", ""))
			return
		}
		logger.Error("Failed to get assignment rule for update", err, map[string]interface{}{
			"rule_id": ruleID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get assignment rule", err.Error()))
		return
	}

	// Update fields
	updates := map[string]interface{}{
		"name":                     req.Name,
		"description":              req.Description,
		"type":                     req.Type,
		"conditions":               req.Conditions,
		"settings":                 req.Settings,
		"assignee_ids":             req.AssigneeIDs,
		"fallback_user_id":         req.FallbackUserID,
		"business_hours_start":     req.BusinessHoursStart,
		"business_hours_end":       req.BusinessHoursEnd,
		"working_days":             req.WorkingDays,
		"max_assignments_per_hour": req.MaxAssignmentsPerHour,
		"max_assignments_per_day":  req.MaxAssignmentsPerDay,
		"updated_by":               *userID,
	}

	// Set optional fields
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.BusinessHoursEnabled != nil {
		updates["business_hours_enabled"] = *req.BusinessHoursEnabled
	}
	if req.Timezone != nil {
		updates["timezone"] = *req.Timezone
	}

	if err := h.db.Model(&rule).Updates(updates).Error; err != nil {
		logger.Error("Failed to update assignment rule", err, map[string]interface{}{
			"rule_id": ruleID,
			"user_id": *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to update assignment rule", err.Error()))
		return
	}

	// Reload updated rule
	h.db.First(&rule, ruleID)

	logger.Info("Assignment rule updated", map[string]interface{}{
		"rule_id":   rule.ID,
		"rule_name": rule.Name,
		"user_id":   *userID,
	})

	c.JSON(http.StatusOK, NewSuccessResponse("Assignment rule updated successfully", rule))
}

// DeleteAssignmentRule godoc
// @Summary Delete assignment rule
// @Description Soft delete an assignment rule
// @Tags assignment-rules
// @Accept json
// @Produce json
// @Param id path int true "Rule ID"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /assignment-rules/{id} [delete]
func (h *AssignmentRuleHandler) DeleteAssignmentRule(c *gin.Context) {
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
	var rule models.AssignmentRule
	if err := h.db.Where("deleted_at IS NULL").First(&rule, ruleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, NewErrorResponse("Assignment rule not found", ""))
			return
		}
		logger.Error("Failed to get assignment rule for deletion", err, map[string]interface{}{
			"rule_id": ruleID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get assignment rule", err.Error()))
		return
	}

	// Soft delete the rule
	if err := h.db.Model(&rule).Updates(map[string]interface{}{
		"updated_by": *userID,
	}).Delete(&rule).Error; err != nil {
		logger.Error("Failed to delete assignment rule", err, map[string]interface{}{
			"rule_id": ruleID,
			"user_id": *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to delete assignment rule", err.Error()))
		return
	}

	logger.Info("Assignment rule deleted", map[string]interface{}{
		"rule_id":   rule.ID,
		"rule_name": rule.Name,
		"user_id":   *userID,
	})

	c.JSON(http.StatusOK, NewSuccessResponse("Assignment rule deleted successfully", nil))
}

// ToggleAssignmentRule godoc
// @Summary Toggle assignment rule status
// @Description Toggle an assignment rule between active and inactive
// @Tags assignment-rules
// @Accept json
// @Produce json
// @Param id path int true "Rule ID"
// @Success 200 {object} APIResponse{data=models.AssignmentRule}
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /assignment-rules/{id}/toggle [post]
func (h *AssignmentRuleHandler) ToggleAssignmentRule(c *gin.Context) {
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

	// Get existing rule
	var rule models.AssignmentRule
	if err := h.db.Where("deleted_at IS NULL").First(&rule, ruleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, NewErrorResponse("Assignment rule not found", ""))
			return
		}
		logger.Error("Failed to get assignment rule for toggle", err, map[string]interface{}{
			"rule_id": ruleID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get assignment rule", err.Error()))
		return
	}

	// Toggle status
	newStatus := models.AssignmentRuleInactive
	if rule.Status == models.AssignmentRuleInactive {
		newStatus = models.AssignmentRuleActive
	}

	if err := h.db.Model(&rule).Updates(map[string]interface{}{
		"status":     newStatus,
		"updated_by": *userID,
	}).Error; err != nil {
		logger.Error("Failed to toggle assignment rule", err, map[string]interface{}{
			"rule_id": ruleID,
			"user_id": *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to toggle assignment rule", err.Error()))
		return
	}

	// Reload updated rule
	h.db.First(&rule, ruleID)

	logger.Info("Assignment rule toggled", map[string]interface{}{
		"rule_id":    rule.ID,
		"rule_name":  rule.Name,
		"new_status": newStatus,
		"user_id":    *userID,
	})

	c.JSON(http.StatusOK, NewSuccessResponse("Assignment rule status updated successfully", rule))
}

// TestAssignmentRule godoc
// @Summary Test assignment rule
// @Description Test an assignment rule against a contact to see if it matches
// @Tags assignment-rules
// @Accept json
// @Produce json
// @Param id path int true "Rule ID"
// @Param test body map[string]interface{} true "Test data (contact_id or contact data)"
// @Success 200 {object} APIResponse{data=map[string]interface{}}
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /assignment-rules/{id}/test [post]
func (h *AssignmentRuleHandler) TestAssignmentRule(c *gin.Context) {
	ruleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid rule ID", err.Error()))
		return
	}

	var testData map[string]interface{}
	if err := c.ShouldBindJSON(&testData); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid test data", err.Error()))
		return
	}

	// Get assignment rule
	var rule models.AssignmentRule
	if err := h.db.Where("deleted_at IS NULL").First(&rule, ruleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, NewErrorResponse("Assignment rule not found", ""))
			return
		}
		logger.Error("Failed to get assignment rule for testing", err, map[string]interface{}{
			"rule_id": ruleID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get assignment rule", err.Error()))
		return
	}

	// Get contact for testing
	var contact models.Contact
	if contactID, exists := testData["contact_id"]; exists {
		if id, ok := contactID.(float64); ok {
			if err := h.db.Preload("ContactType").Preload("ContactSource").
				First(&contact, uint(id)).Error; err != nil {
				c.JSON(http.StatusBadRequest, NewErrorResponse("Contact not found", ""))
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid contact_id", ""))
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, NewErrorResponse("contact_id is required for testing", ""))
		return
	}

	// Test the rule using assignment service
	_ = services.NewAssignmentService(h.db) // For future use
	
	// Create a simplified evaluation method for testing
	matches := h.evaluateRuleForTesting(&rule, &contact, testData)
	
	var selectedAssignee uint
	var assigneeError string
	if matches {
		if assignee, err := h.selectAssigneeForTesting(&rule, &contact); err != nil {
			assigneeError = err.Error()
		} else {
			selectedAssignee = assignee
		}
	}

	result := map[string]interface{}{
		"rule_matches":      matches,
		"selected_assignee": selectedAssignee,
		"assignee_error":    assigneeError,
		"rule_name":         rule.Name,
		"rule_type":         rule.Type,
		"contact_id":        contact.ID,
		"contact_name":      contact.GetFullName(),
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Rule test completed", result))
}

// Helper methods for testing

func (h *AssignmentRuleHandler) evaluateRuleForTesting(rule *models.AssignmentRule, contact *models.Contact, contextData map[string]interface{}) bool {
	// Simplified rule evaluation for testing
	for _, condition := range rule.Conditions {
		if !h.evaluateConditionForTesting(condition, contact, contextData) {
			return false
		}
	}
	return true
}

func (h *AssignmentRuleHandler) evaluateConditionForTesting(condition models.AssignmentCondition, contact *models.Contact, contextData map[string]interface{}) bool {
	// This is a simplified version of condition evaluation
	fieldValue := h.getFieldValueForTesting(condition.Field, contact, contextData)
	
	switch condition.Operator {
	case "equals":
		return fieldValue == condition.Value
	case "contains":
		if str1, ok1 := fieldValue.(string); ok1 {
			if str2, ok2 := condition.Value.(string); ok2 {
				return strings.Contains(strings.ToLower(str1), strings.ToLower(str2))
			}
		}
	case "greater_than":
		if num1, ok1 := fieldValue.(float64); ok1 {
			if num2, ok2 := condition.Value.(float64); ok2 {
				return num1 > num2
			}
		}
	}
	return false
}

func (h *AssignmentRuleHandler) getFieldValueForTesting(fieldName string, contact *models.Contact, contextData map[string]interface{}) interface{} {
	// Check context data first
	if val, exists := contextData[fieldName]; exists {
		return val
	}

	// Map contact fields
	switch fieldName {
	case "contact_type_id":
		return float64(contact.ContactTypeID)
	case "contact_source_id":
		return float64(contact.ContactSourceID)
	case "priority":
		return string(contact.Priority)
	case "lead_score":
		return float64(contact.LeadScore)
	case "estimated_value":
		return contact.EstimatedValue
	case "country":
		return contact.Country
	}
	return nil
}

func (h *AssignmentRuleHandler) selectAssigneeForTesting(rule *models.AssignmentRule, contact *models.Contact) (uint, error) {
	// Extract user IDs from JSONArray
	var userIDs []uint
	for _, val := range rule.AssigneeIDs {
		if id, ok := val.(float64); ok {
			userIDs = append(userIDs, uint(id))
		}
	}

	if len(userIDs) == 0 {
		if rule.FallbackUserID != nil {
			return *rule.FallbackUserID, nil
		}
		return 0, errors.New("no assignees available")
	}

	// For testing, just return the first available assignee
	return userIDs[0], nil
}