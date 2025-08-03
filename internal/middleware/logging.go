package middleware

import (
	"bytes"
	"contact-service/pkg/errors"
	"contact-service/pkg/logger"
	"encoding/json"
	"io"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestLogger middleware logs all HTTP requests with detailed information
func RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Extract user ID from context if available
		var userID *uint
		if param.Keys != nil {
			if uid, exists := param.Keys["user_id"]; exists {
				if id, ok := uid.(uint); ok {
					userID = &id
				}
			}
		}

		// Log the API request
		logger.LogAPIRequest(
			param.Method,
			param.Path,
			userID,
			param.Latency,
			param.StatusCode,
		)

		// Return empty string as we handle logging internally
		return ""
	})
}

// DetailedRequestLogger provides comprehensive request/response logging
func DetailedRequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID
		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		// Start time
		startTime := time.Now()

		// Create response writer wrapper to capture response body
		responseWriter := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = responseWriter

		// Capture request body if it exists
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime)

		// Extract user ID from context
		var userID *uint
		if uid, exists := c.Get("user_id"); exists {
			if id, ok := uid.(uint); ok {
				userID = &id
			}
		}

		// Log detailed request information
		fields := map[string]interface{}{
			"request_id":    requestID,
			"method":        c.Request.Method,
			"path":          c.Request.URL.Path,
			"query":         c.Request.URL.RawQuery,
			"status_code":   c.Writer.Status(),
			"duration_ms":   duration.Milliseconds(),
			"client_ip":     c.ClientIP(),
			"user_agent":    c.Request.UserAgent(),
			"referer":       c.Request.Referer(),
			"content_type":  c.Request.Header.Get("Content-Type"),
			"content_length": c.Request.ContentLength,
			"response_size": responseWriter.body.Len(),
		}

		if userID != nil {
			fields["user_id"] = *userID
		}

		// Add request headers (filtered)
		requestHeaders := make(map[string]string)
		for key, values := range c.Request.Header {
			// Filter sensitive headers
			if !isSensitiveHeader(key) {
				requestHeaders[key] = values[0]
			}
		}
		fields["request_headers"] = requestHeaders

		// Add request body for non-GET requests (truncated if too large)
		if c.Request.Method != "GET" && len(requestBody) > 0 {
			if len(requestBody) > 1000 {
				fields["request_body"] = string(requestBody[:1000]) + "...[truncated]"
			} else {
				// Try to parse as JSON for better formatting
				var jsonBody interface{}
				if err := json.Unmarshal(requestBody, &jsonBody); err == nil {
					fields["request_body"] = jsonBody
				} else {
					fields["request_body"] = string(requestBody)
				}
			}
		}

		// Add response body for errors or if explicitly enabled
		if c.Writer.Status() >= 400 || shouldLogResponseBody(c) {
			responseBody := responseWriter.body.String()
			if len(responseBody) > 0 {
				if len(responseBody) > 1000 {
					fields["response_body"] = responseBody[:1000] + "...[truncated]"
				} else {
					// Try to parse as JSON for better formatting
					var jsonBody interface{}
					if err := json.Unmarshal([]byte(responseBody), &jsonBody); err == nil {
						fields["response_body"] = jsonBody
					} else {
						fields["response_body"] = responseBody
					}
				}
			}
		}

		// Log based on status code
		if c.Writer.Status() >= 500 {
			logger.Error("HTTP request failed with server error", nil, fields)
		} else if c.Writer.Status() >= 400 {
			logger.Warn("HTTP request failed with client error", fields)
		} else if duration > 5*time.Second {
			logger.Warn("Slow HTTP request detected", fields)
		} else {
			logger.Info("HTTP request completed", fields)
		}

		// Log performance metrics
		logger.LogPerformanceMetric(
			"http_request_duration",
			float64(duration.Milliseconds()),
			"ms",
			map[string]string{
				"method":   c.Request.Method,
				"endpoint": c.FullPath(),
				"status":   strconv.Itoa(c.Writer.Status()),
			},
		)
	}
}

// ErrorLoggingMiddleware logs errors with comprehensive context
func ErrorLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			requestID, _ := c.Get("request_id")
			userID, _ := c.Get("user_id")

			for _, ginErr := range c.Errors {
				fields := map[string]interface{}{
					"request_id": requestID,
					"method":     c.Request.Method,
					"path":       c.Request.URL.Path,
					"status":     c.Writer.Status(),
					"client_ip":  c.ClientIP(),
					"user_agent": c.Request.UserAgent(),
				}

				if userID != nil {
					fields["user_id"] = userID
				}

				// Handle different error types
				switch err := ginErr.Err.(type) {
				case *errors.AppError:
					// Enhanced logging for AppError
					fields["error_code"] = err.Code
					fields["error_category"] = err.Category
					fields["error_severity"] = err.Severity
					fields["retryable"] = err.Retryable
					
					if err.Context != nil {
						for k, v := range err.Context {
							fields["ctx_"+k] = v
						}
					}

					if err.Metadata != nil {
						for k, v := range err.Metadata {
							fields["meta_"+k] = v
						}
					}

					logger.Error("Application error occurred", err, fields)
				default:
					// Standard error logging
					logger.Error("Request processing error", err, fields)
				}
			}
		}
	}
}

// PanicRecoveryMiddleware recovers from panics and logs them
func PanicRecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID, _ := c.Get("request_id")
		userID, _ := c.Get("user_id")

		fields := map[string]interface{}{
			"request_id": requestID,
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"client_ip":  c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
			"panic":      recovered,
		}

		if userID != nil {
			fields["user_id"] = userID
		}

		// Create a critical error
		panicErr := errors.NewAppError(
			errors.ErrCodeInternal,
			"System panic occurred",
			nil,
		).WithContext("panic_value", recovered).
			WithMetadata("request_id", requestID)

		if userID != nil {
			if uid, ok := userID.(uint); ok {
				panicErr.UserID = &uid
			}
		}

		panicErr.Endpoint = c.Request.URL.Path
		panicErr.Method = c.Request.Method
		panicErr.IPAddress = c.ClientIP()
		panicErr.UserAgent = c.Request.UserAgent()
		panicErr.Severity = errors.SeverityCritical

		logger.Error("PANIC RECOVERED", nil, fields)

		// Log as security event as well (panics might indicate attacks)
		logger.LogSecurityEvent(
			"panic_recovered",
			panicErr.UserID,
			c.ClientIP(),
			map[string]interface{}{
				"panic_value": recovered,
				"endpoint":    c.Request.URL.Path,
				"method":      c.Request.Method,
			},
		)

		// Return error response
		c.JSON(panicErr.HTTPStatus, panicErr.ToHTTPResponse())
	})
}

// SecurityEventLogger logs security-related events
func SecurityEventLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Capture original values
		originalStatus := c.Writer.Status()
		
		c.Next()

		// Log security events based on response
		userID, _ := c.Get("user_id")
		var uid *uint
		if userID != nil {
			if id, ok := userID.(uint); ok {
				uid = &id
			}
		}

		// Failed authentication attempts
		if c.Writer.Status() == 401 {
			logger.LogSecurityEvent(
				"authentication_failed",
				uid,
				c.ClientIP(),
				map[string]interface{}{
					"endpoint":   c.Request.URL.Path,
					"method":     c.Request.Method,
					"user_agent": c.Request.UserAgent(),
				},
			)
		}

		// Access denied events
		if c.Writer.Status() == 403 {
			logger.LogSecurityEvent(
				"access_denied",
				uid,
				c.ClientIP(),
				map[string]interface{}{
					"endpoint":   c.Request.URL.Path,
					"method":     c.Request.Method,
					"user_agent": c.Request.UserAgent(),
				},
			)
		}

		// Suspicious activity (multiple 4xx errors)
		if c.Writer.Status() >= 400 && c.Writer.Status() < 500 {
			// This could be enhanced with rate limiting logic
			// For now, just log the event
			if c.Writer.Status() != originalStatus {
				logger.LogSecurityEvent(
					"suspicious_activity",
					uid,
					c.ClientIP(),
					map[string]interface{}{
						"endpoint":    c.Request.URL.Path,
						"method":      c.Request.Method,
						"status_code": c.Writer.Status(),
						"user_agent":  c.Request.UserAgent(),
					},
				)
			}
		}
	}
}

// DatabaseOperationLogger logs database operations from context
func DatabaseOperationLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check for database operation context
		if operations, exists := c.Get("db_operations"); exists {
			if ops, ok := operations.([]map[string]interface{}); ok {
				for _, op := range ops {
					operation := op["operation"].(string)
					table := op["table"].(string)
					duration := op["duration"].(time.Duration)
					recordID := op["record_id"]
					err := op["error"]
					
					var dbErr error
					if err != nil {
						dbErr = err.(error)
					}

					logger.LogDatabaseOperation(operation, table, recordID, duration, dbErr)
				}
			}
		}
	}
}

// BusinessEventLogger logs business events from context
func BusinessEventLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check for business events in context
		if events, exists := c.Get("business_events"); exists {
			if eventList, ok := events.([]map[string]interface{}); ok {
				for _, event := range eventList {
					eventName := event["event"].(string)
					entityType := event["entity_type"].(string)
					entityID := event["entity_id"]
					details := event["details"].(map[string]interface{})

					logger.LogBusinessEvent(eventName, entityType, entityID, details)
				}
			}
		}
	}
}

// responseBodyWriter wraps gin.ResponseWriter to capture response body
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Helper functions

// isSensitiveHeader checks if a header contains sensitive information
func isSensitiveHeader(header string) bool {
	sensitiveHeaders := map[string]bool{
		"authorization": true,
		"cookie":        true,
		"set-cookie":    true,
		"x-api-key":     true,
		"x-auth-token":  true,
	}
	
	return sensitiveHeaders[header] || sensitiveHeaders[header]
}

// shouldLogResponseBody determines if response body should be logged
func shouldLogResponseBody(c *gin.Context) bool {
	// Log response body for debugging in development
	if gin.Mode() == gin.DebugMode {
		return true
	}

	// Check for explicit logging flag
	if logResponse, exists := c.Get("log_response_body"); exists {
		if log, ok := logResponse.(bool); ok {
			return log
		}
	}

	return false
}

// SetLogResponseBody sets a flag to log response body for this request
func SetLogResponseBody(c *gin.Context) {
	c.Set("log_response_body", true)
}

// AddDatabaseOperation adds a database operation to be logged
func AddDatabaseOperation(c *gin.Context, operation, table string, recordID interface{}, duration time.Duration, err error) {
	op := map[string]interface{}{
		"operation": operation,
		"table":     table,
		"record_id": recordID,
		"duration":  duration,
		"error":     err,
	}

	operations, exists := c.Get("db_operations")
	if !exists {
		operations = []map[string]interface{}{op}
	} else {
		if ops, ok := operations.([]map[string]interface{}); ok {
			operations = append(ops, op)
		}
	}

	c.Set("db_operations", operations)
}

// AddBusinessEvent adds a business event to be logged
func AddBusinessEvent(c *gin.Context, event, entityType string, entityID interface{}, details map[string]interface{}) {
	businessEvent := map[string]interface{}{
		"event":       event,
		"entity_type": entityType,
		"entity_id":   entityID,
		"details":     details,
	}

	events, exists := c.Get("business_events")
	if !exists {
		events = []map[string]interface{}{businessEvent}
	} else {
		if eventList, ok := events.([]map[string]interface{}); ok {
			events = append(eventList, businessEvent)
		}
	}

	c.Set("business_events", events)
}