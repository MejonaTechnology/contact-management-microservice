package handlers

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"contact-service/internal/services"
)

// BulkHandler handles bulk operations for contacts
type BulkHandler struct {
	bulkService *services.BulkService
}

// NewBulkHandler creates a new bulk handler
func NewBulkHandler(bulkService *services.BulkService) *BulkHandler {
	return &BulkHandler{
		bulkService: bulkService,
	}
}

// ImportContactsRequest represents the import request structure
type ImportContactsRequest struct {
	SkipHeader bool `form:"skip_header" json:"skip_header"`
}

// @Summary Import contacts from CSV file
// @Description Import multiple contacts from uploaded CSV file with validation and duplicate detection
// @Tags Bulk Operations
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "CSV file containing contact data"
// @Param skip_header formData boolean false "Skip first row as header"
// @Success 200 {object} APIResponse{data=services.BulkImportResult} "Import completed successfully"
// @Failure 400 {object} ErrorResponse "Invalid request or file format"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 413 {object} ErrorResponse "File too large"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/bulk/contacts/import [post]
func (h *BulkHandler) ImportContacts(c *gin.Context) {
	var request ImportContactsRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, NewValidationErrorResponse(err.Error()))
		return
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("File upload failed", err.Error()))
		return
	}
	defer file.Close()

	// Validate file type
	if !isCSVFile(header) {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid file type", "Only CSV files are supported"))
		return
	}

	// Check file size (max 10MB)
	if header.Size > 10*1024*1024 {
		c.JSON(http.StatusRequestEntityTooLarge, NewErrorResponse("File too large", "Maximum file size is 10MB"))
		return
	}

	// Import contacts
	result, err := h.bulkService.ImportContactsFromCSV(file, request.SkipHeader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Import failed", err.Error()))
		return
	}

	// Determine response status based on results
	statusCode := http.StatusOK
	message := "Import completed successfully"
	
	if result.ErrorCount > 0 {
		if result.SuccessCount == 0 {
			statusCode = http.StatusBadRequest
			message = "Import failed - all records had errors"
		} else {
			message = fmt.Sprintf("Import completed with %d errors out of %d records", result.ErrorCount, result.TotalRecords)
		}
	}

	c.JSON(statusCode, NewSuccessResponse(message, result))
}

// @Summary Export contacts to CSV
// @Description Export contacts to CSV format with filtering and field selection options
// @Tags Bulk Operations
// @Accept json
// @Produce text/csv
// @Security BearerAuth
// @Param format query string false "Export format (csv, json)" default(csv)
// @Param status query string false "Filter by contact status"
// @Param type_id query integer false "Filter by contact type ID"
// @Param source_id query integer false "Filter by contact source ID"
// @Param fields query string false "Comma-separated list of fields to export"
// @Param limit query integer false "Maximum number of records to export" default(10000)
// @Param sort_by query string false "Sort field" default(created_at)
// @Param sort_order query string false "Sort order (asc, desc)" default(desc)
// @Success 200 {file} file "CSV file containing contact data"
// @Failure 400 {object} ErrorResponse "Invalid query parameters"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Export failed"
// @Router /api/v1/bulk/contacts/export [get]
func (h *BulkHandler) ExportContacts(c *gin.Context) {
	// Parse query parameters
	format := c.DefaultQuery("format", "csv")
	status := c.Query("status")
	typeIDStr := c.Query("type_id")
	sourceIDStr := c.Query("source_id")
	fieldsStr := c.Query("fields")
	limitStr := c.DefaultQuery("limit", "10000")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	// Validate format
	exportFormat := services.ExportFormat(format)
	if exportFormat != services.ExportFormatCSV && exportFormat != services.ExportFormatJSON {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid format", "Supported formats: csv, json"))
		return
	}

	// Parse limit
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100000 {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid limit", "Limit must be between 1 and 100000"))
		return
	}

	// Build export request
	request := services.ExportRequest{
		Format:    exportFormat,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Limit:     limit,
		Filters:   make(map[string]interface{}),
	}

	// Add filters
	if status != "" {
		request.Filters["status"] = status
	}
	if typeIDStr != "" {
		if typeID, err := strconv.ParseUint(typeIDStr, 10, 32); err == nil {
			request.Filters["type_id"] = float64(typeID)
		}
	}
	if sourceIDStr != "" {
		if sourceID, err := strconv.ParseUint(sourceIDStr, 10, 32); err == nil {
			request.Filters["source_id"] = float64(sourceID)
		}
	}

	// Parse fields
	if fieldsStr != "" {
		request.Fields = strings.Split(fieldsStr, ",")
		// Trim whitespace from fields
		for i, field := range request.Fields {
			request.Fields[i] = strings.TrimSpace(field)
		}
	}

	// Export data
	var data []byte
	var contentType string
	var filename string

	switch exportFormat {
	case services.ExportFormatCSV:
		data, err = h.bulkService.ExportContactsToCSV(request)
		contentType = "text/csv"
		filename = fmt.Sprintf("contacts_%s.csv", getCurrentTimestamp())
		
	case services.ExportFormatJSON:
		data, err = h.bulkService.ExportContactsToJSON(request)
		contentType = "application/json"
		filename = fmt.Sprintf("contacts_%s.json", getCurrentTimestamp())
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Export failed", err.Error()))
		return
	}

	// Set response headers
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Length", strconv.Itoa(len(data)))

	// Send file data
	c.Data(http.StatusOK, contentType, data)
}

// @Summary Bulk update contacts
// @Description Update multiple contacts with the same values based on conditions
// @Tags Bulk Operations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body services.BulkUpdateRequest true "Bulk update request"
// @Success 200 {object} APIResponse{data=services.BulkUpdateResult} "Bulk update completed"
// @Failure 400 {object} ErrorResponse "Invalid request data"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Bulk update failed"
// @Router /api/v1/bulk/contacts/update [post]
func (h *BulkHandler) BulkUpdateContacts(c *gin.Context) {
	var request services.BulkUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, NewValidationErrorResponse(err.Error()))
		return
	}

	// Validate request
	if len(request.ContactIDs) == 0 {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request", "No contact IDs provided"))
		return
	}

	if len(request.Updates) == 0 {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request", "No updates provided"))
		return
	}

	// Limit number of contacts that can be updated at once
	if len(request.ContactIDs) > 1000 {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Too many contacts", "Maximum 1000 contacts can be updated at once"))
		return
	}

	// Perform bulk update
	result, err := h.bulkService.BulkUpdateContacts(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Bulk update failed", err.Error()))
		return
	}

	message := fmt.Sprintf("Updated %d contacts successfully", result.UpdatedCount)
	if result.ErrorCount > 0 {
		message += fmt.Sprintf(" (%d errors, %d skipped)", result.ErrorCount, result.SkippedCount)
	}

	c.JSON(http.StatusOK, NewSuccessResponse(message, result))
}

// @Summary Bulk delete contacts
// @Description Delete multiple contacts by their IDs
// @Tags Bulk Operations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body BulkDeleteRequest true "Bulk delete request"
// @Success 200 {object} APIResponse{data=services.BulkUpdateResult} "Bulk delete completed"
// @Failure 400 {object} ErrorResponse "Invalid request data"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Bulk delete failed"
// @Router /api/v1/bulk/contacts/delete [post]
func (h *BulkHandler) BulkDeleteContacts(c *gin.Context) {
	var request BulkDeleteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, NewValidationErrorResponse(err.Error()))
		return
	}

	// Validate request
	if len(request.ContactIDs) == 0 {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request", "No contact IDs provided"))
		return
	}

	// Limit number of contacts that can be deleted at once
	if len(request.ContactIDs) > 100 {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Too many contacts", "Maximum 100 contacts can be deleted at once"))
		return
	}

	// Perform bulk delete
	result, err := h.bulkService.BulkDeleteContacts(request.ContactIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Bulk delete failed", err.Error()))
		return
	}

	message := fmt.Sprintf("Deleted %d contacts successfully", result.UpdatedCount)
	if result.ErrorCount > 0 {
		message += fmt.Sprintf(" (%d errors)", result.ErrorCount)
	}

	c.JSON(http.StatusOK, NewSuccessResponse(message, result))
}

// @Summary Get bulk operation template
// @Description Download CSV template for bulk contact import
// @Tags Bulk Operations
// @Produce text/csv
// @Success 200 {file} file "CSV template file"
// @Router /api/v1/bulk/contacts/template [get]
func (h *BulkHandler) GetImportTemplate(c *gin.Context) {
	// CSV template content
	template := `Name,Email,Phone,Company,Position,Status,Notes
John Doe,john.doe@example.com,+1-555-123-4567,Acme Corp,Manager,new,Interested in web development
Jane Smith,jane.smith@techcorp.com,+1-555-987-6543,TechCorp,CTO,contacted,Requires custom software`

	// Set response headers
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=\"contact_import_template.csv\"")
	c.Header("Content-Length", strconv.Itoa(len(template)))

	c.String(http.StatusOK, template)
}

// @Summary Get bulk operation status
// @Description Get status and statistics of recent bulk operations
// @Tags Bulk Operations
// @Produce json
// @Security BearerAuth
// @Success 200 {object} APIResponse{data=BulkOperationStats} "Bulk operation statistics"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /api/v1/bulk/contacts/status [get]
func (h *BulkHandler) GetBulkOperationStatus(c *gin.Context) {
	// This would typically query a job queue or operation log
	// For now, return static information
	stats := BulkOperationStats{
		TotalImports:    0,
		TotalExports:    0,
		TotalUpdates:    0,
		TotalDeletes:    0,
		RecentImports:   make([]BulkOperationInfo, 0),
		RecentExports:   make([]BulkOperationInfo, 0),
		SystemLimits: BulkOperationLimits{
			MaxImportSize:      10 * 1024 * 1024, // 10MB
			MaxExportRecords:   100000,
			MaxUpdateBatchSize: 1000,
			MaxDeleteBatchSize: 100,
		},
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Bulk operation status retrieved", stats))
}

// Request/Response structures
type BulkDeleteRequest struct {
	ContactIDs []uint `json:"contact_ids" binding:"required,min=1"`
}

type BulkOperationStats struct {
	TotalImports  int                   `json:"total_imports"`
	TotalExports  int                   `json:"total_exports"`
	TotalUpdates  int                   `json:"total_updates"`
	TotalDeletes  int                   `json:"total_deletes"`
	RecentImports []BulkOperationInfo   `json:"recent_imports"`
	RecentExports []BulkOperationInfo   `json:"recent_exports"`
	SystemLimits  BulkOperationLimits   `json:"system_limits"`
}

type BulkOperationInfo struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	RecordCount int    `json:"record_count"`
	CreatedAt   string `json:"created_at"`
	CompletedAt string `json:"completed_at,omitempty"`
}

type BulkOperationLimits struct {
	MaxImportSize      int64 `json:"max_import_size"`
	MaxExportRecords   int   `json:"max_export_records"`
	MaxUpdateBatchSize int   `json:"max_update_batch_size"`
	MaxDeleteBatchSize int   `json:"max_delete_batch_size"`
}

// Helper functions
func isCSVFile(header *multipart.FileHeader) bool {
	// Check file extension
	filename := strings.ToLower(header.Filename)
	if !strings.HasSuffix(filename, ".csv") {
		return false
	}

	// Check MIME type
	contentType := header.Header.Get("Content-Type")
	validTypes := []string{
		"text/csv",
		"application/csv",
		"text/plain",
		"application/octet-stream", // Some browsers send this for CSV
	}

	for _, validType := range validTypes {
		if strings.Contains(contentType, validType) {
			return true
		}
	}

	return false
}

func getCurrentTimestamp() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}