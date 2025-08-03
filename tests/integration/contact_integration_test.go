// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"contact-service/internal/handlers"
	"contact-service/internal/middleware"
	"contact-service/internal/models"
	"contact-service/internal/repository"
	"contact-service/internal/services"
	"contact-service/pkg/database"
)

// ContactIntegrationTestSuite defines the integration test suite
type ContactIntegrationTestSuite struct {
	suite.Suite
	db     *gorm.DB
	router *gin.Engine
	server *httptest.Server
	authToken string
}

// SetupSuite runs once before all tests in the suite
func (suite *ContactIntegrationTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	
	// Setup test database
	var err error
	suite.db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		suite.T().Fatalf("Failed to connect to test database: %v", err)
	}

	// Set global database connection
	database.DB = suite.db

	// Run migrations
	err = suite.db.AutoMigrate(
		&models.Contact{},
		&models.ContactType{},
		&models.ContactSource{},
		&models.AdminUser{},
		&models.Assignment{},
		&models.Appointment{},
	)
	if err != nil {
		suite.T().Fatalf("Failed to run migrations: %v", err)
	}

	// Seed test data
	suite.seedTestData()

	// Setup router
	suite.setupRouter()

	// Start test server
	suite.server = httptest.NewServer(suite.router)

	// Get auth token
	suite.authToken = suite.getAuthToken()
}

// TearDownSuite runs once after all tests in the suite
func (suite *ContactIntegrationTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
}

// SetupTest runs before each test
func (suite *ContactIntegrationTestSuite) SetupTest() {
	// Clean up contacts table for each test
	suite.db.Exec("DELETE FROM contacts")
}

// setupRouter configures the test router with all routes and middleware
func (suite *ContactIntegrationTestSuite) setupRouter() {
	suite.router = gin.New()
	suite.router.Use(gin.Recovery())

	// Setup repositories
	contactRepo := repository.NewContactRepository(suite.db)
	
	// Setup services
	contactService := services.NewContactService(contactRepo)
	
	// Setup handlers
	contactHandler := handlers.NewContactHandler(contactService)
	authHandler := handlers.NewAuthHandler(nil) // Mock auth for tests

	// Setup routes
	api := suite.router.Group("/api/v1")
	
	// Auth routes (for getting test token)
	auth := api.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
	}

	// Protected routes
	contacts := api.Group("/contacts")
	contacts.Use(middleware.AuthMiddleware()) // This will be mocked for tests
	{
		contacts.GET("", contactHandler.ListContacts)
		contacts.POST("", contactHandler.CreateContact)
		contacts.GET("/:id", contactHandler.GetContact)
		contacts.PUT("/:id", contactHandler.UpdateContact)
		contacts.DELETE("/:id", contactHandler.DeleteContact)
		contacts.PUT("/:id/status", contactHandler.UpdateContactStatus)
		contacts.GET("/search", contactHandler.SearchContacts)
	}

	// Public routes
	public := api.Group("/public")
	{
		public.POST("/contact", contactHandler.SubmitContact)
	}
}

// seedTestData creates initial test data
func (suite *ContactIntegrationTestSuite) seedTestData() {
	// Create test admin user
	adminUser := &models.AdminUser{
		Name:     "Test Admin",
		Email:    "admin@test.com",
		Password: "$2a$10$example.hash", // bcrypt hash for "password123"
		Role:     "admin",
	}
	suite.db.Create(adminUser)

	// Create test contact types
	contactTypes := []models.ContactType{
		{Name: "Sales Inquiry", Description: "General sales inquiries"},
		{Name: "Support Request", Description: "Technical support requests"},
		{Name: "Partnership", Description: "Business partnership inquiries"},
	}
	suite.db.Create(&contactTypes)

	// Create test contact sources
	contactSources := []models.ContactSource{
		{Name: "Website", Description: "Website contact form"},
		{Name: "Email", Description: "Direct email"},
		{Name: "Phone", Description: "Phone call"},
		{Name: "Referral", Description: "Customer referral"},
	}
	suite.db.Create(&contactSources)
}

// getAuthToken gets a valid JWT token for testing
func (suite *ContactIntegrationTestSuite) getAuthToken() string {
	// For integration tests, we'll use a mock token
	return "Bearer test-jwt-token"
}

// makeAuthenticatedRequest makes an HTTP request with authentication header
func (suite *ContactIntegrationTestSuite) makeAuthenticatedRequest(method, url string, body interface{}) *http.Response {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req, _ := http.NewRequest(method, suite.server.URL+url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", suite.authToken)

	client := &http.Client{}
	resp, _ := client.Do(req)
	return resp
}

// Test CreateContact endpoint
func (suite *ContactIntegrationTestSuite) TestCreateContact_Success() {
	// Arrange
	contactData := map[string]interface{}{
		"name":      "John Doe",
		"email":     "john@example.com",
		"phone":     "+1-555-123-4567",
		"company":   "Acme Corp",
		"position":  "Manager",
		"type_id":   1,
		"source_id": 1,
		"notes":     "Test contact",
	}

	// Act
	resp := suite.makeAuthenticatedRequest("POST", "/api/v1/contacts", contactData)
	defer resp.Body.Close()

	// Assert
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	var response handlers.APIResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "Contact created successfully", response.Message)
	assert.NotNil(suite.T(), response.Data)

	// Verify contact was created in database
	var contact models.Contact
	err = suite.db.Where("email = ?", "john@example.com").First(&contact).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "John Doe", contact.Name)
	assert.Equal(suite.T(), "new", contact.Status)
}

func (suite *ContactIntegrationTestSuite) TestCreateContact_ValidationError() {
	// Arrange - invalid data
	contactData := map[string]interface{}{
		"name":  "", // Empty name
		"email": "invalid-email", // Invalid email
	}

	// Act
	resp := suite.makeAuthenticatedRequest("POST", "/api/v1/contacts", contactData)
	defer resp.Body.Close()

	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)

	var response handlers.APIResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
	assert.Contains(suite.T(), response.Message, "validation")
}

func (suite *ContactIntegrationTestSuite) TestCreateContact_DuplicateEmail() {
	// Arrange - create existing contact
	existingContact := &models.Contact{
		Name:   "Existing User",
		Email:  "existing@example.com",
		Status: "new",
	}
	suite.db.Create(existingContact)

	contactData := map[string]interface{}{
		"name":  "New User",
		"email": "existing@example.com", // Same email
	}

	// Act
	resp := suite.makeAuthenticatedRequest("POST", "/api/v1/contacts", contactData)
	defer resp.Body.Close()

	// Assert
	assert.Equal(suite.T(), http.StatusConflict, resp.StatusCode)

	var response handlers.APIResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
	assert.Contains(suite.T(), response.Message, "already exists")
}

// Test GetContact endpoint
func (suite *ContactIntegrationTestSuite) TestGetContact_Success() {
	// Arrange - create test contact
	contact := &models.Contact{
		Name:     "John Doe",
		Email:    "john@example.com",
		Phone:    "+1-555-123-4567",
		Company:  "Acme Corp",
		Status:   "new",
		TypeID:   1,
		SourceID: 1,
	}
	suite.db.Create(contact)

	// Act
	resp := suite.makeAuthenticatedRequest("GET", fmt.Sprintf("/api/v1/contacts/%d", contact.ID), nil)
	defer resp.Body.Close()

	// Assert
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var response handlers.APIResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)

	// Check response data
	contactData := response.Data.(map[string]interface{})
	assert.Equal(suite.T(), "John Doe", contactData["name"])
	assert.Equal(suite.T(), "john@example.com", contactData["email"])
}

func (suite *ContactIntegrationTestSuite) TestGetContact_NotFound() {
	// Act
	resp := suite.makeAuthenticatedRequest("GET", "/api/v1/contacts/999", nil)
	defer resp.Body.Close()

	// Assert
	assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)

	var response handlers.APIResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
	assert.Contains(suite.T(), response.Message, "not found")
}

// Test ListContacts endpoint
func (suite *ContactIntegrationTestSuite) TestListContacts_Success() {
	// Arrange - create test contacts
	contacts := []models.Contact{
		{
			Name:   "John Doe",
			Email:  "john@example.com",
			Status: "new",
		},
		{
			Name:   "Jane Smith",
			Email:  "jane@example.com",
			Status: "contacted",
		},
		{
			Name:   "Bob Johnson",
			Email:  "bob@example.com",
			Status: "new",
		},
	}
	suite.db.Create(&contacts)

	// Act
	resp := suite.makeAuthenticatedRequest("GET", "/api/v1/contacts?page=1&limit=10", nil)
	defer resp.Body.Close()

	// Assert
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var response handlers.APIResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)
	assert.NotNil(suite.T(), response.Meta)

	// Check pagination metadata
	meta := response.Meta.(map[string]interface{})
	assert.Equal(suite.T(), float64(3), meta["total"])
	assert.Equal(suite.T(), float64(1), meta["page"])
}

func (suite *ContactIntegrationTestSuite) TestListContacts_WithFilters() {
	// Arrange - create test contacts with different statuses
	contacts := []models.Contact{
		{Name: "John Doe", Email: "john@example.com", Status: "new"},
		{Name: "Jane Smith", Email: "jane@example.com", Status: "contacted"},
		{Name: "Bob Johnson", Email: "bob@example.com", Status: "new"},
	}
	suite.db.Create(&contacts)

	// Act - filter by status
	resp := suite.makeAuthenticatedRequest("GET", "/api/v1/contacts?status=new", nil)
	defer resp.Body.Close()

	// Assert
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var response handlers.APIResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	// Should return only contacts with "new" status
	meta := response.Meta.(map[string]interface{})
	assert.Equal(suite.T(), float64(2), meta["total"]) // 2 contacts with "new" status
}

// Test UpdateContact endpoint
func (suite *ContactIntegrationTestSuite) TestUpdateContact_Success() {
	// Arrange - create test contact
	contact := &models.Contact{
		Name:   "John Doe",
		Email:  "john@example.com",
		Status: "new",
	}
	suite.db.Create(contact)

	updateData := map[string]interface{}{
		"name":     "John Updated",
		"email":    "john.updated@example.com",
		"company":  "Updated Corp",
		"position": "Senior Manager",
	}

	// Act
	resp := suite.makeAuthenticatedRequest("PUT", fmt.Sprintf("/api/v1/contacts/%d", contact.ID), updateData)
	defer resp.Body.Close()

	// Assert
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var response handlers.APIResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	// Verify contact was updated in database
	var updatedContact models.Contact
	suite.db.First(&updatedContact, contact.ID)
	assert.Equal(suite.T(), "John Updated", updatedContact.Name)
	assert.Equal(suite.T(), "john.updated@example.com", updatedContact.Email)
	assert.Equal(suite.T(), "Updated Corp", updatedContact.Company)
}

// Test DeleteContact endpoint
func (suite *ContactIntegrationTestSuite) TestDeleteContact_Success() {
	// Arrange - create test contact
	contact := &models.Contact{
		Name:   "John Doe",
		Email:  "john@example.com",
		Status: "new",
	}
	suite.db.Create(contact)

	// Act
	resp := suite.makeAuthenticatedRequest("DELETE", fmt.Sprintf("/api/v1/contacts/%d", contact.ID), nil)
	defer resp.Body.Close()

	// Assert
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var response handlers.APIResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	// Verify contact was deleted from database
	var deletedContact models.Contact
	err = suite.db.First(&deletedContact, contact.ID).Error
	assert.Error(suite.T(), err) // Should return "record not found" error
}

// Test UpdateContactStatus endpoint
func (suite *ContactIntegrationTestSuite) TestUpdateContactStatus_Success() {
	// Arrange - create test contact
	contact := &models.Contact{
		Name:   "John Doe",
		Email:  "john@example.com",
		Status: "new",
	}
	suite.db.Create(contact)

	statusUpdate := map[string]interface{}{
		"status": "contacted",
	}

	// Act
	resp := suite.makeAuthenticatedRequest("PUT", fmt.Sprintf("/api/v1/contacts/%d/status", contact.ID), statusUpdate)
	defer resp.Body.Close()

	// Assert
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var response handlers.APIResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	// Verify status was updated in database
	var updatedContact models.Contact
	suite.db.First(&updatedContact, contact.ID)
	assert.Equal(suite.T(), "contacted", updatedContact.Status)
}

// Test SearchContacts endpoint
func (suite *ContactIntegrationTestSuite) TestSearchContacts_Success() {
	// Arrange - create test contacts
	contacts := []models.Contact{
		{Name: "John Doe", Email: "john@example.com", Company: "Acme Corp"},
		{Name: "Jane Smith", Email: "jane@example.com", Company: "Beta Inc"},
		{Name: "John Johnson", Email: "johnj@example.com", Company: "Gamma LLC"},
	}
	suite.db.Create(&contacts)

	// Act - search for "john"
	resp := suite.makeAuthenticatedRequest("GET", "/api/v1/contacts/search?q=john", nil)
	defer resp.Body.Close()

	// Assert
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var response handlers.APIResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	// Should return contacts with "john" in name or email
	meta := response.Meta.(map[string]interface{})
	assert.True(suite.T(), meta["total"].(float64) >= 2) // At least 2 matches
}

// Test Public Contact Submission endpoint
func (suite *ContactIntegrationTestSuite) TestSubmitContact_Success() {
	// Arrange
	contactData := map[string]interface{}{
		"name":    "Public User",
		"email":   "public@example.com",
		"phone":   "+1-555-987-6543",
		"message": "I'm interested in your services",
	}

	// Act - public endpoint doesn't need authentication
	req, _ := http.NewRequest("POST", suite.server.URL+"/api/v1/public/contact", bytes.NewBuffer(nil))
	jsonData, _ := json.Marshal(contactData)
	req.Body = http.NoBody
	req.Body = httptest.NewRecorder().Body
	req.Body = httptest.NewRecorder().Result().Body
	req = httptest.NewRequest("POST", "/api/v1/public/contact", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response handlers.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	// Verify contact was created in database
	var contact models.Contact
	err = suite.db.Where("email = ?", "public@example.com").First(&contact).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Public User", contact.Name)
	assert.Equal(suite.T(), "new", contact.Status)
}

// Test Performance and Load
func (suite *ContactIntegrationTestSuite) TestListContacts_Performance() {
	// Arrange - create many test contacts
	var contacts []models.Contact
	for i := 0; i < 1000; i++ {
		contacts = append(contacts, models.Contact{
			Name:   fmt.Sprintf("User %d", i),
			Email:  fmt.Sprintf("user%d@example.com", i),
			Status: "new",
		})
	}
	suite.db.CreateInBatches(&contacts, 100)

	// Act - measure response time
	start := time.Now()
	resp := suite.makeAuthenticatedRequest("GET", "/api/v1/contacts?page=1&limit=50", nil)
	duration := time.Since(start)
	defer resp.Body.Close()

	// Assert
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	assert.Less(suite.T(), duration, 1*time.Second) // Should respond within 1 second

	var response handlers.APIResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	meta := response.Meta.(map[string]interface{})
	assert.Equal(suite.T(), float64(1000), meta["total"])
}

// Run the test suite
func TestContactIntegrationSuite(t *testing.T) {
	// Skip integration tests if not explicitly requested
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests. Set RUN_INTEGRATION_TESTS=true to run.")
	}

	suite.Run(t, new(ContactIntegrationTestSuite))
}