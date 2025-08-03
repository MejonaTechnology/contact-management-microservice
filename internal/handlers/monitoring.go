package handlers

import (
	"contact-service/internal/services"
	"contact-service/pkg/database"
	"contact-service/pkg/errors"
	"contact-service/pkg/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// MonitoringHandler handles monitoring and health check requests
type MonitoringHandler struct {
	monitoringService *services.MonitoringService
}

// NewMonitoringHandler creates a new monitoring handler
func NewMonitoringHandler() *MonitoringHandler {
	return &MonitoringHandler{
		monitoringService: services.NewMonitoringService(database.DB),
	}
}

// GetSystemHealth godoc
// @Summary Get system health status
// @Description Get comprehensive system health information including error rates, metrics, and alerts
// @Tags monitoring
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=map[string]interface{}}
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /monitoring/health [get]
func (h *MonitoringHandler) GetSystemHealth(c *gin.Context) {
	healthStatus := h.monitoringService.GetHealthStatus()

	c.JSON(http.StatusOK, NewSuccessResponse("System health retrieved successfully", healthStatus))
}

// GetErrorStats godoc
// @Summary Get error statistics
// @Description Get detailed error statistics including counts by code, severity, and category
// @Tags monitoring
// @Accept json
// @Produce json
// @Param hours query int false "Hours to look back for statistics" default(24)
// @Success 200 {object} APIResponse{data=errors.ErrorStats}
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /monitoring/errors [get]
func (h *MonitoringHandler) GetErrorStats(c *gin.Context) {
	errorStats := h.monitoringService.GetErrorStats()

	c.JSON(http.StatusOK, NewSuccessResponse("Error statistics retrieved successfully", errorStats))
}

// GetSystemMetrics godoc
// @Summary Get system metrics
// @Description Get system performance and business metrics
// @Tags monitoring
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=map[string]interface{}}
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /monitoring/metrics [get]
func (h *MonitoringHandler) GetSystemMetrics(c *gin.Context) {
	metrics := h.monitoringService.GetSystemMetrics()

	c.JSON(http.StatusOK, NewSuccessResponse("System metrics retrieved successfully", metrics))
}

// GetActiveAlerts godoc
// @Summary Get active alerts
// @Description Get all currently active system alerts
// @Tags monitoring
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=[]services.Alert}
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /monitoring/alerts [get]
func (h *MonitoringHandler) GetActiveAlerts(c *gin.Context) {
	alerts := h.monitoringService.GetActiveAlerts()

	c.JSON(http.StatusOK, NewSuccessResponse("Active alerts retrieved successfully", alerts))
}

// AcknowledgeAlert godoc
// @Summary Acknowledge an alert
// @Description Acknowledge a system alert to mark it as handled
// @Tags monitoring
// @Accept json
// @Produce json
// @Param id path string true "Alert ID"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /monitoring/alerts/{id}/acknowledge [post]
func (h *MonitoringHandler) AcknowledgeAlert(c *gin.Context) {
	alertID := c.Param("id")
	if alertID == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Alert ID is required", ""))
		return
	}

	// Get user ID from context
	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	err := h.monitoringService.AcknowledgeAlert(alertID, *userID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.HTTPStatus, NewErrorResponse(appErr.Message, ""))
		} else {
			c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to acknowledge alert", err.Error()))
		}
		return
	}

	logger.Info("Alert acknowledged", map[string]interface{}{
		"alert_id": alertID,
		"user_id":  *userID,
	})

	c.JSON(http.StatusOK, NewSuccessResponse("Alert acknowledged successfully", nil))
}

// CreateAlert godoc
// @Summary Create a system alert
// @Description Create a new system alert for monitoring purposes
// @Tags monitoring
// @Accept json
// @Produce json
// @Param alert body CreateAlertRequest true "Alert data"
// @Success 201 {object} APIResponse{data=map[string]string}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /monitoring/alerts [post]
func (h *MonitoringHandler) CreateAlert(c *gin.Context) {
	var req CreateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request data", err.Error()))
		return
	}

	alertID := h.monitoringService.CreateAlert(
		req.Type,
		req.Severity,
		req.Title,
		req.Message,
		req.Conditions,
	)

	response := map[string]string{
		"alert_id": alertID,
	}

	c.JSON(http.StatusCreated, NewSuccessResponse("Alert created successfully", response))
}

// TrackError godoc
// @Summary Track an error
// @Description Manually track an error for monitoring purposes
// @Tags monitoring
// @Accept json
// @Produce json
// @Param error body TrackErrorRequest true "Error data"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /monitoring/errors/track [post]
func (h *MonitoringHandler) TrackError(c *gin.Context) {
	var req TrackErrorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request data", err.Error()))
		return
	}

	// Create AppError from request
	appErr := errors.NewAppError(
		errors.ErrorCode(req.Code),
		req.Message,
		nil,
	)
	
	if req.Severity != "" {
		appErr.Severity = errors.ErrorSeverity(req.Severity)
	}
	if req.Category != "" {
		appErr.Category = errors.ErrorCategory(req.Category)
	}
	
	appErr.Endpoint = req.Endpoint
	appErr.Context = req.Context

	// Get user ID from context if available
	if userID := getUserIDFromContext(c); userID != nil {
		appErr.UserID = userID
	}

	// Track the error
	h.monitoringService.TrackError(appErr)

	c.JSON(http.StatusOK, NewSuccessResponse("Error tracked successfully", nil))
}

// TrackMetric godoc
// @Summary Track a metric
// @Description Manually track a performance or business metric
// @Tags monitoring
// @Accept json
// @Produce json
// @Param metric body TrackMetricRequest true "Metric data"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /monitoring/metrics/track [post]
func (h *MonitoringHandler) TrackMetric(c *gin.Context) {
	var req TrackMetricRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request data", err.Error()))
		return
	}

	metricType := req.Type
	if metricType == "" {
		metricType = "performance"
	}

	h.monitoringService.TrackMetric(
		req.Name,
		req.Value,
		req.Unit,
		req.Tags,
		metricType,
	)

	c.JSON(http.StatusOK, NewSuccessResponse("Metric tracked successfully", nil))
}

// GetSystemLogs godoc
// @Summary Get system logs
// @Description Get recent system logs with filtering options
// @Tags monitoring
// @Accept json
// @Produce json
// @Param level query string false "Log level filter (debug, info, warn, error)"
// @Param hours query int false "Hours to look back" default(1)
// @Param limit query int false "Maximum number of logs to return" default(100)
// @Success 200 {object} APIResponse{data=[]LogEntry}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /monitoring/logs [get]
func (h *MonitoringHandler) GetSystemLogs(c *gin.Context) {
	// Parse query parameters
	level := c.Query("level")
	hoursStr := c.Query("hours")
	limitStr := c.Query("limit")

	hours := 1
	if hoursStr != "" {
		if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 {
			hours = h
		}
	}

	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	// For now, return a placeholder response
	// In a real implementation, you would query log files or a log aggregation service
	logs := []LogEntry{
		{
			Timestamp: "2025-01-01T15:30:00Z",
			Level:     "info",
			Message:   "System started successfully",
			Service:   "contact-service",
			RequestID: "req_123",
		},
		{
			Timestamp: "2025-01-01T15:30:15Z",
			Level:     "error",
			Message:   "Database connection failed",
			Service:   "contact-service",
			Error:     "connection timeout",
		},
	}

	// Filter by level if specified
	if level != "" {
		filteredLogs := make([]LogEntry, 0)
		for _, log := range logs {
			if log.Level == level {
				filteredLogs = append(filteredLogs, log)
			}
		}
		logs = filteredLogs
	}

	// Apply limit
	if len(logs) > limit {
		logs = logs[:limit]
	}

	response := map[string]interface{}{
		"logs":   logs,
		"total":  len(logs),
		"hours":  hours,
		"level":  level,
		"limit":  limit,
	}

	c.JSON(http.StatusOK, NewSuccessResponse("System logs retrieved successfully", response))
}

// SetLogLevel godoc
// @Summary Set log level
// @Description Dynamically change the system log level
// @Tags monitoring
// @Accept json
// @Produce json
// @Param request body SetLogLevelRequest true "Log level data"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /monitoring/logs/level [post]
func (h *MonitoringHandler) SetLogLevel(c *gin.Context) {
	var req SetLogLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request data", err.Error()))
		return
	}

	// Validate log level
	validLevels := []string{"debug", "info", "warn", "error", "fatal"}
	isValid := false
	for _, level := range validLevels {
		if level == req.Level {
			isValid = true
			break
		}
	}

	if !isValid {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid log level", "Valid levels: debug, info, warn, error, fatal"))
		return
	}

	// Set the log level
	if err := logger.SetLevel(req.Level); err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to set log level", err.Error()))
		return
	}

	logger.Info("Log level changed", map[string]interface{}{
		"new_level": req.Level,
		"user_id":   getUserIDFromContext(c),
	})

	c.JSON(http.StatusOK, NewSuccessResponse("Log level updated successfully", map[string]string{
		"level": req.Level,
	}))
}

// Request/Response types

// CreateAlertRequest represents the request to create an alert
type CreateAlertRequest struct {
	Type       string                 `json:"type" binding:"required"`
	Severity   string                 `json:"severity" binding:"required,oneof=low medium high critical"`
	Title      string                 `json:"title" binding:"required,min=3,max=255"`
	Message    string                 `json:"message" binding:"required,min=3"`
	Conditions map[string]interface{} `json:"conditions"`
}

// TrackErrorRequest represents the request to track an error
type TrackErrorRequest struct {
	Code     string                 `json:"code" binding:"required"`
	Message  string                 `json:"message" binding:"required"`
	Severity string                 `json:"severity,omitempty"`
	Category string                 `json:"category,omitempty"`
	Endpoint string                 `json:"endpoint,omitempty"`
	Context  map[string]interface{} `json:"context,omitempty"`
}

// TrackMetricRequest represents the request to track a metric
type TrackMetricRequest struct {
	Name  string            `json:"name" binding:"required"`
	Value float64           `json:"value" binding:"required"`
	Unit  string            `json:"unit" binding:"required"`
	Type  string            `json:"type,omitempty"` // performance, business
	Tags  map[string]string `json:"tags,omitempty"`
}

// SetLogLevelRequest represents the request to change log level
type SetLogLevelRequest struct {
	Level string `json:"level" binding:"required,oneof=debug info warn error fatal"`
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Service   string `json:"service"`
	RequestID string `json:"request_id,omitempty"`
	UserID    *uint  `json:"user_id,omitempty"`
	Error     string `json:"error,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
}