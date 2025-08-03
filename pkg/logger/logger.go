package logger

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *logrus.Logger

// LogConfig holds logging configuration
type LogConfig struct {
	Level      string
	Format     string
	File       bool
	FilePath   string
	MaxSize    int // MB
	MaxBackups int
	MaxAge     int // days
	Compress   bool
}

// InitLogger initializes the global logger
func InitLogger() {
	config := loadLogConfig()
	
	Logger = logrus.New()
	
	// Set log level
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	Logger.SetLevel(level)
	
	// Set formatter
	if config.Format == "json" {
		Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "function",
			},
		})
	} else {
		Logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     false,
		})
	}
	
	// Set output
	var writers []io.Writer
	writers = append(writers, os.Stdout)
	
	// Add file output if enabled
	if config.File {
		// Ensure log directory exists
		logDir := filepath.Dir(config.FilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			logrus.Warnf("Failed to create log directory: %v", err)
		} else {
			fileWriter := &lumberjack.Logger{
				Filename:   config.FilePath,
				MaxSize:    config.MaxSize,
				MaxBackups: config.MaxBackups,
				MaxAge:     config.MaxAge,
				Compress:   config.Compress,
				LocalTime:  true,
			}
			writers = append(writers, fileWriter)
		}
	}
	
	// Set multi-writer output
	if len(writers) > 1 {
		Logger.SetOutput(io.MultiWriter(writers...))
	} else {
		Logger.SetOutput(writers[0])
	}
	
	// Add context hook for consistent fields
	Logger.AddHook(&ContextHook{})
	
	Logger.Info("Logger initialized successfully")
}

// loadLogConfig loads logging configuration from environment variables
func loadLogConfig() *LogConfig {
	return &LogConfig{
		Level:      getEnv("LOG_LEVEL", "info"),
		Format:     getEnv("LOG_FORMAT", "text"),
		File:       getEnvBool("LOG_FILE_ENABLED", true),
		FilePath:   getEnv("LOG_FILE_PATH", "./logs/contact-service.log"),
		MaxSize:    getEnvInt("LOG_MAX_SIZE", 100),
		MaxBackups: getEnvInt("LOG_MAX_BACKUPS", 10),
		MaxAge:     getEnvInt("LOG_MAX_AGE", 30),
		Compress:   getEnvBool("LOG_COMPRESS", true),
	}
}

// ContextHook adds consistent context fields to all log entries
type ContextHook struct{}

func (hook *ContextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *ContextHook) Fire(entry *logrus.Entry) error {
	entry.Data["service"] = "contact-service"
	entry.Data["version"] = getEnv("APP_VERSION", "1.0.0")
	entry.Data["environment"] = getEnv("APP_ENV", "development")
	return nil
}

// Structured logging functions

// Info logs an info message with optional fields
func Info(message string, fields ...map[string]interface{}) {
	entry := Logger.WithFields(mergeFields(fields...))
	entry.Info(message)
}

// Error logs an error message with optional fields
func Error(message string, err error, fields ...map[string]interface{}) {
	allFields := mergeFields(fields...)
	if err != nil {
		allFields["error"] = err.Error()
	}
	entry := Logger.WithFields(allFields)
	entry.Error(message)
}

// Warn logs a warning message with optional fields
func Warn(message string, fields ...map[string]interface{}) {
	entry := Logger.WithFields(mergeFields(fields...))
	entry.Warn(message)
}

// Debug logs a debug message with optional fields
func Debug(message string, fields ...map[string]interface{}) {
	entry := Logger.WithFields(mergeFields(fields...))
	entry.Debug(message)
}

// Fatal logs a fatal message and exits
func Fatal(message string, err error, fields ...map[string]interface{}) {
	allFields := mergeFields(fields...)
	if err != nil {
		allFields["error"] = err.Error()
	}
	entry := Logger.WithFields(allFields)
	entry.Fatal(message)
}

// WithFields creates a new log entry with additional fields
func WithFields(fields map[string]interface{}) *logrus.Entry {
	return Logger.WithFields(fields)
}

// WithError creates a new log entry with an error field
func WithError(err error) *logrus.Entry {
	return Logger.WithError(err)
}

// Activity-specific logging functions

// LogContactActivity logs contact-related activities
func LogContactActivity(contactID uint, activity string, details map[string]interface{}) {
	fields := map[string]interface{}{
		"contact_id": contactID,
		"activity":   activity,
		"category":   "contact_activity",
	}
	
	// Merge additional details
	for k, v := range details {
		fields[k] = v
	}
	
	Info("Contact activity recorded", fields)
}

// LogAPIRequest logs incoming API requests
func LogAPIRequest(method, path string, userID *uint, duration time.Duration, statusCode int) {
	fields := map[string]interface{}{
		"method":      method,
		"path":        path,
		"duration_ms": duration.Milliseconds(),
		"status_code": statusCode,
		"category":    "api_request",
	}
	
	if userID != nil {
		fields["user_id"] = *userID
	}
	
	if statusCode >= 400 {
		Error("API request failed", nil, fields)
	} else {
		Info("API request completed", fields)
	}
}

// LogDatabaseOperation logs database operations
func LogDatabaseOperation(operation, table string, recordID interface{}, duration time.Duration, err error) {
	fields := map[string]interface{}{
		"operation":   operation,
		"table":       table,
		"duration_ms": duration.Milliseconds(),
		"category":    "database",
	}
	
	if recordID != nil {
		fields["record_id"] = recordID
	}
	
	if err != nil {
		Error("Database operation failed", err, fields)
	} else {
		Debug("Database operation completed", fields)
	}
}

// LogBusinessEvent logs important business events
func LogBusinessEvent(event string, entityType string, entityID interface{}, details map[string]interface{}) {
	fields := map[string]interface{}{
		"event":       event,
		"entity_type": entityType,
		"category":    "business_event",
	}
	
	if entityID != nil {
		fields["entity_id"] = entityID
	}
	
	// Merge additional details
	for k, v := range details {
		fields[k] = v
	}
	
	Info("Business event occurred", fields)
}

// LogSecurityEvent logs security-related events
func LogSecurityEvent(event string, userID *uint, ipAddress string, details map[string]interface{}) {
	fields := map[string]interface{}{
		"event":      event,
		"ip_address": ipAddress,
		"category":   "security",
		"severity":   "high",
	}
	
	if userID != nil {
		fields["user_id"] = *userID
	}
	
	// Merge additional details
	for k, v := range details {
		fields[k] = v
	}
	
	Warn("Security event detected", fields)
}

// LogPerformanceMetric logs performance metrics
func LogPerformanceMetric(metric string, value float64, unit string, tags map[string]string) {
	fields := map[string]interface{}{
		"metric":   metric,
		"value":    value,
		"unit":     unit,
		"category": "performance",
	}
	
	// Add tags
	for k, v := range tags {
		fields["tag_"+k] = v
	}
	
	Debug("Performance metric recorded", fields)
}

// Utility functions

func mergeFields(fieldMaps ...map[string]interface{}) logrus.Fields {
	result := make(logrus.Fields)
	for _, fields := range fieldMaps {
		for k, v := range fields {
			result[k] = v
		}
	}
	return result
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// Cleanup closes any file handles used by the logger
func Cleanup() {
	if Logger != nil {
		Info("Logger shutting down")
	}
}

// SetLevel dynamically changes the log level
func SetLevel(level string) error {
	parsedLevel, err := logrus.ParseLevel(strings.ToLower(level))
	if err != nil {
		return err
	}
	
	Logger.SetLevel(parsedLevel)
	Info("Log level changed", map[string]interface{}{"new_level": level})
	return nil
}

// GetLevel returns the current log level
func GetLevel() string {
	if Logger == nil {
		return "unknown"
	}
	return Logger.GetLevel().String()
}