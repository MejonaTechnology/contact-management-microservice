package handlers

import (
	"contact-service/internal/services"
	"contact-service/pkg/database"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

// HealthHandler handles health check requests
type HealthHandler struct {
	monitoringService *services.MonitoringService
}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		monitoringService: services.NewMonitoringService(database.DB),
	}
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Check the health status of the contact service with comprehensive monitoring data
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=map[string]interface{}}
// @Success 503 {object} APIResponse{data=map[string]interface{}}
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	uptime := time.Since(startTime)
	version := getEnv("APP_VERSION", "1.0.0")
	environment := getEnv("APP_ENV", "development")
	
	// Get comprehensive health status from monitoring service
	healthStatus := h.monitoringService.GetHealthStatus()
	
	services := make(map[string]interface{})
	overallStatus := healthStatus["status"].(string)
	
	// Check database health
	dbHealth := database.HealthCheck()
	services["database"] = dbHealth
	if dbStatus, ok := dbHealth["status"].(string); ok && dbStatus != "healthy" {
		overallStatus = "unhealthy"
	}
	
	// Check memory usage
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	memoryHealth := map[string]interface{}{
		"status": "healthy",
		"allocated_mb": bToMb(memStats.Alloc),
		"total_allocated_mb": bToMb(memStats.TotalAlloc),
		"system_mb": bToMb(memStats.Sys),
		"gc_count": memStats.NumGC,
	}
	
	// Check if memory usage is concerning
	if memStats.Alloc > 500*1024*1024 { // 500MB threshold
		memoryHealth["status"] = "warning"
		if overallStatus == "healthy" {
			overallStatus = "warning"
		}
	}
	services["memory"] = memoryHealth
	
	// Add error statistics
	errorStats := h.monitoringService.GetErrorStats()
	services["errors"] = map[string]interface{}{
		"status": "healthy",
		"total_errors": errorStats.Total,
		"error_rate": healthStatus["error_rate"],
	}
	
	// Check error rate
	if errorRate, ok := healthStatus["error_rate"].(float64); ok && errorRate > 5 {
		services["errors"].(map[string]interface{})["status"] = "warning"
		if overallStatus == "healthy" {
			overallStatus = "warning"
		}
	}
	
	// Get active alerts
	activeAlerts := h.monitoringService.GetActiveAlerts()
	services["alerts"] = map[string]interface{}{
		"status": "healthy",
		"active_count": len(activeAlerts),
		"alerts": activeAlerts,
	}
	
	// Check for critical alerts
	for _, alert := range activeAlerts {
		if alert.Severity == "critical" {
			services["alerts"].(map[string]interface{})["status"] = "critical"
			overallStatus = "critical"
			break
		} else if alert.Severity == "high" {
			services["alerts"].(map[string]interface{})["status"] = "warning"
			if overallStatus == "healthy" {
				overallStatus = "warning"
			}
		}
	}
	
	// Determine HTTP status code
	statusCode := http.StatusOK
	if overallStatus == "critical" || overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}
	
	response := map[string]interface{}{
		"status": overallStatus,
		"version": version,
		"environment": environment,
		"uptime": uptime.String(),
		"uptime_seconds": uptime.Seconds(),
		"health_score": healthStatus["health_score"],
		"services": services,
		"timestamp": time.Now(),
		"monitoring": map[string]interface{}{
			"total_errors": healthStatus["total_errors"],
			"active_alerts": healthStatus["active_alerts"],
			"last_updated": healthStatus["last_updated"],
		},
	}
	
	c.JSON(statusCode, NewSuccessResponse("Health check completed", response))
}

// ReadinessCheck godoc
// @Summary Readiness check endpoint
// @Description Check if the service is ready to handle requests
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse
// @Success 503 {object} APIResponse
// @Router /ready [get]
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	// Check if database is connected and ready
	if !database.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, NewErrorResponse("Service not ready", "Database not connected"))
		return
	}
	
	c.JSON(http.StatusOK, NewSuccessResponse("Service is ready", map[string]interface{}{
		"ready": true,
		"timestamp": time.Now(),
	}))
}

// LivenessCheck godoc
// @Summary Liveness check endpoint
// @Description Check if the service is alive
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse
// @Router /alive [get]
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, NewSuccessResponse("Service is alive", map[string]interface{}{
		"alive": true,
		"timestamp": time.Now(),
		"uptime": time.Since(startTime).String(),
	}))
}

// MetricsCheck godoc
// @Summary Comprehensive metrics endpoint
// @Description Get detailed service metrics including performance, errors, and system stats
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=map[string]interface{}}
// @Router /metrics [get]
func (h *HealthHandler) MetricsCheck(c *gin.Context) {
	uptime := time.Since(startTime)
	
	// Get database connection stats
	dbStats := database.GetConnectionStats()
	
	// Get comprehensive system metrics from monitoring service
	systemMetrics := h.monitoringService.GetSystemMetrics()
	
	// Get error statistics
	errorStats := h.monitoringService.GetErrorStats()
	
	// Get memory statistics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	// Get Go runtime metrics
	runtimeMetrics := map[string]interface{}{
		"go_version": runtime.Version(),
		"go_routines": runtime.NumGoroutine(),
		"go_max_procs": runtime.GOMAXPROCS(0),
		"memory": map[string]interface{}{
			"allocated_mb": bToMb(memStats.Alloc),
			"total_allocated_mb": bToMb(memStats.TotalAlloc),
			"system_mb": bToMb(memStats.Sys),
			"heap_allocated_mb": bToMb(memStats.HeapAlloc),
			"heap_system_mb": bToMb(memStats.HeapSys),
			"heap_idle_mb": bToMb(memStats.HeapIdle),
			"heap_in_use_mb": bToMb(memStats.HeapInuse),
			"heap_released_mb": bToMb(memStats.HeapReleased),
			"heap_objects": memStats.HeapObjects,
			"gc_count": memStats.NumGC,
			"gc_pause_total_ns": memStats.PauseTotalNs,
		},
	}
	
	metrics := map[string]interface{}{
		"service": map[string]interface{}{
			"uptime_seconds": uptime.Seconds(),
			"version": getEnv("APP_VERSION", "1.0.0"),
			"environment": getEnv("APP_ENV", "development"),
			"start_time": startTime,
		},
		"database": dbStats,
		"runtime": runtimeMetrics,
		"system": systemMetrics,
		"errors": map[string]interface{}{
			"total": errorStats.Total,
			"by_code": errorStats.ByCode,
			"by_severity": errorStats.BySeverity,
			"by_category": errorStats.ByCategory,
			"recent_count": len(errorStats.Recent),
		},
		"performance": systemMetrics["performance_metrics"],
		"business": systemMetrics["business_metrics"],
		"timestamp": time.Now(),
	}
	
	c.JSON(http.StatusOK, NewSuccessResponse("Comprehensive metrics retrieved", metrics))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// DeepHealthCheck godoc
// @Summary Deep health check endpoint
// @Description Perform comprehensive health checks including dependency validation
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=map[string]interface{}}
// @Success 503 {object} APIResponse{data=map[string]interface{}}
// @Router /health/deep [get]
func (h *HealthHandler) DeepHealthCheck(c *gin.Context) {
	startCheck := time.Now()
	
	checks := make(map[string]interface{})
	overallStatus := "healthy"
	
	// Database connectivity check
	dbStart := time.Now()
	dbHealth := database.HealthCheck()
	dbDuration := time.Since(dbStart)
	checks["database"] = map[string]interface{}{
		"status": dbHealth["status"],
		"duration_ms": dbDuration.Milliseconds(),
		"details": dbHealth,
	}
	
	if dbStatus, ok := dbHealth["status"].(string); ok && dbStatus != "healthy" {
		overallStatus = "unhealthy"
	}
	
	// Database query performance test
	queryStart := time.Now()
	queryErr := database.TestQuery()
	queryDuration := time.Since(queryStart)
	queryStatus := "healthy"
	if queryErr != nil {
		queryStatus = "unhealthy"
		overallStatus = "unhealthy"
	}
	checks["database_query"] = map[string]interface{}{
		"status": queryStatus,
		"duration_ms": queryDuration.Milliseconds(),
		"error": queryErr,
	}
	
	// Memory pressure check
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	memoryStatus := "healthy"
	if memStats.Alloc > 1024*1024*1024 { // 1GB threshold
		memoryStatus = "critical"
		overallStatus = "critical"
	} else if memStats.Alloc > 500*1024*1024 { // 500MB threshold  
		memoryStatus = "warning"
		if overallStatus == "healthy" {
			overallStatus = "warning"
		}
	}
	
	checks["memory"] = map[string]interface{}{
		"status": memoryStatus,
		"allocated_mb": bToMb(memStats.Alloc),
		"heap_in_use_mb": bToMb(memStats.HeapInuse),
		"gc_count": memStats.NumGC,
	}
	
	// Error rate check
	errorStats := h.monitoringService.GetErrorStats()
	healthStatus := h.monitoringService.GetHealthStatus()
	errorRate := healthStatus["error_rate"].(float64)
	
	errorStatus := "healthy"
	if errorRate > 10 {
		errorStatus = "critical"
		overallStatus = "critical"
	} else if errorRate > 5 {
		errorStatus = "warning"
		if overallStatus == "healthy" {
			overallStatus = "warning"
		}
	}
	
	checks["errors"] = map[string]interface{}{
		"status": errorStatus,
		"total_errors": errorStats.Total,
		"error_rate": errorRate,
		"recent_errors": len(errorStats.Recent),
	}
	
	// Active alerts check
	activeAlerts := h.monitoringService.GetActiveAlerts()
	alertStatus := "healthy"
	criticalAlerts := 0
	highAlerts := 0
	
	for _, alert := range activeAlerts {
		if alert.Severity == "critical" {
			criticalAlerts++
		} else if alert.Severity == "high" {
			highAlerts++
		}
	}
	
	if criticalAlerts > 0 {
		alertStatus = "critical"
		overallStatus = "critical"
	} else if highAlerts > 0 {
		alertStatus = "warning"
		if overallStatus == "healthy" {
			overallStatus = "warning"
		}
	}
	
	checks["alerts"] = map[string]interface{}{
		"status": alertStatus,
		"total_active": len(activeAlerts),
		"critical_alerts": criticalAlerts,
		"high_alerts": highAlerts,
	}
	
	totalDuration := time.Since(startCheck)
	
	// Determine status code
	statusCode := http.StatusOK
	if overallStatus == "critical" || overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}
	
	response := map[string]interface{}{
		"status": overallStatus,
		"health_score": healthStatus["health_score"],
		"check_duration_ms": totalDuration.Milliseconds(),
		"checks": checks,
		"timestamp": time.Now(),
		"recommendations": generateHealthRecommendations(checks),
	}
	
	c.JSON(statusCode, NewSuccessResponse("Deep health check completed", response))
}

// StatusCheck godoc
// @Summary Quick status endpoint
// @Description Get a quick status overview of the service
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=map[string]interface{}}
// @Router /status [get]
func (h *HealthHandler) StatusCheck(c *gin.Context) {
	healthStatus := h.monitoringService.GetHealthStatus()
	
	response := map[string]interface{}{
		"status": healthStatus["status"],
		"health_score": healthStatus["health_score"],
		"uptime_seconds": time.Since(startTime).Seconds(),
		"version": getEnv("APP_VERSION", "1.0.0"),
		"environment": getEnv("APP_ENV", "development"),
		"active_alerts": healthStatus["active_alerts"],
		"error_rate": healthStatus["error_rate"],
		"timestamp": time.Now(),
	}
	
	c.JSON(http.StatusOK, NewSuccessResponse("Status retrieved", response))
}

// generateHealthRecommendations provides actionable recommendations based on health check results
func generateHealthRecommendations(checks map[string]interface{}) []string {
	recommendations := make([]string, 0)
	
	// Database recommendations
	if dbCheck, ok := checks["database"].(map[string]interface{}); ok {
		if status, ok := dbCheck["status"].(string); ok && status != "healthy" {
			recommendations = append(recommendations, "Check database connectivity and configuration")
		}
		if duration, ok := dbCheck["duration_ms"].(int64); ok && duration > 1000 {
			recommendations = append(recommendations, "Database response time is slow, consider connection pooling optimization")
		}
	}
	
	// Memory recommendations
	if memCheck, ok := checks["memory"].(map[string]interface{}); ok {
		if status, ok := memCheck["status"].(string); ok {
			if status == "critical" {
				recommendations = append(recommendations, "Memory usage is critical, consider increasing memory allocation or optimizing memory usage")
			} else if status == "warning" {
				recommendations = append(recommendations, "Memory usage is elevated, monitor for memory leaks")
			}
		}
	}
	
	// Error recommendations
	if errorCheck, ok := checks["errors"].(map[string]interface{}); ok {
		if status, ok := errorCheck["status"].(string); ok && status != "healthy" {
			recommendations = append(recommendations, "High error rate detected, investigate recent errors and their causes")
		}
	}
	
	// Alert recommendations
	if alertCheck, ok := checks["alerts"].(map[string]interface{}); ok {
		if criticalCount, ok := alertCheck["critical_alerts"].(int); ok && criticalCount > 0 {
			recommendations = append(recommendations, "Critical alerts are active, immediate attention required")
		}
		if highCount, ok := alertCheck["high_alerts"].(int); ok && highCount > 0 {
			recommendations = append(recommendations, "High severity alerts detected, review and address promptly")
		}
	}
	
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "System is operating normally")
	}
	
	return recommendations
}

// bToMb converts bytes to megabytes
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}