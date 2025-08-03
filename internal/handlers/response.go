package handlers

import (
	"math"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// APIResponse represents the standard API response structure
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Meta      interface{} `json:"meta,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// ErrorInfo represents error information in API responses
type ErrorInfo struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Page         int   `json:"page"`
	PageSize     int   `json:"page_size"`
	Total        int64 `json:"total"`
	TotalPages   int   `json:"total_pages"`
	HasNextPage  bool  `json:"has_next_page"`
	HasPrevPage  bool  `json:"has_prev_page"`
	NextPage     *int  `json:"next_page,omitempty"`
	PrevPage     *int  `json:"prev_page,omitempty"`
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(message string, data interface{}) *APIResponse {
	return &APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(message, details string) *APIResponse {
	return &APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Message: message,
			Details: details,
		},
		Timestamp: time.Now(),
	}
}

// NewErrorResponseWithCode creates a new error response with error code
func NewErrorResponseWithCode(code, message, details string) *APIResponse {
	return &APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now(),
	}
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse(message string, data interface{}, meta *PaginationMeta) *APIResponse {
	return &APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Meta:      meta,
		Timestamp: time.Now(),
	}
}

// NewPaginationMeta creates pagination metadata
func NewPaginationMeta(page, pageSize int, total int64) *PaginationMeta {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	
	meta := &PaginationMeta{
		Page:         page,
		PageSize:     pageSize,
		Total:        total,
		TotalPages:   totalPages,
		HasNextPage:  page < totalPages,
		HasPrevPage:  page > 1,
	}

	if meta.HasNextPage {
		nextPage := page + 1
		meta.NextPage = &nextPage
	}

	if meta.HasPrevPage {
		prevPage := page - 1
		meta.PrevPage = &prevPage
	}

	return meta
}

// NewValidationErrorResponse creates a validation error response
func NewValidationErrorResponse(details string) *APIResponse {
	return NewErrorResponseWithCode("VALIDATION_ERROR", "Validation failed", details)
}

// NewNotFoundResponse creates a not found error response
func NewNotFoundResponse(resource string) *APIResponse {
	return NewErrorResponseWithCode("NOT_FOUND", resource+" not found", "")
}

// NewConflictResponse creates a conflict error response
func NewConflictResponse(message string) *APIResponse {
	return NewErrorResponseWithCode("CONFLICT", message, "")
}

// NewUnauthorizedResponse creates an unauthorized error response
func NewUnauthorizedResponse() *APIResponse {
	return NewErrorResponseWithCode("UNAUTHORIZED", "Authentication required", "")
}

// NewForbiddenResponse creates a forbidden error response
func NewForbiddenResponse() *APIResponse {
	return NewErrorResponseWithCode("FORBIDDEN", "Access denied", "")
}

// NewInternalErrorResponse creates an internal server error response
func NewInternalErrorResponse() *APIResponse {
	return NewErrorResponseWithCode("INTERNAL_ERROR", "Internal server error", "")
}

// NewRateLimitResponse creates a rate limit error response
func NewRateLimitResponse() *APIResponse {
	return NewErrorResponseWithCode("RATE_LIMIT", "Rate limit exceeded", "")
}

// StatsResponse represents statistics response structure
type StatsResponse struct {
	Total       int64                  `json:"total"`
	Breakdowns  map[string]interface{} `json:"breakdowns"`
	Trends      map[string]interface{} `json:"trends,omitempty"`
	Comparisons map[string]interface{} `json:"comparisons,omitempty"`
}

// NewStatsResponse creates a new statistics response
func NewStatsResponse(message string, total int64, breakdowns map[string]interface{}) *APIResponse {
	stats := &StatsResponse{
		Total:      total,
		Breakdowns: breakdowns,
	}
	
	return NewSuccessResponse(message, stats)
}

// HealthResponse represents health check response structure
type HealthResponse struct {
	Status      string                 `json:"status"`
	Version     string                 `json:"version"`
	Timestamp   time.Time              `json:"timestamp"`
	Uptime      string                 `json:"uptime"`
	Environment string                 `json:"environment"`
	Services    map[string]interface{} `json:"services"`
}

// NewHealthResponse creates a new health check response
func NewHealthResponse(status, version, environment string, uptime time.Duration, services map[string]interface{}) *APIResponse {
	health := &HealthResponse{
		Status:      status,
		Version:     version,
		Timestamp:   time.Now(),
		Uptime:      uptime.String(),
		Environment: environment,
		Services:    services,
	}
	
	return &APIResponse{
		Success:   status == "healthy",
		Message:   "Health check completed",
		Data:      health,
		Timestamp: time.Now(),
	}
}

// PaginatedResponse represents a paginated response structure
type PaginatedResponse struct {
	Items      interface{}     `json:"items"`
	Pagination *PaginationMeta `json:"pagination"`
}

// NewPaginatedResponseWithItems creates a new paginated response with items and pagination metadata
func NewPaginatedResponseWithItems(items interface{}, total int, page int, limit int) *PaginatedResponse {
	meta := NewPaginationMeta(page, limit, int64(total))
	return &PaginatedResponse{
		Items:      items,
		Pagination: meta,
	}
}

// parsePaginationParams parses pagination parameters from gin context
func parsePaginationParams(c *gin.Context) (page int, limit int) {
	// Default values
	page = 1
	limit = 20

	// Parse page parameter
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// Parse limit parameter
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	return page, limit
}