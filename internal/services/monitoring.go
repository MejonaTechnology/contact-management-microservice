package services

import (
	"contact-service/internal/models"
	"contact-service/pkg/errors"
	"contact-service/pkg/logger"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
)

// MonitoringService handles system monitoring and error tracking
type MonitoringService struct {
	db          *gorm.DB
	errorStats  *ErrorStatsTracker
	metrics     *MetricsTracker
	alerts      *AlertManager
	mu          sync.RWMutex
}

// ErrorStatsTracker tracks error statistics in memory
type ErrorStatsTracker struct {
	mu            sync.RWMutex
	errorCounts   map[errors.ErrorCode]int
	severityCounts map[errors.ErrorSeverity]int
	categoryCounts map[errors.ErrorCategory]int
	recentErrors  []ErrorEvent
	maxRecent     int
	startTime     time.Time
}

// MetricsTracker tracks performance and business metrics
type MetricsTracker struct {
	mu              sync.RWMutex
	performanceMetrics map[string]*MetricData
	businessMetrics    map[string]*MetricData
}

// AlertManager manages system alerts and notifications
type AlertManager struct {
	mu              sync.RWMutex
	activeAlerts    map[string]*Alert
	alertRules      []AlertRule
	notificationsCh chan Alert
}

// ErrorEvent represents an error occurrence
type ErrorEvent struct {
	Code      errors.ErrorCode     `json:"code"`
	Message   string               `json:"message"`
	Severity  errors.ErrorSeverity `json:"severity"`
	Category  errors.ErrorCategory `json:"category"`
	Endpoint  string               `json:"endpoint"`
	UserID    *uint                `json:"user_id,omitempty"`
	Timestamp time.Time            `json:"timestamp"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// MetricData represents metric information
type MetricData struct {
	Name      string                 `json:"name"`
	Value     float64                `json:"value"`
	Unit      string                 `json:"unit"`
	Tags      map[string]string      `json:"tags"`
	Timestamp time.Time              `json:"timestamp"`
	History   []MetricPoint          `json:"history"`
}

// MetricPoint represents a single metric measurement
type MetricPoint struct {
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

// Alert represents a system alert
type Alert struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Conditions  map[string]interface{} `json:"conditions"`
	Timestamp   time.Time              `json:"timestamp"`
	IsActive    bool                   `json:"is_active"`
	AckedBy     *uint                  `json:"acked_by,omitempty"`
	AckedAt     *time.Time             `json:"acked_at,omitempty"`
}

// AlertRule defines conditions for triggering alerts
type AlertRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // error_rate, performance, business
	Conditions  map[string]interface{} `json:"conditions"`
	Severity    string                 `json:"severity"`
	Enabled     bool                   `json:"enabled"`
	Cooldown    time.Duration          `json:"cooldown"`
	LastTriggered *time.Time           `json:"last_triggered,omitempty"`
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(db *gorm.DB) *MonitoringService {
	service := &MonitoringService{
		db: db,
		errorStats: &ErrorStatsTracker{
			errorCounts:    make(map[errors.ErrorCode]int),
			severityCounts: make(map[errors.ErrorSeverity]int),
			categoryCounts: make(map[errors.ErrorCategory]int),
			recentErrors:   make([]ErrorEvent, 0),
			maxRecent:      100,
			startTime:      time.Now(),
		},
		metrics: &MetricsTracker{
			performanceMetrics: make(map[string]*MetricData),
			businessMetrics:    make(map[string]*MetricData),
		},
		alerts: &AlertManager{
			activeAlerts:    make(map[string]*Alert),
			alertRules:      make([]AlertRule, 0),
			notificationsCh: make(chan Alert, 100),
		},
	}

	// Initialize default alert rules
	service.initializeDefaultAlertRules()

	// Start background processes
	go service.processAlerts()
	go service.cleanupOldMetrics()

	return service
}

// TrackError tracks an error occurrence
func (s *MonitoringService) TrackError(appErr *errors.AppError) {
	s.errorStats.mu.Lock()
	defer s.errorStats.mu.Unlock()

	// Increment counters
	s.errorStats.errorCounts[appErr.Code]++
	s.errorStats.severityCounts[appErr.Severity]++
	s.errorStats.categoryCounts[appErr.Category]++

	// Add to recent errors
	errorEvent := ErrorEvent{
		Code:      appErr.Code,
		Message:   appErr.Message,
		Severity:  appErr.Severity,
		Category:  appErr.Category,
		Endpoint:  appErr.Endpoint,
		UserID:    appErr.UserID,
		Timestamp: appErr.Timestamp,
		Context:   appErr.Context,
	}

	s.errorStats.recentErrors = append(s.errorStats.recentErrors, errorEvent)

	// Keep only recent errors
	if len(s.errorStats.recentErrors) > s.errorStats.maxRecent {
		s.errorStats.recentErrors = s.errorStats.recentErrors[1:]
	}

	// Store in database for persistence
	go s.persistErrorEvent(errorEvent)

	// Check alert conditions
	s.checkErrorAlerts(appErr)

	logger.Debug("Error tracked in monitoring", map[string]interface{}{
		"error_code": appErr.Code,
		"severity":   appErr.Severity,
		"category":   appErr.Category,
	})
}

// TrackMetric tracks a performance or business metric
func (s *MonitoringService) TrackMetric(name string, value float64, unit string, tags map[string]string, metricType string) {
	s.metrics.mu.Lock()
	defer s.metrics.mu.Unlock()

	var metricsMap map[string]*MetricData
	switch metricType {
	case "performance":
		metricsMap = s.metrics.performanceMetrics
	case "business":
		metricsMap = s.metrics.businessMetrics
	default:
		metricsMap = s.metrics.performanceMetrics
	}

	key := name
	if tags != nil {
		// Create unique key including tags
		key = fmt.Sprintf("%s_%v", name, tags)
	}

	metric := metricsMap[key]
	if metric == nil {
		metric = &MetricData{
			Name:      name,
			Unit:      unit,
			Tags:      tags,
			History:   make([]MetricPoint, 0),
		}
		metricsMap[key] = metric
	}

	// Update metric
	metric.Value = value
	metric.Timestamp = time.Now()

	// Add to history
	point := MetricPoint{
		Value:     value,
		Timestamp: metric.Timestamp,
	}
	metric.History = append(metric.History, point)

	// Keep only last 100 points
	if len(metric.History) > 100 {
		metric.History = metric.History[1:]
	}

	// Store performance metric in database
	go s.persistPerformanceMetric(name, value, unit, tags, metricType)

	// Check metric-based alerts
	s.checkMetricAlerts(name, value, tags)

	logger.Debug("Metric tracked", map[string]interface{}{
		"metric": name,
		"value":  value,
		"unit":   unit,
		"type":   metricType,
	})
}

// GetErrorStats returns current error statistics
func (s *MonitoringService) GetErrorStats() *errors.ErrorStats {
	s.errorStats.mu.RLock()
	defer s.errorStats.mu.RUnlock()

	total := 0
	for _, count := range s.errorStats.errorCounts {
		total += count
	}

	// Convert recent errors to AppError slice
	recentAppErrors := make([]errors.AppError, len(s.errorStats.recentErrors))
	for i, event := range s.errorStats.recentErrors {
		recentAppErrors[i] = errors.AppError{
			Code:      event.Code,
			Message:   event.Message,
			Severity:  event.Severity,
			Category:  event.Category,
			Endpoint:  event.Endpoint,
			UserID:    event.UserID,
			Timestamp: event.Timestamp,
			Context:   event.Context,
		}
	}

	return &errors.ErrorStats{
		Total:      total,
		ByCode:     s.errorStats.errorCounts,
		BySeverity: s.errorStats.severityCounts,
		ByCategory: s.errorStats.categoryCounts,
		Recent:     recentAppErrors,
	}
}

// GetSystemMetrics returns current system metrics
func (s *MonitoringService) GetSystemMetrics() map[string]interface{} {
	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()

	uptime := time.Since(s.errorStats.startTime)

	metrics := map[string]interface{}{
		"uptime_seconds":       uptime.Seconds(),
		"error_rate":          s.calculateErrorRate(),
		"performance_metrics": s.metrics.performanceMetrics,
		"business_metrics":    s.metrics.businessMetrics,
		"active_alerts":       len(s.alerts.activeAlerts),
	}

	return metrics
}

// GetActiveAlerts returns currently active alerts
func (s *MonitoringService) GetActiveAlerts() []Alert {
	s.alerts.mu.RLock()
	defer s.alerts.mu.RUnlock()

	alerts := make([]Alert, 0, len(s.alerts.activeAlerts))
	for _, alert := range s.alerts.activeAlerts {
		if alert.IsActive {
			alerts = append(alerts, *alert)
		}
	}

	return alerts
}

// AcknowledgeAlert acknowledges an alert
func (s *MonitoringService) AcknowledgeAlert(alertID string, userID uint) error {
	s.alerts.mu.Lock()
	defer s.alerts.mu.Unlock()

	alert, exists := s.alerts.activeAlerts[alertID]
	if !exists {
		return errors.NewNotFoundError("Alert", alertID)
	}

	now := time.Now()
	alert.AckedBy = &userID
	alert.AckedAt = &now
	alert.IsActive = false

	logger.Info("Alert acknowledged", map[string]interface{}{
		"alert_id": alertID,
		"user_id":  userID,
	})

	return nil
}

// CreateAlert creates a new alert
func (s *MonitoringService) CreateAlert(alertType, severity, title, message string, conditions map[string]interface{}) string {
	alertID := fmt.Sprintf("alert_%d", time.Now().UnixNano())

	alert := &Alert{
		ID:         alertID,
		Type:       alertType,
		Severity:   severity,
		Title:      title,
		Message:    message,
		Conditions: conditions,
		Timestamp:  time.Now(),
		IsActive:   true,
	}

	s.alerts.mu.Lock()
	s.alerts.activeAlerts[alertID] = alert
	s.alerts.mu.Unlock()

	// Send to notification channel
	select {
	case s.alerts.notificationsCh <- *alert:
	default:
		logger.Warn("Alert notification channel full", map[string]interface{}{
			"alert_id": alertID,
		})
	}

	// Persist alert to database
	go s.persistAlert(alert)

	logger.Warn("Alert created", map[string]interface{}{
		"alert_id": alertID,
		"type":     alertType,
		"severity": severity,
		"title":    title,
	})

	return alertID
}

// GetHealthStatus returns overall system health status
func (s *MonitoringService) GetHealthStatus() map[string]interface{} {
	errorStats := s.GetErrorStats()
	systemMetrics := s.GetSystemMetrics()
	activeAlerts := s.GetActiveAlerts()

	// Calculate health score (0-100)
	healthScore := s.calculateHealthScore(errorStats, activeAlerts)

	status := "healthy"
	if healthScore < 50 {
		status = "critical"
	} else if healthScore < 70 {
		status = "degraded"
	} else if healthScore < 90 {
		status = "warning"
	}

	return map[string]interface{}{
		"status":         status,
		"health_score":   healthScore,
		"uptime":         systemMetrics["uptime_seconds"],
		"error_rate":     systemMetrics["error_rate"],
		"active_alerts":  len(activeAlerts),
		"total_errors":   errorStats.Total,
		"last_updated":   time.Now(),
	}
}

// Private methods

func (s *MonitoringService) initializeDefaultAlertRules() {
	rules := []AlertRule{
		{
			ID:   "high_error_rate",
			Name: "High Error Rate",
			Type: "error_rate",
			Conditions: map[string]interface{}{
				"threshold":     10, // errors per minute
				"window":        "1m",
				"min_requests":  10,
			},
			Severity: "high",
			Enabled:  true,
			Cooldown: 5 * time.Minute,
		},
		{
			ID:   "critical_errors",
			Name: "Critical Errors Detected",
			Type: "error_rate",
			Conditions: map[string]interface{}{
				"severity":   "critical",
				"threshold":  1, // any critical error
				"window":     "1m",
			},
			Severity: "critical",
			Enabled:  true,
			Cooldown: 1 * time.Minute,
		},
		{
			ID:   "high_response_time",
			Name: "High Response Time",
			Type: "performance",
			Conditions: map[string]interface{}{
				"metric":    "http_request_duration",
				"threshold": 5000, // 5 seconds
				"window":    "5m",
			},
			Severity: "medium",
			Enabled:  true,
			Cooldown: 10 * time.Minute,
		},
	}

	s.alerts.alertRules = rules
}

func (s *MonitoringService) checkErrorAlerts(appErr *errors.AppError) {
	for _, rule := range s.alerts.alertRules {
		if !rule.Enabled || rule.Type != "error_rate" {
			continue
		}

		// Check cooldown
		if rule.LastTriggered != nil && time.Since(*rule.LastTriggered) < rule.Cooldown {
			continue
		}

		shouldTrigger := s.evaluateErrorAlertRule(rule, appErr)
		if shouldTrigger {
			s.triggerAlert(rule, map[string]interface{}{
				"error_code": appErr.Code,
				"severity":   appErr.Severity,
				"endpoint":   appErr.Endpoint,
			})
		}
	}
}

func (s *MonitoringService) checkMetricAlerts(metricName string, value float64, tags map[string]string) {
	for _, rule := range s.alerts.alertRules {
		if !rule.Enabled || rule.Type != "performance" {
			continue
		}

		// Check cooldown
		if rule.LastTriggered != nil && time.Since(*rule.LastTriggered) < rule.Cooldown {
			continue
		}

		shouldTrigger := s.evaluateMetricAlertRule(rule, metricName, value, tags)
		if shouldTrigger {
			s.triggerAlert(rule, map[string]interface{}{
				"metric": metricName,
				"value":  value,
				"tags":   tags,
			})
		}
	}
}

func (s *MonitoringService) evaluateErrorAlertRule(rule AlertRule, appErr *errors.AppError) bool {
	// Simplified rule evaluation - in production, this would be more sophisticated
	if severityCondition, exists := rule.Conditions["severity"]; exists {
		if string(appErr.Severity) == severityCondition {
			return true
		}
	}

	if threshold, exists := rule.Conditions["threshold"]; exists {
		if thresholdFloat, ok := threshold.(float64); ok {
			// Check error rate in time window
			errorRate := s.calculateErrorRate()
			if errorRate > thresholdFloat {
				return true
			}
		}
	}

	return false
}

func (s *MonitoringService) evaluateMetricAlertRule(rule AlertRule, metricName string, value float64, tags map[string]string) bool {
	if ruleMetric, exists := rule.Conditions["metric"]; exists {
		if ruleMetric != metricName {
			return false
		}
	}

	if threshold, exists := rule.Conditions["threshold"]; exists {
		if thresholdFloat, ok := threshold.(float64); ok {
			return value > thresholdFloat
		}
	}

	return false
}

func (s *MonitoringService) triggerAlert(rule AlertRule, context map[string]interface{}) {
	alertID := s.CreateAlert(
		rule.Type,
		rule.Severity,
		rule.Name,
		fmt.Sprintf("Alert triggered: %s", rule.Name),
		context,
	)

	// Update rule last triggered time
	now := time.Now()
	for i := range s.alerts.alertRules {
		if s.alerts.alertRules[i].ID == rule.ID {
			s.alerts.alertRules[i].LastTriggered = &now
			break
		}
	}

	logger.Warn("Alert triggered", map[string]interface{}{
		"alert_id":  alertID,
		"rule_id":   rule.ID,
		"rule_name": rule.Name,
		"context":   context,
	})
}

func (s *MonitoringService) calculateErrorRate() float64 {
	s.errorStats.mu.RLock()
	defer s.errorStats.mu.RUnlock()

	if len(s.errorStats.recentErrors) == 0 {
		return 0
	}

	// Calculate errors in the last minute
	oneMinuteAgo := time.Now().Add(-1 * time.Minute)
	recentErrorCount := 0

	for _, err := range s.errorStats.recentErrors {
		if err.Timestamp.After(oneMinuteAgo) {
			recentErrorCount++
		}
	}

	return float64(recentErrorCount)
}

func (s *MonitoringService) calculateHealthScore(errorStats *errors.ErrorStats, activeAlerts []Alert) int {
	score := 100

	// Deduct points for errors
	if errorStats.Total > 0 {
		score -= min(50, errorStats.Total/10) // Max 50 points deduction for errors
	}

	// Deduct points for active alerts
	for _, alert := range activeAlerts {
		switch alert.Severity {
		case "critical":
			score -= 30
		case "high":
			score -= 20
		case "medium":
			score -= 10
		case "low":
			score -= 5
		}
	}

	// Ensure score doesn't go below 0
	if score < 0 {
		score = 0
	}

	return score
}

func (s *MonitoringService) processAlerts() {
	for alert := range s.alerts.notificationsCh {
		// Process alert notifications
		// This could send emails, Slack messages, etc.
		logger.Info("Processing alert notification", map[string]interface{}{
			"alert_id": alert.ID,
			"type":     alert.Type,
			"severity": alert.Severity,
		})

		// Here you would integrate with notification services
		// For now, just log the alert
	}
}

func (s *MonitoringService) cleanupOldMetrics() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.metrics.mu.Lock()
		
		// Clean up old metric history points (older than 24 hours)
		cutoff := time.Now().Add(-24 * time.Hour)
		
		for _, metric := range s.metrics.performanceMetrics {
			newHistory := make([]MetricPoint, 0)
			for _, point := range metric.History {
				if point.Timestamp.After(cutoff) {
					newHistory = append(newHistory, point)
				}
			}
			metric.History = newHistory
		}
		
		for _, metric := range s.metrics.businessMetrics {
			newHistory := make([]MetricPoint, 0)
			for _, point := range metric.History {
				if point.Timestamp.After(cutoff) {
					newHistory = append(newHistory, point)
				}
			}
			metric.History = newHistory
		}
		
		s.metrics.mu.Unlock()
		
		logger.Debug("Cleaned up old metrics")
	}
}

func (s *MonitoringService) persistErrorEvent(event ErrorEvent) {
	// Store activity log entry
	activityLog := models.ActivityLog{
		Action:      "error_occurred",
		EntityType:  "system",
		Description: fmt.Sprintf("Error occurred: %s - %s", event.Code, event.Message),
		Metadata: map[string]interface{}{
			"error_code": event.Code,
			"severity":   event.Severity,
			"category":   event.Category,
			"endpoint":   event.Endpoint,
			"context":    event.Context,
		},
		UserID:    event.UserID,
		CreatedAt: event.Timestamp,
	}

	if err := s.db.Create(&activityLog).Error; err != nil {
		logger.Error("Failed to persist error event", err, map[string]interface{}{
			"error_code": event.Code,
		})
	}
}

func (s *MonitoringService) persistPerformanceMetric(name string, value float64, unit string, tags map[string]string, metricType string) {
	// Convert tags to JSONArray format
	var tagsArray models.JSONArray
	for k, v := range tags {
		tagsArray = append(tagsArray, map[string]string{k: v})
	}

	performanceMetric := models.PerformanceMetric{
		MetricName: name,
		MetricType: metricType,
		Value:      value,
		Unit:       unit,
		Tags:       tagsArray,
		RecordedAt: time.Now(),
	}

	if err := s.db.Create(&performanceMetric).Error; err != nil {
		logger.Error("Failed to persist performance metric", err, map[string]interface{}{
			"metric": name,
			"value":  value,
		})
	}
}

func (s *MonitoringService) persistAlert(alert *Alert) {
	systemAlert := models.SystemAlert{
		Type:     alert.Type,
		Priority: alert.Severity,
		Title:    alert.Title,
		Message:  alert.Message,
		IsActive: alert.IsActive,
		CreatedAt: alert.Timestamp,
	}

	if err := s.db.Create(&systemAlert).Error; err != nil {
		logger.Error("Failed to persist alert", err, map[string]interface{}{
			"alert_id": alert.ID,
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}