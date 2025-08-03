package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	// General errors
	ErrCodeInternal        ErrorCode = "INTERNAL_ERROR"
	ErrCodeValidation      ErrorCode = "VALIDATION_ERROR"
	ErrCodeNotFound        ErrorCode = "NOT_FOUND"
	ErrCodeUnauthorized    ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden       ErrorCode = "FORBIDDEN"
	ErrCodeConflict        ErrorCode = "CONFLICT"
	ErrCodeRateLimit       ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrCodeTimeout         ErrorCode = "TIMEOUT"
	ErrCodeBadRequest      ErrorCode = "BAD_REQUEST"

	// Database errors
	ErrCodeDatabaseConnection ErrorCode = "DATABASE_CONNECTION_ERROR"
	ErrCodeDatabaseQuery      ErrorCode = "DATABASE_QUERY_ERROR"
	ErrCodeDatabaseConstraint ErrorCode = "DATABASE_CONSTRAINT_ERROR"
	ErrCodeDatabaseTransaction ErrorCode = "DATABASE_TRANSACTION_ERROR"

	// Business logic errors
	ErrCodeContactNotFound     ErrorCode = "CONTACT_NOT_FOUND"
	ErrCodeContactExists       ErrorCode = "CONTACT_ALREADY_EXISTS"
	ErrCodeInvalidStatus       ErrorCode = "INVALID_STATUS_TRANSITION"
	ErrCodeAssignmentFailed    ErrorCode = "ASSIGNMENT_FAILED"
	ErrCodeSchedulingConflict  ErrorCode = "SCHEDULING_CONFLICT"
	ErrCodeAppointmentNotFound ErrorCode = "APPOINTMENT_NOT_FOUND"
	ErrCodeInvalidPermission   ErrorCode = "INVALID_PERMISSION"

	// External service errors
	ErrCodeEmailService    ErrorCode = "EMAIL_SERVICE_ERROR"
	ErrCodeCalendarService ErrorCode = "CALENDAR_SERVICE_ERROR"
	ErrCodeFileService     ErrorCode = "FILE_SERVICE_ERROR"
	ErrCodeNotificationService ErrorCode = "NOTIFICATION_SERVICE_ERROR"

	// Authentication/Authorization errors
	ErrCodeInvalidToken     ErrorCode = "INVALID_TOKEN"
	ErrCodeTokenExpired     ErrorCode = "TOKEN_EXPIRED"
	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	ErrCodeAccountLocked    ErrorCode = "ACCOUNT_LOCKED"
	ErrCodeInsufficientRole ErrorCode = "INSUFFICIENT_ROLE"
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity string

const (
	SeverityLow      ErrorSeverity = "low"
	SeverityMedium   ErrorSeverity = "medium"
	SeverityHigh     ErrorSeverity = "high"
	SeverityCritical ErrorSeverity = "critical"
)

// ErrorCategory represents the category of an error for better organization
type ErrorCategory string

const (
	CategoryValidation  ErrorCategory = "validation"
	CategoryDatabase    ErrorCategory = "database"
	CategoryBusiness    ErrorCategory = "business"
	CategorySecurity    ErrorCategory = "security"
	CategoryIntegration ErrorCategory = "integration"
	CategorySystem      ErrorCategory = "system"
)

// AppError represents a comprehensive application error
type AppError struct {
	// Core error information
	Code        ErrorCode     `json:"code"`
	Message     string        `json:"message"`
	Description string        `json:"description,omitempty"`
	Severity    ErrorSeverity `json:"severity"`
	Category    ErrorCategory `json:"category"`

	// HTTP information
	HTTPStatus int `json:"http_status"`

	// Context information
	UserID     *uint                  `json:"user_id,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
	Endpoint   string                 `json:"endpoint,omitempty"`
	Method     string                 `json:"method,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`

	// Technical details
	InternalError  error               `json:"-"`
	StackTrace     string              `json:"stack_trace,omitempty"`
	Timestamp      time.Time           `json:"timestamp"`
	Source         string              `json:"source,omitempty"` // File and line where error occurred
	Retryable      bool                `json:"retryable"`
	RetryAfter     *time.Duration      `json:"retry_after,omitempty"`

	// Related errors
	FieldErrors    []FieldError        `json:"field_errors,omitempty"`
	RelatedErrors  []AppError          `json:"related_errors,omitempty"`

	// Metadata
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Tags           []string            `json:"tags,omitempty"`
}

// FieldError represents validation errors for specific fields
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
	Value   interface{} `json:"value,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// WithContext adds context information to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithMetadata adds metadata to the error
func (e *AppError) WithMetadata(key string, value interface{}) *AppError {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// WithTag adds a tag to the error
func (e *AppError) WithTag(tag string) *AppError {
	e.Tags = append(e.Tags, tag)
	return e
}

// WithFieldError adds a field validation error
func (e *AppError) WithFieldError(field, message, code string, value interface{}) *AppError {
	e.FieldErrors = append(e.FieldErrors, FieldError{
		Field:   field,
		Message: message,
		Code:    code,
		Value:   value,
	})
	return e
}

// WithRetry marks the error as retryable with optional delay
func (e *AppError) WithRetry(retryAfter *time.Duration) *AppError {
	e.Retryable = true
	e.RetryAfter = retryAfter
	return e
}

// ToJSON converts the error to JSON representation
func (e *AppError) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// HTTPResponse represents the HTTP response format for errors
type HTTPResponse struct {
	Success   bool        `json:"success"`
	Error     *AppError   `json:"error"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	Path      string      `json:"path,omitempty"`
	Method    string      `json:"method,omitempty"`
}

// ToHTTPResponse converts AppError to HTTP response format
func (e *AppError) ToHTTPResponse() *HTTPResponse {
	return &HTTPResponse{
		Success:   false,
		Error:     e,
		RequestID: e.RequestID,
		Timestamp: e.Timestamp,
		Path:      e.Endpoint,
		Method:    e.Method,
	}
}

// Constructor functions for common errors

// NewAppError creates a new application error with stack trace
func NewAppError(code ErrorCode, message string, internalErr error) *AppError {
	// Capture stack trace
	stackTrace := captureStackTrace(2)
	source := getSource(2)

	err := &AppError{
		Code:          code,
		Message:       message,
		InternalError: internalErr,
		StackTrace:    stackTrace,
		Source:        source,
		Timestamp:     time.Now(),
		HTTPStatus:    getHTTPStatusForCode(code),
		Severity:      getSeverityForCode(code),
		Category:      getCategoryForCode(code),
		Context:       make(map[string]interface{}),
		Metadata:      make(map[string]interface{}),
	}

	return err
}

// NewValidationError creates a validation error
func NewValidationError(message string, fieldErrors ...FieldError) *AppError {
	err := NewAppError(ErrCodeValidation, message, nil)
	err.FieldErrors = fieldErrors
	return err
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string, identifier interface{}) *AppError {
	return NewAppError(
		ErrCodeNotFound,
		fmt.Sprintf("%s not found", resource),
		nil,
	).WithContext("resource", resource).WithContext("identifier", identifier)
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *AppError {
	if message == "" {
		message = "Authentication required"
	}
	return NewAppError(ErrCodeUnauthorized, message, nil)
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) *AppError {
	if message == "" {
		message = "Access forbidden"
	}
	return NewAppError(ErrCodeForbidden, message, nil)
}

// NewConflictError creates a conflict error
func NewConflictError(message string) *AppError {
	return NewAppError(ErrCodeConflict, message, nil)
}

// NewInternalError creates an internal server error
func NewInternalError(message string, internalErr error) *AppError {
	if message == "" {
		message = "Internal server error"
	}
	return NewAppError(ErrCodeInternal, message, internalErr)
}

// NewDatabaseError creates a database error
func NewDatabaseError(operation string, internalErr error) *AppError {
	return NewAppError(
		ErrCodeDatabaseQuery,
		fmt.Sprintf("Database operation failed: %s", operation),
		internalErr,
	).WithContext("operation", operation)
}

// NewBusinessError creates a business logic error
func NewBusinessError(code ErrorCode, message string) *AppError {
	return NewAppError(code, message, nil)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(operation string, timeout time.Duration) *AppError {
	return NewAppError(
		ErrCodeTimeout,
		fmt.Sprintf("Operation timed out: %s", operation),
		nil,
	).WithContext("operation", operation).
		WithContext("timeout", timeout.String()).
		WithRetry(&timeout)
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(retryAfter time.Duration) *AppError {
	return NewAppError(
		ErrCodeRateLimit,
		"Rate limit exceeded",
		nil,
	).WithRetry(&retryAfter)
}

// Helper functions

// captureStackTrace captures the current stack trace
func captureStackTrace(skip int) string {
	const maxStackSize = 50
	stack := make([]uintptr, maxStackSize)
	length := runtime.Callers(skip, stack[:])
	
	if length == 0 {
		return "No stack trace available"
	}

	frames := runtime.CallersFrames(stack[:length])
	var stackTrace string

	for {
		frame, more := frames.Next()
		stackTrace += fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
		
		if !more {
			break
		}
	}

	return stackTrace
}

// getSource returns the file and line where the error occurred
func getSource(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	return fmt.Sprintf("%s:%d", file, line)
}

// Map error codes to HTTP status codes
func getHTTPStatusForCode(code ErrorCode) int {
	switch code {
	case ErrCodeValidation, ErrCodeBadRequest:
		return http.StatusBadRequest
	case ErrCodeUnauthorized, ErrCodeInvalidToken, ErrCodeTokenExpired, ErrCodeInvalidCredentials:
		return http.StatusUnauthorized
	case ErrCodeForbidden, ErrCodeInsufficientRole, ErrCodeAccountLocked:
		return http.StatusForbidden
	case ErrCodeNotFound, ErrCodeContactNotFound, ErrCodeAppointmentNotFound:
		return http.StatusNotFound
	case ErrCodeConflict, ErrCodeContactExists, ErrCodeSchedulingConflict:
		return http.StatusConflict
	case ErrCodeRateLimit:
		return http.StatusTooManyRequests
	case ErrCodeTimeout:
		return http.StatusRequestTimeout
	default:
		return http.StatusInternalServerError
	}
}

// Map error codes to severity levels
func getSeverityForCode(code ErrorCode) ErrorSeverity {
	switch code {
	case ErrCodeValidation, ErrCodeBadRequest, ErrCodeNotFound:
		return SeverityLow
	case ErrCodeUnauthorized, ErrCodeForbidden, ErrCodeConflict:
		return SeverityMedium
	case ErrCodeDatabaseConnection, ErrCodeDatabaseTransaction, ErrCodeInternal:
		return SeverityHigh
	case ErrCodeAccountLocked:
		return SeverityCritical
	default:
		return SeverityMedium
	}
}

// Map error codes to categories
func getCategoryForCode(code ErrorCode) ErrorCategory {
	switch code {
	case ErrCodeValidation, ErrCodeBadRequest:
		return CategoryValidation
	case ErrCodeDatabaseConnection, ErrCodeDatabaseQuery, ErrCodeDatabaseConstraint, ErrCodeDatabaseTransaction:
		return CategoryDatabase
	case ErrCodeContactNotFound, ErrCodeContactExists, ErrCodeInvalidStatus, ErrCodeAssignmentFailed, ErrCodeSchedulingConflict:
		return CategoryBusiness
	case ErrCodeUnauthorized, ErrCodeForbidden, ErrCodeInvalidToken, ErrCodeTokenExpired, ErrCodeAccountLocked:
		return CategorySecurity
	case ErrCodeEmailService, ErrCodeCalendarService, ErrCodeFileService, ErrCodeNotificationService:
		return CategoryIntegration
	default:
		return CategorySystem
	}
}

// WrapError wraps an existing error as an AppError
func WrapError(err error, code ErrorCode, message string) *AppError {
	if err == nil {
		return nil
	}

	// If it's already an AppError, return it
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}

	return NewAppError(code, message, err)
}

// IsRetryable checks if an error is retryable
func IsRetryableError(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Retryable
	}
	return false
}

// GetRetryAfter gets the retry delay from an error
func GetRetryAfter(err error) *time.Duration {
	if appErr, ok := err.(*AppError); ok {
		return appErr.RetryAfter
	}
	return nil
}

// ErrorMetrics represents error metrics for monitoring
type ErrorMetrics struct {
	ErrorCode  ErrorCode     `json:"error_code"`
	Count      int           `json:"count"`
	Severity   ErrorSeverity `json:"severity"`
	Category   ErrorCategory `json:"category"`
	LastSeen   time.Time     `json:"last_seen"`
	Endpoints  []string      `json:"endpoints"`
}

// ErrorStats tracks error statistics
type ErrorStats struct {
	Total      int                      `json:"total"`
	ByCode     map[ErrorCode]int        `json:"by_code"`
	BySeverity map[ErrorSeverity]int    `json:"by_severity"`
	ByCategory map[ErrorCategory]int    `json:"by_category"`
	Recent     []AppError               `json:"recent"`
	Metrics    []ErrorMetrics           `json:"metrics"`
}