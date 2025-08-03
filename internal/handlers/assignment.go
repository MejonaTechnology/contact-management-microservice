package handlers

import (
	"contact-service/internal/models"
	"contact-service/internal/services"
	"contact-service/pkg/database"
	"contact-service/pkg/logger"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AssignmentHandler handles contact assignment requests
type AssignmentHandler struct {
	assignmentService *services.AssignmentService
}

// NewAssignmentHandler creates a new assignment handler
func NewAssignmentHandler() *AssignmentHandler {
	return &AssignmentHandler{
		assignmentService: services.NewAssignmentService(database.DB),
	}
}

// AssignContactAutomatically godoc
// @Summary Automatically assign a contact
// @Description Automatically assign a contact based on assignment rules
// @Tags assignments
// @Accept json
// @Produce json
// @Param id path int true "Contact ID"
// @Param context body map[string]interface{} false "Context data for assignment"
// @Success 200 {object} APIResponse{data=models.ContactAssignment}
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /assignments/auto/{id} [post]
func (h *AssignmentHandler) AssignContactAutomatically(c *gin.Context) {
	contactID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid contact ID", err.Error()))
		return
	}

	// Parse context data (optional)
	var contextData map[string]interface{}
	if err := c.ShouldBindJSON(&contextData); err != nil {
		contextData = make(map[string]interface{})
	}

	assignment, err := h.assignmentService.AssignContactAutomatically(uint(contactID), contextData)
	if err != nil {
		logger.Error("Failed to assign contact automatically", err, map[string]interface{}{
			"contact_id": contactID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to assign contact", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Contact assigned automatically", assignment))
}

// AssignContactManually godoc
// @Summary Manually assign a contact
// @Description Manually assign a contact to a specific user
// @Tags assignments
// @Accept json
// @Produce json
// @Param assignment body models.ContactAssignmentRequest true "Assignment details"
// @Success 200 {object} APIResponse{data=models.ContactAssignment}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /assignments/manual [post]
func (h *AssignmentHandler) AssignContactManually(c *gin.Context) {
	var req models.ContactAssignmentRequest
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

	assignment, err := h.assignmentService.AssignContactManually(&req, *userID)
	if err != nil {
		logger.Error("Failed to assign contact manually", err, map[string]interface{}{
			"contact_id":     req.ContactID,
			"assigned_to_id": req.AssignedToID,
			"assigned_by_id": *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to assign contact", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Contact assigned successfully", assignment))
}

// BulkAssignContacts godoc
// @Summary Bulk assign contacts
// @Description Assign multiple contacts to a user in bulk
// @Tags assignments
// @Accept json
// @Produce json
// @Param assignment body models.BulkAssignmentRequest true "Bulk assignment details"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /assignments/bulk [post]
func (h *AssignmentHandler) BulkAssignContacts(c *gin.Context) {
	var req models.BulkAssignmentRequest
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

	err := h.assignmentService.BulkAssignContacts(&req, *userID)
	if err != nil {
		logger.Error("Failed to bulk assign contacts", err, map[string]interface{}{
			"contact_count":  len(req.ContactIDs),
			"assigned_to_id": req.AssignedToID,
			"assigned_by_id": *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to assign contacts", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Contacts assigned successfully", nil))
}

// UnassignContact godoc
// @Summary Unassign a contact
// @Description Remove assignment from a contact
// @Tags assignments
// @Accept json
// @Produce json
// @Param id path int true "Contact ID"  
// @Param reason body map[string]string false "Unassignment reason"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /assignments/unassign/{id} [post]
func (h *AssignmentHandler) UnassignContact(c *gin.Context) {
	contactID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid contact ID", err.Error()))
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	// Parse reason (optional)
	var reasonData map[string]string
	if err := c.ShouldBindJSON(&reasonData); err != nil {
		reasonData = make(map[string]string)
	}
	
	reason := reasonData["reason"]
	if reason == "" {
		reason = "Manual unassignment"
	}

	err = h.assignmentService.UnassignContact(uint(contactID), *userID, reason)
	if err != nil {
		logger.Error("Failed to unassign contact", err, map[string]interface{}{
			"contact_id":       contactID,
			"unassigned_by_id": *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to unassign contact", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Contact unassigned successfully", nil))
}

// ReassignContact godoc
// @Summary Reassign a contact
// @Description Reassign a contact from one user to another
// @Tags assignments
// @Accept json
// @Produce json
// @Param id path int true "Contact ID"
// @Param assignment body models.ContactAssignmentRequest true "New assignment details"
// @Success 200 {object} APIResponse{data=models.ContactAssignment}
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /assignments/reassign/{id} [post]
func (h *AssignmentHandler) ReassignContact(c *gin.Context) {
	contactID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid contact ID", err.Error()))
		return
	}

	var req models.ContactAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request data", err.Error()))
		return
	}

	// Override contact ID from URL
	req.ContactID = uint(contactID)

	// Get user ID from context (set by auth middleware)
	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	assignment, err := h.assignmentService.AssignContactManually(&req, *userID)
	if err != nil {
		logger.Error("Failed to reassign contact", err, map[string]interface{}{
			"contact_id":     contactID,
			"assigned_to_id": req.AssignedToID,
			"assigned_by_id": *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to reassign contact", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Contact reassigned successfully", assignment))
}

// GetUserWorkload godoc
// @Summary Get user workload
// @Description Get current workload information for a user
// @Tags assignments
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} APIResponse{data=models.UserWorkloadResponse}
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /assignments/workload/{id} [get]
func (h *AssignmentHandler) GetUserWorkload(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid user ID", err.Error()))
		return
	}

	workload, err := h.assignmentService.GetUserWorkload(uint(userID))
	if err != nil {
		logger.Error("Failed to get user workload", err, map[string]interface{}{
			"user_id": userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get workload", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("User workload retrieved successfully", workload))
}

// GetMyWorkload godoc
// @Summary Get current user's workload
// @Description Get current workload information for the authenticated user
// @Tags assignments
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=models.UserWorkloadResponse}
// @Failure 401 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /assignments/my-workload [get]
func (h *AssignmentHandler) GetMyWorkload(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	workload, err := h.assignmentService.GetUserWorkload(*userID)
	if err != nil {
		logger.Error("Failed to get current user workload", err, map[string]interface{}{
			"user_id": *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get workload", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Your workload retrieved successfully", workload))
}

// GetAllWorkloads godoc
// @Summary Get all user workloads
// @Description Get workload information for all users
// @Tags assignments
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=[]models.UserWorkloadResponse}
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /assignments/workloads [get]
func (h *AssignmentHandler) GetAllWorkloads(c *gin.Context) {
	// Get all active users
	var users []models.AdminUser
	if err := database.DB.Where("is_active = ?", true).Find(&users).Error; err != nil {
		logger.Error("Failed to get active users", err, nil)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get users", err.Error()))
		return
	}

	var workloads []models.UserWorkloadResponse
	for _, user := range users {
		workload, err := h.assignmentService.GetUserWorkload(user.ID)
		if err != nil {
			logger.Error("Failed to get workload for user", err, map[string]interface{}{
				"user_id": user.ID,
			})
			continue
		}
		workloads = append(workloads, *workload)
	}

	c.JSON(http.StatusOK, NewSuccessResponse("User workloads retrieved successfully", workloads))
}

// GetContactAssignmentHistory godoc
// @Summary Get contact assignment history
// @Description Get the assignment history for a specific contact
// @Tags assignments
// @Accept json
// @Produce json
// @Param id path int true "Contact ID"
// @Success 200 {object} APIResponse{data=[]models.AssignmentHistory}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /assignments/history/{id} [get]
func (h *AssignmentHandler) GetContactAssignmentHistory(c *gin.Context) {
	contactID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid contact ID", err.Error()))
		return
	}

	var history []models.AssignmentHistory
	if err := database.DB.Where("contact_id = ?", contactID).
		Preload("Contact").
		Preload("FromUser").
		Preload("ToUser").
		Preload("ChangedBy").
		Preload("Rule").
		Order("created_at DESC").
		Find(&history).Error; err != nil {
		logger.Error("Failed to get assignment history", err, map[string]interface{}{
			"contact_id": contactID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get assignment history", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Assignment history retrieved successfully", history))
}

// GetMyAssignments godoc
// @Summary Get current user's assignments
// @Description Get all active contact assignments for the authenticated user
// @Tags assignments
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param status query string false "Assignment status filter"
// @Success 200 {object} APIResponse{data=PaginatedResponse{items=[]models.ContactAssignment}}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /assignments/my-assignments [get]
func (h *AssignmentHandler) GetMyAssignments(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	// Parse pagination parameters
	page, limit := parsePaginationParams(c)
	offset := (page - 1) * limit

	// Parse filters
	status := c.Query("status")

	// Build query
	query := database.DB.Where("assigned_to_id = ?", *userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Get total count
	var total int64
	if err := query.Model(&models.ContactAssignment{}).Count(&total).Error; err != nil {
		logger.Error("Failed to count assignments", err, map[string]interface{}{
			"user_id": *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get assignments", err.Error()))
		return
	}

	// Get assignments with preloads
	var assignments []models.ContactAssignment
	if err := query.Preload("Contact").
		Preload("Contact.ContactType").
		Preload("Contact.ContactSource").
		Preload("AssignedBy").
		Preload("Rule").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&assignments).Error; err != nil {
		logger.Error("Failed to get assignments", err, map[string]interface{}{
			"user_id": *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get assignments", err.Error()))
		return
	}

	response := NewPaginatedResponseWithItems(assignments, int(total), page, limit)
	c.JSON(http.StatusOK, NewSuccessResponse("Assignments retrieved successfully", response))
}

// AcceptAssignment godoc
// @Summary Accept an assignment
// @Description Mark an assignment as accepted by the assigned user
// @Tags assignments
// @Accept json
// @Produce json
// @Param id path int true "Assignment ID"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 403 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /assignments/{id}/accept [post]
func (h *AssignmentHandler) AcceptAssignment(c *gin.Context) {
	assignmentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid assignment ID", err.Error()))
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	// Get assignment and verify ownership
	var assignment models.ContactAssignment
	if err := database.DB.First(&assignment, assignmentID).Error; err != nil {
		c.JSON(http.StatusNotFound, NewErrorResponse("Assignment not found", ""))
		return
	}

	if assignment.AssignedToID != *userID {
		c.JSON(http.StatusForbidden, NewErrorResponse("Not authorized to accept this assignment", ""))
		return
	}

	// Update assignment as accepted
	now := time.Now()
	if err := database.DB.Model(&assignment).Updates(map[string]interface{}{
		"accepted_at": now,
	}).Error; err != nil {
		logger.Error("Failed to accept assignment", err, map[string]interface{}{
			"assignment_id": assignmentID,
			"user_id":       *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to accept assignment", err.Error()))
		return
	}

	logger.Info("Assignment accepted", map[string]interface{}{
		"assignment_id": assignmentID,
		"user_id":       *userID,
		"contact_id":    assignment.ContactID,
	})

	c.JSON(http.StatusOK, NewSuccessResponse("Assignment accepted successfully", nil))
}

