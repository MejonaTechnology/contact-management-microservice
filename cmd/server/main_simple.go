package main

import (
	"contact-service/pkg/database"
	"contact-service/pkg/logger"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

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

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(cors.Default())

	// Health check endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "OK",
			"service": "Contact Service",
			"version": "1.0.0",
		})
	})

	// Dashboard endpoints (working implementation)
	router.GET("/api/v1/dashboard/contacts", func(c *gin.Context) {
		db := database.GetDB()
		var contacts []map[string]interface{}

		// Query contact_submissions table
		rows, err := db.Raw("SELECT id, name, email, phone, subject, message, status, created_at FROM contact_submissions ORDER BY created_at DESC LIMIT 10").Rows()
		if err != nil {
			c.JSON(500, gin.H{"success": false, "error": "Database error: " + err.Error()})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var id int
			var name, email, status string
			var phone, subject, message, created_at interface{}

			rows.Scan(&id, &name, &email, &phone, &subject, &message, &status, &created_at)
			contacts = append(contacts, map[string]interface{}{
				"id":         id,
				"name":       name,
				"email":      email,
				"phone":      phone,
				"subject":    subject,
				"message":    message,
				"status":     status,
				"created_at": created_at,
			})
		}

		c.JSON(200, gin.H{
			"success": true,
			"message": "Contacts retrieved successfully",
			"data":    contacts,
			"meta":    gin.H{"total": len(contacts), "page": 1, "per_page": 10},
		})
	})

	router.GET("/api/v1/dashboard/contacts/stats", func(c *gin.Context) {
		db := database.GetDB()
		var total, new, inProgress, resolved int64

		db.Raw("SELECT COUNT(*) FROM contact_submissions").Scan(&total)
		db.Raw("SELECT COUNT(*) FROM contact_submissions WHERE status = 'new'").Scan(&new)
		db.Raw("SELECT COUNT(*) FROM contact_submissions WHERE status = 'in_progress'").Scan(&inProgress)
		db.Raw("SELECT COUNT(*) FROM contact_submissions WHERE status = 'resolved'").Scan(&resolved)

		c.JSON(200, gin.H{
			"success": true,
			"message": "Contact statistics retrieved successfully",
			"data": gin.H{
				"total":       total,
				"new":         new,
				"in_progress": inProgress,
				"resolved":    resolved,
			},
		})
	})

	router.POST("/api/v1/dashboard/contact", func(c *gin.Context) {
		var req struct {
			Name    string `json:"name" binding:"required"`
			Email   string `json:"email" binding:"required,email"`
			Phone   string `json:"phone"`
			Subject string `json:"subject"`
			Message string `json:"message" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"success": false, "error": err.Error()})
			return
		}

		db := database.GetDB()
		err := db.Exec("INSERT INTO contact_submissions (name, email, phone, subject, message, status) VALUES (?, ?, ?, ?, ?, 'new')",
			req.Name, req.Email, req.Phone, req.Subject, req.Message).Error

		if err != nil {
			c.JSON(500, gin.H{"success": false, "error": "Failed to create contact: " + err.Error()})
			return
		}

		c.JSON(201, gin.H{
			"success": true,
			"message": "Contact created successfully",
			"data": gin.H{
				"name":    req.Name,
				"email":   req.Email,
				"status":  "new",
				"message": "Contact submission received",
			},
		})
	})

	// Simple auth endpoint
	router.POST("/api/v1/auth/login", func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"success": false, "error": err.Error()})
			return
		}

		// Simple admin check
		if req.Email == "admin@mejona.com" && req.Password == "admin123" {
			c.JSON(200, gin.H{
				"success": true,
				"message": "Login successful",
				"data": gin.H{
					"user": gin.H{
						"id":    1,
						"email": req.Email,
						"role":  "admin",
					},
					"token": "simple-test-token",
				},
			})
			return
		}

		c.JSON(401, gin.H{"success": false, "error": "Invalid credentials"})
	})

	// Test endpoint
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Test endpoint working", "status": "success"})
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Starting Contact Service on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}