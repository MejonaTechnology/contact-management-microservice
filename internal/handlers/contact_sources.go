package handlers

import (
	"net/http"
	"strconv"

	"contact-service/internal/models"
	"contact-service/internal/services"
	"contact-service/pkg/logger"

	"github.com/gin-gonic/gin"
)

// ContactSourceHandler handles contact source management
type ContactSourceHandler struct {
	contactSourceService *services.ContactSourceService
}

// NewContactSourceHandler creates a new contact source handler
func NewContactSourceHandler() *ContactSourceHandler {
	return &ContactSourceHandler{
		contactSourceService: services.NewContactSourceService(),
	}
}

// GetContactSources retrieves all contact sources
func GetContactSources(c *gin.Context) {
	handler := NewContactSourceHandler()
	
	activeOnly := c.Query("active_only") == "true"
	
	contactSources, err := handler.contactSourceService.ListContactSources(activeOnly)
	if err != nil {
		logger.Error("Failed to retrieve contact sources", err, map[string]interface{}{
			"active_only": activeOnly,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to retrieve contact sources", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Contact sources retrieved successfully", contactSources))
}

// CreateContactSource creates a new contact source
func CreateContactSource(c *gin.Context) {
	handler := NewContactSourceHandler()
	
	var req models.ContactSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid input", err.Error()))
		return
	}

	userID := getUserIDFromContext(c)
	
	contactSource, err := handler.contactSourceService.CreateContactSource(&req, userID)
	if err != nil {
		logger.Error("Failed to create contact source", err, map[string]interface{}{
			"name":       req.Name,
			"created_by": userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to create contact source", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, NewSuccessResponse("Contact source created successfully", contactSource))
}

// UpdateContactSource updates an existing contact source
func UpdateContactSource(c *gin.Context) {
	handler := NewContactSourceHandler()
	
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid contact source ID", "ID must be a valid number"))
		return
	}

	var req models.ContactSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid input", err.Error()))
		return
	}

	userID := getUserIDFromContext(c)
	
	contactSource, err := handler.contactSourceService.UpdateContactSource(uint(id), &req, userID)
	if err != nil {
		logger.Error("Failed to update contact source", err, map[string]interface{}{
			"id":         id,
			"name":       req.Name,
			"updated_by": userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to update contact source", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Contact source updated successfully", contactSource))
}

// DeleteContactSource deletes a contact source
func DeleteContactSource(c *gin.Context) {
	handler := NewContactSourceHandler()
	
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid contact source ID", "ID must be a valid number"))
		return
	}

	userID := getUserIDFromContext(c)
	
	err = handler.contactSourceService.DeleteContactSource(uint(id), userID)
	if err != nil {
		logger.Error("Failed to delete contact source", err, map[string]interface{}{
			"id":         id,
			"deleted_by": userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to delete contact source", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Contact source deleted successfully", nil))
}