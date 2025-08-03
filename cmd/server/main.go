package main

import (
	"contact-service/internal/handlers"
	"contact-service/internal/middleware"
	"contact-service/pkg/database"
	"contact-service/pkg/logger"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var startTime = time.Now()

// @title Contact Management Microservice API
// @version 1.0
// @description Professional contact management microservice for Mejona Technology
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.mejona.com/support
// @contact.email support@mejona.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8081
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize logger
	logger.InitLogger()

	// Initialize database
	if err := database.InitDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Initialize Gin router
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	
	// Add essential middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	// Initialize handlers
	authHandler := handlers.NewAuthHandler()
	dashboardHandler := handlers.NewDashboardContactHandler()

	// ===== HEALTH CHECK ENDPOINTS =====
	router.GET("/health", simpleHealthCheck)
	router.GET("/health/deep", deepHealthCheck)
	router.GET("/status", statusCheck)
	router.GET("/ready", readinessCheck)
	router.GET("/alive", livenessCheck)
	router.GET("/metrics", metricsCheck)

	// ===== DASHBOARD ENDPOINTS =====
	router.GET("/api/v1/dashboard/contacts", dashboardHandler.GetContactSubmissions)
	router.GET("/api/v1/dashboard/contacts/stats", dashboardHandler.GetContactSubmissionStats)
	router.POST("/api/v1/dashboard/contact", dashboardHandler.CreateContactSubmission)
	router.PUT("/api/v1/dashboard/contacts/:id/status", dashboardHandler.UpdateContactSubmissionStatus)
	router.GET("/api/v1/dashboard/contacts/:id", dashboardHandler.GetContactSubmission)
	router.GET("/api/v1/dashboard/contacts/export", dashboardHandler.ExportContactSubmissions)
	router.POST("/api/v1/dashboard/contacts/bulk-update", dashboardHandler.BulkUpdateContactSubmissions)

	// ===== API ROUTES =====
	api := router.Group("/api/v1")
	{
		// Authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", middleware.AuthMiddleware(), authHandler.Logout)
			auth.GET("/profile", middleware.AuthMiddleware(), authHandler.GetProfile)
			auth.POST("/change-password", middleware.AuthMiddleware(), authHandler.ChangePassword)
			auth.GET("/validate", middleware.AuthMiddleware(), authHandler.ValidateToken)
		}

		// Public contact submission endpoints
		public := api.Group("/public")
		{
			public.POST("/contact", dashboardHandler.CreateContactSubmission)
		}

		// Test endpoint
		api.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"success": true,
				"message": "Contact service test endpoint working",
				"data": map[string]interface{}{
					"service": "Contact Management Microservice",
					"version": "1.0.0",
					"status": "operational",
					"timestamp": time.Now(),
				},
			})
		})
	}

	// Swagger documentation
	if os.Getenv("ENABLE_SWAGGER") == "true" {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Contact Service starting on port %s", port)
	log.Printf("All endpoints available:")
	log.Printf("  HEALTH ENDPOINTS:")
	log.Printf("    GET  /health - Basic health check")
	log.Printf("    GET  /health/deep - Deep health check")
	log.Printf("    GET  /status - Quick status")
	log.Printf("    GET  /ready - Readiness check")
	log.Printf("    GET  /alive - Liveness check")
	log.Printf("    GET  /metrics - System metrics")
	log.Printf("  DASHBOARD ENDPOINTS:")
	log.Printf("    GET  /api/v1/dashboard/contacts - List contacts")
	log.Printf("    GET  /api/v1/dashboard/contacts/stats - Contact statistics")
	log.Printf("    GET  /api/v1/dashboard/contacts/:id - Get contact by ID")
	log.Printf("    POST /api/v1/dashboard/contact - Create contact")
	log.Printf("    PUT  /api/v1/dashboard/contacts/:id/status - Update contact status")
	log.Printf("    GET  /api/v1/dashboard/contacts/export - Export contacts")
	log.Printf("    POST /api/v1/dashboard/contacts/bulk-update - Bulk update")
	log.Printf("  AUTH ENDPOINTS:")
	log.Printf("    POST /api/v1/auth/login - Login")
	log.Printf("    POST /api/v1/auth/refresh - Refresh token")
	log.Printf("    POST /api/v1/auth/logout - Logout")
	log.Printf("    GET  /api/v1/auth/profile - Get profile")
	log.Printf("    POST /api/v1/auth/change-password - Change password")
	log.Printf("    GET  /api/v1/auth/validate - Validate token")
	log.Printf("  OTHER ENDPOINTS:")
	log.Printf("    POST /api/v1/public/contact - Public contact submission")
	log.Printf("    GET  /api/v1/test - Test endpoint")
	
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// Simple health check without dependencies
func simpleHealthCheck(c *gin.Context) {
	uptime := time.Since(startTime)
	
	// Check database
	dbHealth := "healthy"
	if db := database.GetDB(); db != nil {
		if sqlDB, err := db.DB(); err == nil {
			if err := sqlDB.Ping(); err != nil {
				dbHealth = "unhealthy"
			}
		} else {
			dbHealth = "unhealthy"
		}
	} else {
		dbHealth = "unhealthy"
	}

	status := "healthy"
	if dbHealth != "healthy" {
		status = "unhealthy"
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Health check completed",
		"data": map[string]interface{}{
			"status":    status,
			"service":   "Contact Management Microservice",
			"version":   getEnv("APP_VERSION", "1.0.0"),
			"uptime":    uptime.String(),
			"database":  dbHealth,
			"timestamp": time.Now(),
		},
	}

	statusCode := http.StatusOK
	if status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

// Deep health check
func deepHealthCheck(c *gin.Context) {
	startCheck := time.Now()
	checks := make(map[string]interface{})
	overallStatus := "healthy"

	// Database check
	dbStart := time.Now()
	dbHealth := "healthy"
	var dbError error
	if db := database.GetDB(); db != nil {
		if sqlDB, err := db.DB(); err == nil {
			if err := sqlDB.Ping(); err != nil {
				dbHealth = "unhealthy"
				dbError = err
				overallStatus = "unhealthy"
			}
		} else {
			dbHealth = "unhealthy"
			dbError = err
			overallStatus = "unhealthy"
		}
	} else {
		dbHealth = "unhealthy"
		overallStatus = "unhealthy"
	}
	dbDuration := time.Since(dbStart)

	checks["database"] = map[string]interface{}{
		"status":      dbHealth,
		"duration_ms": dbDuration.Milliseconds(),
		"error":       dbError,
	}

	// Memory check
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	memoryStatus := "healthy"
	allocatedMB := memStats.Alloc / 1024 / 1024

	if allocatedMB > 1024 { // 1GB
		memoryStatus = "critical"
		overallStatus = "critical"
	} else if allocatedMB > 512 { // 512MB
		memoryStatus = "warning"
		if overallStatus == "healthy" {
			overallStatus = "warning"
		}
	}

	checks["memory"] = map[string]interface{}{
		"status":         memoryStatus,
		"allocated_mb":   allocatedMB,
		"heap_in_use_mb": memStats.HeapInuse / 1024 / 1024,
		"gc_count":       memStats.NumGC,
	}

	totalDuration := time.Since(startCheck)
	statusCode := http.StatusOK
	if overallStatus == "critical" || overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Deep health check completed",
		"data": map[string]interface{}{
			"status":             overallStatus,
			"check_duration_ms":  totalDuration.Milliseconds(),
			"checks":             checks,
			"timestamp":          time.Now(),
		},
	}

	c.JSON(statusCode, response)
}

// Status check
func statusCheck(c *gin.Context) {
	response := map[string]interface{}{
		"success": true,
		"message": "Status retrieved",
		"data": map[string]interface{}{
			"status":         "healthy",
			"uptime_seconds": time.Since(startTime).Seconds(),
			"version":        getEnv("APP_VERSION", "1.0.0"),
			"environment":    getEnv("APP_ENV", "production"),
			"timestamp":      time.Now(),
		},
	}

	c.JSON(http.StatusOK, response)
}

// Readiness check
func readinessCheck(c *gin.Context) {
	// Check if database is connected
	ready := true
	if db := database.GetDB(); db != nil {
		if sqlDB, err := db.DB(); err == nil {
			if err := sqlDB.Ping(); err != nil {
				ready = false
			}
		} else {
			ready = false
		}
	} else {
		ready = false
	}

	statusCode := http.StatusOK
	message := "Service is ready"
	if !ready {
		statusCode = http.StatusServiceUnavailable
		message = "Service not ready"
	}

	response := map[string]interface{}{
		"success": ready,
		"message": message,
		"data": map[string]interface{}{
			"ready":     ready,
			"timestamp": time.Now(),
		},
	}

	c.JSON(statusCode, response)
}

// Liveness check
func livenessCheck(c *gin.Context) {
	response := map[string]interface{}{
		"success": true,
		"message": "Service is alive",
		"data": map[string]interface{}{
			"alive":     true,
			"timestamp": time.Now(),
			"uptime":    time.Since(startTime).String(),
		},
	}

	c.JSON(http.StatusOK, response)
}

// Metrics check
func metricsCheck(c *gin.Context) {
	uptime := time.Since(startTime)
	
	// Get memory statistics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	metrics := map[string]interface{}{
		"service": map[string]interface{}{
			"uptime_seconds": uptime.Seconds(),
			"version":        getEnv("APP_VERSION", "1.0.0"),
			"environment":    getEnv("APP_ENV", "production"),
			"start_time":     startTime,
		},
		"runtime": map[string]interface{}{
			"go_version":   runtime.Version(),
			"go_routines":  runtime.NumGoroutine(),
			"go_max_procs": runtime.GOMAXPROCS(0),
			"memory": map[string]interface{}{
				"allocated_mb":      memStats.Alloc / 1024 / 1024,
				"total_allocated_mb": memStats.TotalAlloc / 1024 / 1024,
				"system_mb":         memStats.Sys / 1024 / 1024,
				"heap_allocated_mb": memStats.HeapAlloc / 1024 / 1024,
				"heap_in_use_mb":    memStats.HeapInuse / 1024 / 1024,
				"gc_count":          memStats.NumGC,
			},
		},
		"timestamp": time.Now(),
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Comprehensive metrics retrieved",
		"data":    metrics,
	}

	c.JSON(http.StatusOK, response)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}