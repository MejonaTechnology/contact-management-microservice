package handlers

import (
	"net/http"
	"strconv"

	"contact-service/internal/models"
	"contact-service/internal/services"
	"contact-service/pkg/logger"

	"github.com/gin-gonic/gin"
)

// ContactTypeHandler handles contact type management
type ContactTypeHandler struct {
	contactTypeService *services.ContactTypeService
}

// NewContactTypeHandler creates a new contact type handler
func NewContactTypeHandler() *ContactTypeHandler {
	return &ContactTypeHandler{
		contactTypeService: services.NewContactTypeService(),
	}
}

// GetContactTypes retrieves all contact types
func GetContactTypes(c *gin.Context) {
	handler := NewContactTypeHandler()
	
	activeOnly := c.Query("active_only") == "true"
	
	contactTypes, err := handler.contactTypeService.ListContactTypes(activeOnly)
	if err != nil {
		logger.Error("Failed to retrieve contact types", err, map[string]interface{}{
			"active_only": activeOnly,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to retrieve contact types", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Contact types retrieved successfully", contactTypes))
}

// CreateContactType creates a new contact type
func CreateContactType(c *gin.Context) {
	handler := NewContactTypeHandler()
	
	var req models.ContactTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid input", err.Error()))
		return
	}

	userID := getUserIDFromContext(c)
	
	contactType, err := handler.contactTypeService.CreateContactType(&req, userID)
	if err != nil {
		logger.Error("Failed to create contact type", err, map[string]interface{}{
			"name":       req.Name,
			"created_by": userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to create contact type", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, NewSuccessResponse("Contact type created successfully", contactType))
}

// UpdateContactType updates an existing contact type
func UpdateContactType(c *gin.Context) {
	handler := NewContactTypeHandler()
	
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid contact type ID", "ID must be a valid number"))
		return
	}

	var req models.ContactTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid input", err.Error()))
		return
	}

	userID := getUserIDFromContext(c)
	
	contactType, err := handler.contactTypeService.UpdateContactType(uint(id), &req, userID)
	if err != nil {
		logger.Error("Failed to update contact type", err, map[string]interface{}{
			"id":         id,
			"name":       req.Name,
			"updated_by": userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to update contact type", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Contact type updated successfully", contactType))
}

// DeleteContactType deletes a contact type
func DeleteContactType(c *gin.Context) {
	handler := NewContactTypeHandler()
	
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid contact type ID", "ID must be a valid number"))
		return
	}

	userID := getUserIDFromContext(c)
	
	err = handler.contactTypeService.DeleteContactType(uint(id), userID)
	if err != nil {
		logger.Error("Failed to delete contact type", err, map[string]interface{}{
			"id":         id,
			"deleted_by": userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to delete contact type", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Contact type deleted successfully", nil))
}