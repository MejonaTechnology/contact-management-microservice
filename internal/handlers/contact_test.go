package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"contact-service/internal/models"
	"contact-service/internal/services"
)

// MockContactService is a mock implementation of ContactService for testing
type MockContactService struct {
	mock.Mock
}

func (m *MockContactService) CreateContact(contact *models.Contact) error {
	args := m.Called(contact)
	return args.Error(0)
}

func (m *MockContactService) GetContact(id uint) (*models.Contact, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Contact), args.Error(1)
}

func (m *MockContactService) UpdateContact(contact *models.Contact) error {
	args := m.Called(contact)
	return args.Error(0)
}

func (m *MockContactService) DeleteContact(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockContactService) ListContacts(params services.ContactListParams) ([]models.Contact, int64, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.Contact), args.Get(1).(int64), args.Error(2)
}

func (m *MockContactService) SearchContacts(query string, params services.ContactListParams) ([]models.Contact, int64, error) {
	args := m.Called(query, params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.Contact), args.Get(1).(int64), args.Error(2)
}

func (m *MockContactService) UpdateContactStatus(id uint, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockContactService) AssignContact(id uint, userID uint) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

// ContactHandlerTestSuite defines the test suite for ContactHandler
type ContactHandlerTestSuite struct {
	suite.Suite
	handler    *ContactHandler
	mockService *MockContactService
	router     *gin.Engine
}

// SetupSuite runs before all tests in the suite
func (suite *ContactHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
}

// SetupTest runs before each test
func (suite *ContactHandlerTestSuite) SetupTest() {
	suite.mockService = new(MockContactService)
	suite.handler = &ContactHandler{
		contactService: suite.mockService,
	}
	suite.router = gin.New()
	
	// Setup routes
	v1 := suite.router.Group("/api/v1")
	{
		contacts := v1.Group("/contacts")
		{
			contacts.GET("", suite.handler.ListContacts)
			contacts.POST("", suite.handler.CreateContact)
			contacts.GET("/:id", suite.handler.GetContact)
			contacts.PUT("/:id", suite.handler.UpdateContact)
			contacts.DELETE("/:id", suite.handler.DeleteContact)
			contacts.PUT("/:id/status", suite.handler.UpdateContactStatus)
			contacts.GET("/search", suite.handler.SearchContacts)
		}
	}
}

// TearDownTest runs after each test
func (suite *ContactHandlerTestSuite) TearDownTest() {
	suite.mockService.AssertExpectations(suite.T())
}

// Test CreateContact endpoint
func (suite *ContactHandlerTestSuite) TestCreateContact_Success() {
	// Arrange
	contactData := CreateContactRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Phone:    "+1-555-123-4567",
		Company:  "Acme Corp",
		Position: "Manager",
		TypeID:   1,
		SourceID: 1,
		Notes:    "Test contact",
	}

	expectedContact := &models.Contact{
		ID:        1,
		Name:      contactData.Name,
		Email:     contactData.Email,
		Phone:     contactData.Phone,
		Company:   contactData.Company,
		Position:  contactData.Position,
		TypeID:    contactData.TypeID,
		SourceID:  contactData.SourceID,
		Notes:     contactData.Notes,
		Status:    "new",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	suite.mockService.On("CreateContact", mock.AnythingOfType("*models.Contact")).Return(nil).Run(func(args mock.Arguments) {
		contact := args.Get(0).(*models.Contact)
		contact.ID = 1
		contact.Status = "new"
		contact.CreatedAt = expectedContact.CreatedAt
		contact.UpdatedAt = expectedContact.UpdatedAt
	})

	jsonData, _ := json.Marshal(contactData)
	req := httptest.NewRequest("POST", "/api/v1/contacts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "Contact created successfully", response.Message)
	assert.NotNil(suite.T(), response.Data)
}

func (suite *ContactHandlerTestSuite) TestCreateContact_ValidationError() {
	// Arrange - invalid data (missing required fields)
	contactData := CreateContactRequest{
		Name: "", // Empty name should fail validation
		Email: "invalid-email", // Invalid email format
	}

	jsonData, _ := json.Marshal(contactData)
	req := httptest.NewRequest("POST", "/api/v1/contacts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
	assert.Contains(suite.T(), response.Message, "validation")
}

// Test GetContact endpoint
func (suite *ContactHandlerTestSuite) TestGetContact_Success() {
	// Arrange
	contactID := uint(1)
	expectedContact := &models.Contact{
		ID:        contactID,
		Name:      "John Doe",
		Email:     "john@example.com",
		Phone:     "+1-555-123-4567",
		Company:   "Acme Corp",
		Position:  "Manager",
		Status:    "new",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	suite.mockService.On("GetContact", contactID).Return(expectedContact, nil)

	req := httptest.NewRequest("GET", "/api/v1/contacts/1", nil)
	w := httptest.NewRecorder()

	// Act
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)
}

func (suite *ContactHandlerTestSuite) TestGetContact_NotFound() {
	// Arrange
	contactID := uint(999)
	suite.mockService.On("GetContact", contactID).Return(nil, services.ErrContactNotFound)

	req := httptest.NewRequest("GET", "/api/v1/contacts/999", nil)
	w := httptest.NewRecorder()

	// Act
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
	assert.Contains(suite.T(), response.Message, "not found")
}

// Test ListContacts endpoint
func (suite *ContactHandlerTestSuite) TestListContacts_Success() {
	// Arrange
	expectedContacts := []models.Contact{
		{
			ID:        1,
			Name:      "John Doe",
			Email:     "john@example.com",
			Status:    "new",
			CreatedAt: time.Now(),
		},
		{
			ID:        2,
			Name:      "Jane Smith",
			Email:     "jane@example.com",
			Status:    "contacted",
			CreatedAt: time.Now(),
		},
	}
	totalCount := int64(2)

	suite.mockService.On("ListContacts", mock.AnythingOfType("services.ContactListParams")).Return(expectedContacts, totalCount, nil)

	req := httptest.NewRequest("GET", "/api/v1/contacts?page=1&limit=10", nil)
	w := httptest.NewRecorder()

	// Act
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)
	assert.NotNil(suite.T(), response.Meta)
}

// Test UpdateContact endpoint
func (suite *ContactHandlerTestSuite) TestUpdateContact_Success() {
	// Arrange
	contactID := uint(1)
	originalContact := &models.Contact{
		ID:       contactID,
		Name:     "John Doe",
		Email:    "john@example.com",
		Status:   "new",
	}

	updateData := UpdateContactRequest{
		Name:     "John Updated",
		Email:    "john.updated@example.com",
		Company:  "Updated Corp",
		Position: "Senior Manager",
	}

	suite.mockService.On("GetContact", contactID).Return(originalContact, nil)
	suite.mockService.On("UpdateContact", mock.AnythingOfType("*models.Contact")).Return(nil)

	jsonData, _ := json.Marshal(updateData)
	req := httptest.NewRequest("PUT", "/api/v1/contacts/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "Contact updated successfully", response.Message)
}

// Test DeleteContact endpoint
func (suite *ContactHandlerTestSuite) TestDeleteContact_Success() {
	// Arrange
	contactID := uint(1)
	suite.mockService.On("DeleteContact", contactID).Return(nil)

	req := httptest.NewRequest("DELETE", "/api/v1/contacts/1", nil)
	w := httptest.NewRecorder()

	// Act
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "Contact deleted successfully", response.Message)
}

// Test UpdateContactStatus endpoint
func (suite *ContactHandlerTestSuite) TestUpdateContactStatus_Success() {
	// Arrange
	contactID := uint(1)
	statusUpdate := StatusUpdateRequest{
		Status: "contacted",
	}

	suite.mockService.On("UpdateContactStatus", contactID, statusUpdate.Status).Return(nil)

	jsonData, _ := json.Marshal(statusUpdate)
	req := httptest.NewRequest("PUT", "/api/v1/contacts/1/status", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "Contact status updated successfully", response.Message)
}

// Test SearchContacts endpoint
func (suite *ContactHandlerTestSuite) TestSearchContacts_Success() {
	// Arrange
	searchQuery := "john"
	expectedContacts := []models.Contact{
		{
			ID:    1,
			Name:  "John Doe",
			Email: "john@example.com",
		},
	}
	totalCount := int64(1)

	suite.mockService.On("SearchContacts", searchQuery, mock.AnythingOfType("services.ContactListParams")).Return(expectedContacts, totalCount, nil)

	req := httptest.NewRequest("GET", "/api/v1/contacts/search?q="+searchQuery, nil)
	w := httptest.NewRecorder()

	// Act
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)
}

// Helper function to convert string to uint
func parseID(idStr string) uint {
	id, _ := strconv.ParseUint(idStr, 10, 32)
	return uint(id)
}

// Run the test suite
func TestContactHandlerSuite(t *testing.T) {
	suite.Run(t, new(ContactHandlerTestSuite))
}

// Individual test functions for running with go test
func TestCreateContact(t *testing.T) {
	suite.Run(t, new(ContactHandlerTestSuite))
}

func TestGetContact(t *testing.T) {
	suite.Run(t, new(ContactHandlerTestSuite))
}

func TestListContacts(t *testing.T) {
	suite.Run(t, new(ContactHandlerTestSuite))
}

func TestUpdateContact(t *testing.T) {
	suite.Run(t, new(ContactHandlerTestSuite))
}

func TestDeleteContact(t *testing.T) {
	suite.Run(t, new(ContactHandlerTestSuite))
}