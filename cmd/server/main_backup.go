package main

import (
	"contact-service/internal/handlers"
	"contact-service/internal/middleware"
	"contact-service/pkg/database"
	"contact-service/pkg/logger"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

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
	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler()
	dashboardHandler := handlers.NewDashboardContactHandler()

	// Health check endpoints
	router.GET("/health", healthHandler.HealthCheck)
	router.GET("/health/deep", healthHandler.DeepHealthCheck)
	router.GET("/status", healthHandler.StatusCheck)
	router.GET("/ready", healthHandler.ReadinessCheck)
	router.GET("/alive", healthHandler.LivenessCheck)
	router.GET("/metrics", healthHandler.MetricsCheck)

	// Working dashboard endpoints (direct implementation)
	router.GET("/api/v1/dashboard/contacts", dashboardHandler.GetContactSubmissions)
	router.GET("/api/v1/dashboard/contacts/stats", dashboardHandler.GetContactSubmissionStats)
	router.POST("/api/v1/dashboard/contact", dashboardHandler.CreateContactSubmission)
	router.PUT("/api/v1/dashboard/contacts/:id/status", dashboardHandler.UpdateContactSubmissionStatus)
	router.GET("/api/v1/dashboard/contacts/:id", dashboardHandler.GetContactSubmission)
	
	// Export endpoint
	router.GET("/api/v1/dashboard/contacts/export", dashboardHandler.ExportContactSubmissions)
	router.POST("/api/v1/dashboard/contacts/bulk-update", dashboardHandler.BulkUpdateContactSubmissions)

	// API routes
	api := router.Group("/api/v1")
	{
		// Authentication routes (no auth required)
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", middleware.AuthMiddleware(), authHandler.Logout)
			auth.GET("/profile", middleware.AuthMiddleware(), authHandler.GetProfile)
			auth.POST("/change-password", middleware.AuthMiddleware(), authHandler.ChangePassword)
			auth.GET("/validate", middleware.AuthMiddleware(), authHandler.ValidateToken)
		}

		// Public contact submission endpoints (no auth required)
		public := api.Group("/public")
		{
			public.POST("/contact", dashboardHandler.CreateContactSubmission)
		}

		// Simple test endpoints
		api.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Contact service test endpoint working", "status": "success"})
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
	log.Printf("Dashboard endpoints available:")
	log.Printf("  GET  /api/v1/dashboard/contacts - List contacts")
	log.Printf("  GET  /api/v1/dashboard/contacts/stats - Contact statistics")
	log.Printf("  POST /api/v1/dashboard/contact - Create contact")
	log.Printf("  PUT  /api/v1/dashboard/contacts/:id/status - Update contact status")
	log.Printf("  GET  /api/v1/dashboard/contacts/:id - Get contact details")
	log.Printf("  GET  /api/v1/dashboard/contacts/export - Export contacts")
	log.Printf("  POST /api/v1/dashboard/contacts/bulk-update - Bulk update contacts")
	
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}