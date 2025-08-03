package middleware

import (
	"contact-service/pkg/errors"
	"contact-service/pkg/logger"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// ErrorHandler provides centralized error handling middleware
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Process any errors that occurred during request handling
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			
			// Convert error to AppError
			appErr := convertToAppError(err, c)
			
			// Enrich error with request context
			enrichErrorWithContext(appErr, c)
			
			// Log the error if not already logged
			if !isErrorLogged(c) {
				logError(appErr, c)
			}
			
			// Track error metrics
			trackErrorMetrics(appErr, c)
			
			// Send error response
			sendErrorResponse(appErr, c)
		}
	}
}

// GlobalErrorHandler handles errors that bubble up to the top level
func GlobalErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Handle panic as critical error
				panicErr := errors.NewAppError(
					errors.ErrCodeInternal,
					"System panic occurred",
					fmt.Errorf("%v", err),
				)
				
				enrichErrorWithContext(panicErr, c)
				panicErr.Severity = errors.SeverityCritical
				
				logger.Error("PANIC in global error handler", nil, map[string]interface{}{
					"panic": err,
					"path":  c.Request.URL.Path,
				})
				
				sendErrorResponse(panicErr, c)
			}
		}()
		
		c.Next()
	}
}

// ValidationErrorHandler specifically handles validation errors
func ValidationErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		// Look for validation errors specifically
		for _, ginErr := range c.Errors {
			if validatorErr, ok := ginErr.Err.(validator.ValidationErrors); ok {
				appErr := convertValidationErrors(validatorErr)
				enrichErrorWithContext(appErr, c)
				sendErrorResponse(appErr, c)
				return
			}
		}
	}
}

// DatabaseErrorHandler handles database-specific errors
func DatabaseErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		// Check for database errors
		for _, ginErr := range c.Errors {
			if isDatabaseError(ginErr.Err) {
				appErr := convertDatabaseError(ginErr.Err)
				enrichErrorWithContext(appErr, c)
				sendErrorResponse(appErr, c)
				return
			}
		}
	}
}

// TimeoutHandler wraps requests with timeout handling
func TimeoutHandler(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create timeout context
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		
		// Replace request context
		c.Request = c.Request.WithContext(ctx)
		
		// Channel to signal completion
		done := make(chan struct{})
		
		go func() {
			defer func() {
				if err := recover(); err != nil {
					// Handle panic in timeout context
					panicErr := errors.NewAppError(
						errors.ErrCodeInternal,
						"Panic during request processing",
						fmt.Errorf("%v", err),
					)
					enrichErrorWithContext(panicErr, c)
					c.Error(panicErr)
				}
				close(done)
			}()
			
			c.Next()
		}()
		
		select {
		case <-done:
			// Request completed normally
			return
		case <-ctx.Done():
			// Request timed out
			timeoutErr := errors.NewTimeoutError(
				fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path),
				timeout,
			)
			enrichErrorWithContext(timeoutErr, c)
			sendErrorResponse(timeoutErr, c)
			c.Abort()
		}
	}
}

// RateLimitErrorHandler handles rate limiting errors
func RateLimitErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		// Check if rate limit was exceeded (set by rate limiter middleware)
		if rateLimited, exists := c.Get("rate_limited"); exists && rateLimited.(bool) {
			retryAfter := 1 * time.Minute // Default retry after
			if retry, exists := c.Get("retry_after"); exists {
				if duration, ok := retry.(time.Duration); ok {
					retryAfter = duration
				}
			}
			
			rateLimitErr := errors.NewRateLimitError(retryAfter)
			enrichErrorWithContext(rateLimitErr, c)
			sendErrorResponse(rateLimitErr, c)
			c.Abort()
		}
	}
}

// Helper functions

// convertToAppError converts various error types to AppError
func convertToAppError(err error, c *gin.Context) *errors.AppError {
	// If already an AppError, return as-is
	if appErr, ok := err.(*errors.AppError); ok {
		return appErr
	}
	
	// Handle specific error types
	if validatorErr, ok := err.(validator.ValidationErrors); ok {
		return convertValidationErrors(validatorErr)
	}
	
	switch {
	case isDatabaseError(err):
		return convertDatabaseError(err)
	case isTimeoutError(err):
		return errors.NewTimeoutError("Request timeout", 30*time.Second)
	case isNotFoundError(err):
		return errors.NewNotFoundError("Resource", "unknown")
	}
	
	// Default to internal error
	return errors.NewInternalError("An unexpected error occurred", err)
}

// convertValidationErrors converts validator errors to AppError
func convertValidationErrors(validatorErrs validator.ValidationErrors) *errors.AppError {
	appErr := errors.NewValidationError("Validation failed")
	
	for _, err := range validatorErrs {
		fieldErr := errors.FieldError{
			Field:   err.Field(),
			Message: getValidationErrorMessage(err),
			Code:    err.Tag(),
			Value:   err.Value(),
		}
		appErr.FieldErrors = append(appErr.FieldErrors, fieldErr)
	}
	
	return appErr
}

// convertDatabaseError converts database errors to AppError
func convertDatabaseError(err error) *errors.AppError {
	switch {
	case err == gorm.ErrRecordNotFound:
		return errors.NewNotFoundError("Record", "unknown")
	case strings.Contains(err.Error(), "connection"):
		return errors.NewAppError(errors.ErrCodeDatabaseConnection, "Database connection error", err)
	case strings.Contains(err.Error(), "constraint"):
		return errors.NewAppError(errors.ErrCodeDatabaseConstraint, "Database constraint violation", err)
	case strings.Contains(err.Error(), "transaction"):
		return errors.NewAppError(errors.ErrCodeDatabaseTransaction, "Database transaction error", err)
	default:
		return errors.NewDatabaseError("unknown", err)
	}
}

// enrichErrorWithContext adds request context to error
func enrichErrorWithContext(appErr *errors.AppError, c *gin.Context) {
	// Add request information
	appErr.Endpoint = c.Request.URL.Path
	appErr.Method = c.Request.Method
	appErr.IPAddress = c.ClientIP()
	appErr.UserAgent = c.Request.UserAgent()
	
	// Add request ID if available
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			appErr.RequestID = id
		}
	}
	
	// Add user ID if available
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uint); ok {
			appErr.UserID = &id
		}
	}
	
	// Add additional context
	appErr.WithContext("query_params", c.Request.URL.RawQuery).
		WithContext("content_type", c.Request.Header.Get("Content-Type")).
		WithContext("content_length", c.Request.ContentLength)
}

// logError logs the error with appropriate level
func logError(appErr *errors.AppError, c *gin.Context) {
	fields := map[string]interface{}{
		"error_code":     appErr.Code,
		"error_category": appErr.Category,
		"error_severity": appErr.Severity,
		"request_id":     appErr.RequestID,
		"endpoint":       appErr.Endpoint,
		"method":         appErr.Method,
		"user_id":        appErr.UserID,
		"ip_address":     appErr.IPAddress,
	}
	
	// Add context and metadata
	for k, v := range appErr.Context {
		fields["ctx_"+k] = v
	}
	for k, v := range appErr.Metadata {
		fields["meta_"+k] = v
	}
	
	// Log based on severity
	switch appErr.Severity {
	case errors.SeverityCritical:
		logger.Error("CRITICAL ERROR", appErr.InternalError, fields)
	case errors.SeverityHigh:
		logger.Error("High severity error", appErr.InternalError, fields)
	case errors.SeverityMedium:
		logger.Warn("Medium severity error", fields)
	case errors.SeverityLow:
		logger.Info("Low severity error", fields)
	default:
		logger.Error("Error occurred", appErr.InternalError, fields)
	}
}

// trackErrorMetrics tracks error metrics for monitoring
func trackErrorMetrics(appErr *errors.AppError, c *gin.Context) {
	// Log performance metric for error tracking
	logger.LogPerformanceMetric(
		"error_count",
		1,
		"count",
		map[string]string{
			"error_code":     string(appErr.Code),
			"error_category": string(appErr.Category),
			"error_severity": string(appErr.Severity),
			"endpoint":       c.FullPath(),
			"method":         c.Request.Method,
		},
	)
	
	// Track in context for potential aggregation
	c.Set("error_tracked", true)
}

// sendErrorResponse sends the error response to client
func sendErrorResponse(appErr *errors.AppError, c *gin.Context) {
	// Don't send response if already sent
	if c.Writer.Written() {
		return
	}
	
	// Set appropriate headers
	c.Header("Content-Type", "application/json")
	
	// Add retry headers for retryable errors
	if appErr.Retryable && appErr.RetryAfter != nil {
		c.Header("Retry-After", strconv.Itoa(int(appErr.RetryAfter.Seconds())))
	}
	
	// Send response
	c.JSON(appErr.HTTPStatus, appErr.ToHTTPResponse())
	c.Abort()
}

// Helper functions for error type detection

func isDatabaseError(err error) bool {
	if err == nil {
		return false
	}
	
	errorStr := err.Error()
	return err == gorm.ErrRecordNotFound ||
		strings.Contains(errorStr, "database") ||
		strings.Contains(errorStr, "sql") ||
		strings.Contains(errorStr, "connection") ||
		strings.Contains(errorStr, "constraint")
}

func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	
	errorStr := err.Error()
	return strings.Contains(errorStr, "timeout") ||
		strings.Contains(errorStr, "deadline exceeded") ||
		strings.Contains(errorStr, "context canceled")
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	
	errorStr := err.Error()
	return err == gorm.ErrRecordNotFound ||
		strings.Contains(errorStr, "not found") ||
		strings.Contains(errorStr, "does not exist")
}

func isErrorLogged(c *gin.Context) bool {
	logged, exists := c.Get("error_logged")
	return exists && logged.(bool)
}

func getValidationErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", err.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", err.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", err.Field(), err.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", err.Field(), err.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", err.Field(), err.Param())
	case "numeric":
		return fmt.Sprintf("%s must be a number", err.Field())
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", err.Field())
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", err.Field())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", err.Field())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", err.Field())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", err.Field(), err.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", err.Field(), err.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", err.Field(), err.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", err.Field(), err.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", err.Field(), err.Param())
	default:
		return fmt.Sprintf("%s is invalid", err.Field())
	}
}

