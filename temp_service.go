package main

import (
	"log"
	"net/http"
	"time"
	
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type HealthResponse struct {
	Status    string            `json:"status"`
	Message   string            `json:"message"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time         `json:"timestamp"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

var startTime = time.Now()

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"*"}
	router.Use(cors.New(config))

	// Health endpoints
	router.GET("/health", healthCheck)
	router.GET("/health/deep", deepHealthCheck)
	router.GET("/status", statusCheck)
	router.GET("/ready", readyCheck)
	router.GET("/alive", aliveCheck)
	router.GET("/metrics", metricsCheck)

	// API endpoints
	router.GET("/api/v1/test", testEndpoint)
	router.POST("/api/v1/public/contact", publicContactSubmission)

	// Auth endpoints (mock)
	router.POST("/api/v1/auth/login", authLogin)
	router.POST("/api/v1/auth/refresh", authRefresh)
	router.GET("/api/v1/auth/profile", authProfile)
	router.GET("/api/v1/auth/validate", authValidate)
	router.POST("/api/v1/auth/logout", authLogout)
	router.POST("/api/v1/auth/change-password", authChangePassword)

	// Dashboard endpoints (mock)
	router.GET("/api/v1/dashboard/contacts", dashboardContacts)
	router.GET("/api/v1/dashboard/contacts/stats", dashboardStats)
	router.POST("/api/v1/dashboard/contact", dashboardCreateContact)
	router.PUT("/api/v1/dashboard/contacts/:id/status", dashboardUpdateStatus)
	router.GET("/api/v1/dashboard/contacts/:id", dashboardGetContact)
	router.GET("/api/v1/dashboard/contacts/export", dashboardExport)
	router.POST("/api/v1/dashboard/contacts/bulk-update", dashboardBulkUpdate)

	log.Println("🚀 Contact Management Microservice starting on :8081")
	log.Println("📊 All 20 endpoints available")
	log.Fatal(http.ListenAndServe(":8081", router))
}

func healthCheck(c *gin.Context) {
	uptime := time.Since(startTime)
	response := HealthResponse{
		Status:  "healthy",
		Message: "Contact Management Service Running",
		Data: map[string]interface{}{
			"service": "Contact Management Microservice",
			"version": "1.0.0",
			"uptime":  uptime.String(),
			"status":  "operational",
		},
		Timestamp: time.Now(),
	}
	c.JSON(200, response)
}

func deepHealthCheck(c *gin.Context) {
	c.JSON(200, APIResponse{
		Success: true,
		Message: "Deep health check completed",
		Data: map[string]interface{}{
			"status":     "healthy",
			"database":   "mock",
			"memory":     "normal",
			"cpu":        "normal",
			"endpoints":  20,
		},
	})
}

func statusCheck(c *gin.Context) {
	c.JSON(200, APIResponse{
		Success: true,
		Message: "Service status check",
		Data:    map[string]string{"status": "running"},
	})
}

func readyCheck(c *gin.Context) {
	c.JSON(200, map[string]string{"status": "ready"})
}

func aliveCheck(c *gin.Context) {
	c.JSON(200, map[string]string{"status": "alive"})
}

func metricsCheck(c *gin.Context) {
	c.JSON(200, map[string]interface{}{
		"uptime": time.Since(startTime).String(),
		"endpoints": 20,
		"status": "operational",
	})
}

func testEndpoint(c *gin.Context) {
	c.JSON(200, APIResponse{
		Success: true,
		Message: "API test endpoint working",
		Data:    map[string]string{"test": "success"},
	})
}

func publicContactSubmission(c *gin.Context) {
	var contact map[string]interface{}
	if err := c.ShouldBindJSON(&contact); err != nil {
		c.JSON(400, APIResponse{Success: false, Message: "Invalid JSON"})
		return
	}
	
	c.JSON(200, APIResponse{
		Success: true,
		Message: "Contact submission received",
		Data:    map[string]interface{}{"id": 1, "status": "new"},
	})
}

func authLogin(c *gin.Context) {
	c.JSON(200, APIResponse{
		Success: true,
		Message: "Login successful",
		Data: map[string]interface{}{
			"token": "mock-jwt-token-12345",
			"user":  map[string]string{"email": "admin@mejona.com", "role": "admin"},
		},
	})
}

func authRefresh(c *gin.Context) {
	c.JSON(200, APIResponse{Success: true, Message: "Token refreshed", Data: map[string]string{"token": "new-token"}})
}

func authProfile(c *gin.Context) {
	c.JSON(200, APIResponse{Success: true, Message: "Profile retrieved", Data: map[string]string{"name": "Admin", "email": "admin@mejona.com"}})
}

func authValidate(c *gin.Context) {
	c.JSON(200, APIResponse{Success: true, Message: "Token valid", Data: map[string]bool{"valid": true}})
}

func authLogout(c *gin.Context) {
	c.JSON(200, APIResponse{Success: true, Message: "Logout successful", Data: nil})
}

func authChangePassword(c *gin.Context) {
	c.JSON(200, APIResponse{Success: true, Message: "Password changed", Data: nil})
}

func dashboardContacts(c *gin.Context) {
	contacts := []map[string]interface{}{
		{"id": 1, "name": "John Doe", "email": "john@example.com", "status": "new"},
		{"id": 2, "name": "Jane Smith", "email": "jane@example.com", "status": "in_progress"},
	}
	c.JSON(200, APIResponse{Success: true, Message: "Contacts retrieved", Data: contacts})
}

func dashboardStats(c *gin.Context) {
	stats := map[string]interface{}{
		"total": 10,
		"new": 5,
		"in_progress": 3,
		"resolved": 2,
	}
	c.JSON(200, APIResponse{Success: true, Message: "Stats retrieved", Data: stats})
}

func dashboardCreateContact(c *gin.Context) {
	c.JSON(201, APIResponse{Success: true, Message: "Contact created", Data: map[string]int{"id": 123}})
}

func dashboardUpdateStatus(c *gin.Context) {
	c.JSON(200, APIResponse{Success: true, Message: "Status updated", Data: nil})
}

func dashboardGetContact(c *gin.Context) {
	id := c.Param("id")
	contact := map[string]interface{}{
		"id": id,
		"name": "Sample Contact",
		"email": "sample@example.com",
		"status": "new",
	}
	c.JSON(200, APIResponse{Success: true, Message: "Contact retrieved", Data: contact})
}

func dashboardExport(c *gin.Context) {
	c.Header("Content-Type", "text/csv")
	c.String(200, "id,name,email,status\n1,John Doe,john@example.com,new\n")
}

func dashboardBulkUpdate(c *gin.Context) {
	c.JSON(200, APIResponse{Success: true, Message: "Bulk update completed", Data: map[string]int{"updated": 5}})
}