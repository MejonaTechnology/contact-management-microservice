package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"contact-service/internal/services"
)

// MockMonitoringService is a mock implementation of MonitoringService
type MockMonitoringService struct {
	mock.Mock
}

func (m *MockMonitoringService) GetHealthStatus() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func (m *MockMonitoringService) TrackError(operation string, err error) {
	m.Called(operation, err)
}

func (m *MockMonitoringService) TrackMetric(name string, value float64, tags map[string]string) {
	m.Called(name, value, tags)
}

func (m *MockMonitoringService) GetMetrics() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func (m *MockMonitoringService) GetAlerts() []services.Alert {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]services.Alert)
}

func (m *MockMonitoringService) CreateAlert(alert services.Alert) error {
	args := m.Called(alert)
	return args.Error(0)
}

// HealthHandlerTestSuite defines the test suite for HealthHandler
type HealthHandlerTestSuite struct {
	suite.Suite
	handler           *HealthHandler
	mockMonitoring    *MockMonitoringService
	router           *gin.Engine
	startTime        time.Time
}

// SetupSuite runs before all tests in the suite
func (suite *HealthHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.startTime = time.Now()
}

// SetupTest runs before each test
func (suite *HealthHandlerTestSuite) SetupTest() {
	suite.mockMonitoring = new(MockMonitoringService)
	suite.handler = &HealthHandler{
		monitoringService: suite.mockMonitoring,
		startTime:        suite.startTime,
	}
	suite.router = gin.New()
	
	// Setup routes
	suite.router.GET("/health", suite.handler.HealthCheck)
	suite.router.GET("/health/deep", suite.handler.DeepHealthCheck)
	suite.router.GET("/status", suite.handler.StatusCheck)
	suite.router.GET("/ready", suite.handler.ReadinessCheck)
	suite.router.GET("/alive", suite.handler.LivenessCheck)
	suite.router.GET("/metrics", suite.handler.MetricsCheck)
}

// TearDownTest runs after each test
func (suite *HealthHandlerTestSuite) TearDownTest() {
	suite.mockMonitoring.AssertExpectations(suite.T())
}

// Test HealthCheck endpoint
func (suite *HealthHandlerTestSuite) TestHealthCheck_Healthy() {
	// Arrange
	healthStatus := map[string]interface{}{
		"health_score":   95.0,
		"database":       "healthy",
		"memory_usage":   45.2,
		"total_errors":   0,
		"active_alerts":  0,
		"recommendations": []string{},
	}

	suite.mockMonitoring.On("GetHealthStatus").Return(healthStatus)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Act
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "Health check completed", response.Message)
	assert.NotNil(suite.T(), response.Data)

	// Check health data structure
	healthData := response.Data.(map[string]interface{})
	assert.Equal(suite.T(), "healthy", healthData["status"])
	assert.NotEmpty(suite.T(), healthData["version"])
	assert.NotEmpty(suite.T(), healthData["uptime"])
	assert.NotEmpty(suite.T(), healthData["environment"])
	assert.NotNil(suite.T(), healthData["services"])
}

func (suite *HealthHandlerTestSuite) TestHealthCheck_Unhealthy() {
	// Arrange
	healthStatus := map[string]interface{}{
		"health_score":  40.0, // Low health score
		"database":      "unhealthy",
		"memory_usage":  85.0, // High memory usage
		"total_errors":  10,
		"active_alerts": 3,
		"recommendations": []string{
			"High memory usage detected",
			"Database connection issues",
		},
	}

	suite.mockMonitoring.On("GetHealthStatus").Return(healthStatus)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Act
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusServiceUnavailable, w.Code)
	
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
	
	healthData := response.Data.(map[string]interface{})
	assert.Equal(suite.T(), "unhealthy", healthData["status"])
	assert.Equal(suite.T(), 40.0, healthData["health_score"])
}

// Test DeepHealthCheck endpoint
func (suite *HealthHandlerTestSuite) TestDeepHealthCheck_Success() {
	// Arrange
	healthStatus := map[string]interface{}{
		"health_score":    90.0,
		"database":        "healthy",
		"memory_usage":    50.0,
		"cpu_usage":       30.0,
		"disk_usage":      60.0,
		"total_errors":    2,
		"active_alerts":   1,
		"query_time":      25.5,
		"connection_pool": map[string]interface{}{
			"open_connections": 5,
			"idle_connections": 3,
			"in_use_connections": 2,
		},
		"recommendations": []string{
			"System performance is good",
		},
		"detailed_analysis": map[string]interface{}{
			"memory_pressure": "low",
			"database_performance": "good",
			"error_trends": "stable",
		},
	}

	suite.mockMonitoring.On("GetHealthStatus").Return(healthStatus)

	req := httptest.NewRequest("GET", "/health/deep", nil)
	w := httptest.NewRecorder()

	// Act
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "Deep health analysis completed", response.Message)
	assert.NotNil(suite.T(), response.Data)

	healthData := response.Data.(map[string]interface{})
	assert.Contains(suite.T(), healthData, "analysis")
	assert.Contains(suite.T(), healthData, "recommendations")
	assert.Contains(suite.T(), healthData, "system_metrics")
}

func (suite *HealthHandlerTestSuite) TestDeepHealthCheck_WithIssues() {
	// Arrange
	healthStatus := map[string]interface{}{
		"health_score":    35.0, // Very low health score
		"database":        "degraded",
		"memory_usage":    90.0, // Critical memory usage
		"cpu_usage":       85.0, // High CPU usage
		"total_errors":    25,
		"active_alerts":   5,
		"query_time":      150.0, // Slow queries
		"recommendations": []string{
			"CRITICAL: Memory usage above 90%",
			"WARNING: High CPU usage detected",
			"ERROR: Database queries are slow",
			"ALERT: Multiple active alerts require attention",
		},
		"detailed_analysis": map[string]interface{}{
			"memory_pressure": "critical",
			"database_performance": "poor",
			"error_trends": "increasing",
		},
	}

	suite.mockMonitoring.On("GetHealthStatus").Return(healthStatus)

	req := httptest.NewRequest("GET", "/health/deep", nil)
	w := httptest.NewRecorder()

	// Act
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusServiceUnavailable, w.Code)
	
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
	assert.Contains(suite.T(), response.Message, "issues detected")
}

// Test StatusCheck endpoint
func (suite *HealthHandlerTestSuite) TestStatusCheck_Success() {
	// Arrange
	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	// Act
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "Service is running", response.Message)
	assert.NotNil(suite.T(), response.Data)

	statusData := response.Data.(map[string]interface{})
	assert.Equal(suite.T(), "running", statusData["status"])
	assert.NotEmpty(suite.T(), statusData["uptime"])
	assert.NotEmpty(suite.T(), statusData["version"])
}

// Test ReadinessCheck endpoint
func (suite *HealthHandlerTestSuite) TestReadinessCheck_Ready() {
	// Arrange
	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	// Act
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "Service is ready", response.Message)
}

// Test LivenessCheck endpoint
func (suite *HealthHandlerTestSuite) TestLivenessCheck_Alive() {
	// Arrange
	req := httptest.NewRequest("GET", "/alive", nil)
	w := httptest.NewRecorder()

	// Act
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "Service is alive", response.Message)
}

// Test MetricsCheck endpoint
func (suite *HealthHandlerTestSuite) TestMetricsCheck_Success() {
	// Arrange
	metrics := map[string]interface{}{
		"system": map[string]interface{}{
			"memory_usage":    45.2,
			"cpu_usage":       30.5,
			"disk_usage":      65.0,
			"goroutines":      100,
			"gc_stats":        map[string]interface{}{
				"num_gc":        50,
				"pause_total":   "15ms",
			},
		},
		"database": map[string]interface{}{
			"open_connections":    5,
			"idle_connections":    3,
			"in_use_connections":  2,
			"max_open_conns":      10,
			"max_idle_conns":      5,
		},
		"application": map[string]interface{}{
			"total_requests":      1000,
			"total_errors":        5,
			"avg_response_time":   150.5,
			"active_connections":  25,
		},
		"business": map[string]interface{}{
			"total_contacts":      500,
			"new_contacts_today":  10,
			"active_assignments":  25,
		},
	}

	suite.mockMonitoring.On("GetMetrics").Return(metrics)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	// Act
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "System metrics retrieved", response.Message)
	assert.NotNil(suite.T(), response.Data)

	metricsData := response.Data.(map[string]interface{})
	assert.Contains(suite.T(), metricsData, "system")
	assert.Contains(suite.T(), metricsData, "database")
	assert.Contains(suite.T(), metricsData, "application")
	assert.Contains(suite.T(), metricsData, "business")
}

// Test health score calculation
func (suite *HealthHandlerTestSuite) TestCalculateHealthScore() {
	// Test healthy system
	healthyStatus := map[string]interface{}{
		"database":      "healthy",
		"memory_usage":  40.0,
		"total_errors":  1,
		"active_alerts": 0,
		"query_time":    20.0,
	}
	score := suite.handler.calculateHealthScore(healthyStatus)
	assert.True(suite.T(), score >= 90.0)

	// Test unhealthy system
	unhealthyStatus := map[string]interface{}{
		"database":      "unhealthy",
		"memory_usage":  95.0,
		"total_errors":  50,
		"active_alerts": 10,
		"query_time":    500.0,
	}
	score = suite.handler.calculateHealthScore(unhealthyStatus)
	assert.True(suite.T(), score <= 30.0)

	// Test moderate issues
	moderateStatus := map[string]interface{}{
		"database":      "degraded",
		"memory_usage":  70.0,
		"total_errors":  10,
		"active_alerts": 2,
		"query_time":    100.0,
	}
	score = suite.handler.calculateHealthScore(moderateStatus)
	assert.True(suite.T(), score >= 50.0 && score <= 75.0)
}

// Test recommendations generation
func (suite *HealthHandlerTestSuite) TestGenerateRecommendations() {
	// Test high memory usage
	status := map[string]interface{}{
		"memory_usage":  85.0,
		"cpu_usage":     30.0,
		"total_errors":  5,
		"active_alerts": 2,
		"query_time":    50.0,
	}
	recommendations := suite.handler.generateRecommendations(status)
	assert.Contains(suite.T(), recommendations, "High memory usage detected")

	// Test multiple issues
	problematicStatus := map[string]interface{}{
		"memory_usage":  90.0,
		"cpu_usage":     85.0,
		"total_errors":  20,
		"active_alerts": 5,
		"query_time":    200.0,
	}
	recommendations = suite.handler.generateRecommendations(problematicStatus)
	assert.True(suite.T(), len(recommendations) >= 4) // Should have multiple recommendations
}

// Test uptime calculation
func (suite *HealthHandlerTestSuite) TestUptimeCalculation() {
	// Test that uptime is calculated correctly
	uptime := time.Since(suite.handler.startTime)
	assert.True(suite.T(), uptime > 0)
	
	// Test uptime string formatting
	uptimeStr := suite.handler.formatUptime(uptime)
	assert.NotEmpty(suite.T(), uptimeStr)
	assert.Contains(suite.T(), uptimeStr, "s") // Should contain seconds
}

// Performance test for health check
func (suite *HealthHandlerTestSuite) TestHealthCheck_Performance() {
	// Arrange
	healthStatus := map[string]interface{}{
		"health_score":  95.0,
		"database":      "healthy",
		"total_errors":  0,
		"active_alerts": 0,
	}
	suite.mockMonitoring.On("GetHealthStatus").Return(healthStatus)

	// Measure performance
	start := time.Now()
	
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	duration := time.Since(start)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Less(suite.T(), duration, 100*time.Millisecond) // Should respond quickly
}

// Run the test suite
func TestHealthHandlerSuite(t *testing.T) {
	suite.Run(t, new(HealthHandlerTestSuite))
}

// Individual test functions
func TestHealthCheck(t *testing.T) {
	suite.Run(t, new(HealthHandlerTestSuite))
}

func TestDeepHealthCheck(t *testing.T) {
	suite.Run(t, new(HealthHandlerTestSuite))
}

func TestStatusCheck(t *testing.T) {
	suite.Run(t, new(HealthHandlerTestSuite))
}