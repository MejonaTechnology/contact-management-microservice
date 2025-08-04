package handlers

import (
	"contact-service/internal/models"
	"contact-service/pkg/database"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DashboardContactHandler handles contact operations in dashboard-compatible format
type DashboardContactHandler struct {
	db *gorm.DB
}

// NewDashboardContactHandler creates a new dashboard-compatible contact handler
func NewDashboardContactHandler() *DashboardContactHandler {
	return &DashboardContactHandler{
		db: database.GetDB(),
	}
}

// GetContactSubmissions retrieves all contact submissions with pagination (dashboard compatible)
func (h *DashboardContactHandler) GetContactSubmissions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	query := h.db.Model(&models.Contact{}).Where("deleted_at IS NULL")

	// Apply filters
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if search != "" {
		query = query.Where("first_name LIKE ? OR last_name LIKE ? OR email LIKE ? OR subject LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get paginated results
	var contacts []models.Contact
	if err := query.Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&contacts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch contacts"})
		return
	}

	pagination := models.CreatePagination(page, limit, total)

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Success: true,
		Message: "Contacts retrieved successfully",
		Data:    contacts,
		Meta:    pagination,
	})
}

// GetContactSubmission retrieves a single contact submission (dashboard compatible)
func (h *DashboardContactHandler) GetContactSubmission(c *gin.Context) {
	id := c.Param("id")
	
	var contact models.Contact
	if err := h.db.Where("deleted_at IS NULL").First(&contact, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch contact"})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Contact retrieved successfully",
		Data:    contact,
	})
}

// CreateContactSubmission creates a new contact submission (dashboard compatible)
func (h *DashboardContactHandler) CreateContactSubmission(c *gin.Context) {
	var req models.ContactSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Honeypot check for spam detection
	if req.Website != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Spam detected"})
		return
	}

	// Convert simple form request to full Contact model for contacts table
	contact := &models.Contact{
		FirstName:             req.Name,
		Email:                 req.Email,
		Phone:                 req.Phone,
		Subject:               req.Subject,
		Message:               &req.Message,
		ContactTypeID:         1, // General Inquiry
		ContactSourceID:       1, // Website
		Status:                models.StatusNew,
		Priority:              models.PriorityMedium,
		Country:               "India",
		PreferredContactMethod: models.ContactMethodEmail,
		DataSource:            "form",
		DataProcessingConsent: true,
	}
	
	// Handle name parsing for first_name/last_name
	if len(req.Name) > 0 {
		parts := strings.Fields(req.Name)
		if len(parts) > 1 {
			contact.FirstName = parts[0]
			lastNameStr := strings.Join(parts[1:], " ")
			contact.LastName = &lastNameStr
		}
	}
	
	// Set source if provided
	if req.Source != nil {
		contact.UTMSource = req.Source
	}
	
	// Set additional fields from request context
	if clientIP := c.ClientIP(); clientIP != "" {
		contact.IPAddress = &clientIP
	}
	
	if userAgent := c.GetHeader("User-Agent"); userAgent != "" {
		contact.UserAgent = &userAgent
	}
	
	if referer := c.GetHeader("Referer"); referer != "" {
		contact.ReferrerURL = &referer
	}

	if err := h.db.Create(&contact).Error; err != nil {
		// Check for duplicate email
		if strings.Contains(err.Error(), "Duplicate entry") {
			c.JSON(http.StatusConflict, gin.H{"error": "Contact with this email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create contact"})
		return
	}

	c.JSON(http.StatusCreated, models.Response{
		Success: true,
		Message: "Contact created successfully",
		Data:    contact,
	})
}

// UpdateContactSubmissionStatus updates the status of a contact submission (dashboard compatible)
func (h *DashboardContactHandler) UpdateContactSubmissionStatus(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")
	
	var req struct {
		Status     string  `json:"status" binding:"required,oneof=new in_progress resolved spam"`
		Response   *string `json:"response"`
		AssignedTo *uint   `json:"assigned_to"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var contact models.ContactSubmission
	if err := h.db.First(&contact, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch contact"})
		return
	}

	// Update contact
	contact.Status = req.Status
	if req.Response != nil {
		contact.ResponseSent = true
	}
	if req.AssignedTo != nil {
		assignedToInt := int(*req.AssignedTo)
		contact.AssignedTo = &assignedToInt
	} else if contact.AssignedTo == nil && userID != nil {
		// Auto-assign to current user if not assigned
		uid := int(userID.(uint))
		contact.AssignedTo = &uid
	}

	if err := h.db.Save(&contact).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update contact"})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Contact updated successfully",
		Data:    contact,
	})
}

// GetContactSubmissionStats returns contact statistics (dashboard compatible)
func (h *DashboardContactHandler) GetContactSubmissionStats(c *gin.Context) {
	var stats struct {
		Total       int64 `json:"total"`
		New         int64 `json:"new"`
		InProgress  int64 `json:"in_progress"`
		Resolved    int64 `json:"resolved"`
		Spam        int64 `json:"spam"`
	}

	// Get stats with error handling
	h.db.Model(&models.ContactSubmission{}).Count(&stats.Total)
	h.db.Model(&models.ContactSubmission{}).Where("status = ?", "new").Count(&stats.New)
	h.db.Model(&models.ContactSubmission{}).Where("status = ?", "in_progress").Count(&stats.InProgress)
	h.db.Model(&models.ContactSubmission{}).Where("status = ?", "resolved").Count(&stats.Resolved)
	h.db.Model(&models.ContactSubmission{}).Where("status = ?", "spam").Count(&stats.Spam)

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Contact statistics retrieved successfully",
		Data:    stats,
	})
}

// ExportContactSubmissions exports contacts to CSV format (dashboard compatible)
func (h *DashboardContactHandler) ExportContactSubmissions(c *gin.Context) {
	status := c.Query("status")
	format := c.DefaultQuery("format", "csv")

	if format != "csv" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only CSV format is supported"})
		return
	}

	query := h.db.Model(&models.ContactSubmission{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var contacts []models.ContactSubmission
	if err := query.Order("created_at DESC").Find(&contacts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch contacts for export"})
		return
	}

	// Generate CSV content
	csvContent := "ID,Name,Email,Phone,Subject,Message,Status,Source,Created At\n"
	for _, contact := range contacts {
		phone := ""
		if contact.Phone != nil {
			phone = *contact.Phone
		}
		subject := ""
		if contact.Subject != nil {
			subject = *contact.Subject
		}
		source := ""
		if contact.Source != nil {
			source = *contact.Source
		}
		
		csvContent += fmt.Sprintf("%d,\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"\n",
			contact.ID, contact.Name, contact.Email, phone, subject, 
			contact.Message, contact.Status, source, contact.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=contacts.csv")
	c.String(http.StatusOK, csvContent)
}

// BulkUpdateContactSubmissions updates multiple contacts (dashboard compatible)
func (h *DashboardContactHandler) BulkUpdateContactSubmissions(c *gin.Context) {
	var req struct {
		IDs        []int  `json:"ids" binding:"required,min=1"`
		Status     string `json:"status"`
		AssignedTo *int   `json:"assigned_to"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.AssignedTo != nil {
		updates["assigned_to"] = *req.AssignedTo
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No updates specified"})
		return
	}

	result := h.db.Model(&models.ContactSubmission{}).
		Where("id IN ?", req.IDs).
		Updates(updates)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update contacts"})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: fmt.Sprintf("Successfully updated %d contacts", result.RowsAffected),
		Data:    gin.H{"updated_count": result.RowsAffected},
	})
}
