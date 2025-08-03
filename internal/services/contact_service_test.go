package services

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"contact-service/internal/models"
)

// MockContactRepository is a mock implementation of ContactRepository
type MockContactRepository struct {
	mock.Mock
}

func (m *MockContactRepository) Create(contact *models.Contact) error {
	args := m.Called(contact)
	return args.Error(0)
}

func (m *MockContactRepository) GetByID(id uint) (*models.Contact, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Contact), args.Error(1)
}

func (m *MockContactRepository) Update(contact *models.Contact) error {
	args := m.Called(contact)
	return args.Error(0)
}

func (m *MockContactRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockContactRepository) List(params ContactListParams) ([]models.Contact, int64, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.Contact), args.Get(1).(int64), args.Error(2)
}

func (m *MockContactRepository) Search(query string, params ContactListParams) ([]models.Contact, int64, error) {
	args := m.Called(query, params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.Contact), args.Get(1).(int64), args.Error(2)
}

func (m *MockContactRepository) GetByEmail(email string) (*models.Contact, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Contact), args.Error(1)
}

func (m *MockContactRepository) UpdateStatus(id uint, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockContactRepository) Assign(id uint, userID uint) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

func (m *MockContactRepository) GetAssignedContacts(userID uint) ([]models.Contact, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Contact), args.Error(1)
}

// ContactServiceTestSuite defines the test suite for ContactService
type ContactServiceTestSuite struct {
	suite.Suite
	service    *ContactService
	mockRepo   *MockContactRepository
}

// SetupTest runs before each test
func (suite *ContactServiceTestSuite) SetupTest() {
	suite.mockRepo = new(MockContactRepository)
	suite.service = &ContactService{
		repo: suite.mockRepo,
	}
}

// TearDownTest runs after each test
func (suite *ContactServiceTestSuite) TearDownTest() {
	suite.mockRepo.AssertExpectations(suite.T())
}

// Test CreateContact
func (suite *ContactServiceTestSuite) TestCreateContact_Success() {
	// Arrange
	contact := &models.Contact{
		Name:     "John Doe",
		Email:    "john@example.com",
		Phone:    "+1-555-123-4567",
		Company:  "Acme Corp",
		Position: "Manager",
		Status:   "new",
	}

	suite.mockRepo.On("GetByEmail", contact.Email).Return(nil, ErrContactNotFound)
	suite.mockRepo.On("Create", contact).Return(nil).Run(func(args mock.Arguments) {
		c := args.Get(0).(*models.Contact)
		c.ID = 1
		c.CreatedAt = time.Now()
		c.UpdatedAt = time.Now()
	})

	// Act
	err := suite.service.CreateContact(contact)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), uint(1), contact.ID)
	assert.Equal(suite.T(), "new", contact.Status)
	assert.NotZero(suite.T(), contact.CreatedAt)
	assert.NotZero(suite.T(), contact.UpdatedAt)
}

func (suite *ContactServiceTestSuite) TestCreateContact_DuplicateEmail() {
	// Arrange
	existingContact := &models.Contact{
		ID:    1,
		Email: "john@example.com",
	}
	
	newContact := &models.Contact{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	suite.mockRepo.On("GetByEmail", newContact.Email).Return(existingContact, nil)

	// Act
	err := suite.service.CreateContact(newContact)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrContactDuplicateEmail, err)
}

func (suite *ContactServiceTestSuite) TestCreateContact_ValidationError() {
	// Arrange - invalid contact (missing required fields)
	contact := &models.Contact{
		Name:  "", // Empty name
		Email: "", // Empty email
	}

	// Act
	err := suite.service.CreateContact(contact)

	// Assert
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "validation")
}

// Test GetContact
func (suite *ContactServiceTestSuite) TestGetContact_Success() {
	// Arrange
	contactID := uint(1)
	expectedContact := &models.Contact{
		ID:        contactID,
		Name:      "John Doe",
		Email:     "john@example.com",
		Status:    "new",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	suite.mockRepo.On("GetByID", contactID).Return(expectedContact, nil)

	// Act
	contact, err := suite.service.GetContact(contactID)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), contact)
	assert.Equal(suite.T(), expectedContact.ID, contact.ID)
	assert.Equal(suite.T(), expectedContact.Name, contact.Name)
	assert.Equal(suite.T(), expectedContact.Email, contact.Email)
}

func (suite *ContactServiceTestSuite) TestGetContact_NotFound() {
	// Arrange
	contactID := uint(999)
	suite.mockRepo.On("GetByID", contactID).Return(nil, ErrContactNotFound)

	// Act
	contact, err := suite.service.GetContact(contactID)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), contact)
	assert.Equal(suite.T(), ErrContactNotFound, err)
}

func (suite *ContactServiceTestSuite) TestGetContact_InvalidID() {
	// Arrange
	contactID := uint(0)

	// Act
	contact, err := suite.service.GetContact(contactID)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), contact)
	assert.Contains(suite.T(), err.Error(), "invalid")
}

// Test UpdateContact
func (suite *ContactServiceTestSuite) TestUpdateContact_Success() {
	// Arrange
	existingContact := &models.Contact{
		ID:        1,
		Name:      "John Doe",
		Email:     "john@example.com",
		Status:    "new",
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now().Add(-time.Hour),
	}

	updatedContact := &models.Contact{
		ID:       1,
		Name:     "John Updated",
		Email:    "john.updated@example.com",
		Company:  "Updated Corp",
		Position: "Senior Manager",
		Status:   "contacted",
	}

	suite.mockRepo.On("GetByID", updatedContact.ID).Return(existingContact, nil)
	suite.mockRepo.On("Update", mock.AnythingOfType("*models.Contact")).Return(nil)

	// Act
	err := suite.service.UpdateContact(updatedContact)

	// Assert
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), updatedContact.UpdatedAt.After(existingContact.UpdatedAt))
}

func (suite *ContactServiceTestSuite) TestUpdateContact_NotFound() {
	// Arrange
	contact := &models.Contact{
		ID:   999,
		Name: "Non-existent Contact",
	}

	suite.mockRepo.On("GetByID", contact.ID).Return(nil, ErrContactNotFound)

	// Act
	err := suite.service.UpdateContact(contact)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrContactNotFound, err)
}

// Test DeleteContact
func (suite *ContactServiceTestSuite) TestDeleteContact_Success() {
	// Arrange
	contactID := uint(1)
	existingContact := &models.Contact{
		ID:     contactID,
		Name:   "John Doe",
		Status: "new",
	}

	suite.mockRepo.On("GetByID", contactID).Return(existingContact, nil)
	suite.mockRepo.On("Delete", contactID).Return(nil)

	// Act
	err := suite.service.DeleteContact(contactID)

	// Assert
	assert.NoError(suite.T(), err)
}

func (suite *ContactServiceTestSuite) TestDeleteContact_NotFound() {
	// Arrange
	contactID := uint(999)
	suite.mockRepo.On("GetByID", contactID).Return(nil, ErrContactNotFound)

	// Act
	err := suite.service.DeleteContact(contactID)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrContactNotFound, err)
}

func (suite *ContactServiceTestSuite) TestDeleteContact_InvalidID() {
	// Arrange
	contactID := uint(0)

	// Act
	err := suite.service.DeleteContact(contactID)

	// Assert
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid")
}

// Test ListContacts
func (suite *ContactServiceTestSuite) TestListContacts_Success() {
	// Arrange
	params := ContactListParams{
		Page:     1,
		Limit:    10,
		Sort:     "created_at",
		Order:    "desc",
		Status:   "new",
	}

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
			Status:    "new",
			CreatedAt: time.Now().Add(-time.Hour),
		},
	}
	totalCount := int64(2)

	suite.mockRepo.On("List", params).Return(expectedContacts, totalCount, nil)

	// Act
	contacts, total, err := suite.service.ListContacts(params)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), contacts)
	assert.Len(suite.T(), contacts, 2)
	assert.Equal(suite.T(), totalCount, total)
	assert.Equal(suite.T(), expectedContacts[0].ID, contacts[0].ID)
	assert.Equal(suite.T(), expectedContacts[1].ID, contacts[1].ID)
}

func (suite *ContactServiceTestSuite) TestListContacts_EmptyResult() {
	// Arrange
	params := ContactListParams{
		Page:   1,
		Limit:  10,
		Status: "non-existent",
	}

	expectedContacts := []models.Contact{}
	totalCount := int64(0)

	suite.mockRepo.On("List", params).Return(expectedContacts, totalCount, nil)

	// Act
	contacts, total, err := suite.service.ListContacts(params)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), contacts)
	assert.Len(suite.T(), contacts, 0)
	assert.Equal(suite.T(), int64(0), total)
}

// Test SearchContacts
func (suite *ContactServiceTestSuite) TestSearchContacts_Success() {
	// Arrange
	query := "john"
	params := ContactListParams{
		Page:  1,
		Limit: 10,
	}

	expectedContacts := []models.Contact{
		{
			ID:    1,
			Name:  "John Doe",
			Email: "john@example.com",
		},
		{
			ID:      2,
			Name:    "Jane Johnson",
			Email:   "jane@example.com",
			Company: "Johnson & Associates",
		},
	}
	totalCount := int64(2)

	suite.mockRepo.On("Search", query, params).Return(expectedContacts, totalCount, nil)

	// Act
	contacts, total, err := suite.service.SearchContacts(query, params)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), contacts)
	assert.Len(suite.T(), contacts, 2)
	assert.Equal(suite.T(), totalCount, total)
}

func (suite *ContactServiceTestSuite) TestSearchContacts_EmptyQuery() {
	// Arrange
	query := ""
	params := ContactListParams{
		Page:  1,
		Limit: 10,
	}

	// Act
	contacts, total, err := suite.service.SearchContacts(query, params)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), contacts)
	assert.Equal(suite.T(), int64(0), total)
	assert.Contains(suite.T(), err.Error(), "empty")
}

// Test UpdateContactStatus
func (suite *ContactServiceTestSuite) TestUpdateContactStatus_Success() {
	// Arrange
	contactID := uint(1)
	newStatus := "contacted"

	existingContact := &models.Contact{
		ID:     contactID,
		Status: "new",
	}

	suite.mockRepo.On("GetByID", contactID).Return(existingContact, nil)
	suite.mockRepo.On("UpdateStatus", contactID, newStatus).Return(nil)

	// Act
	err := suite.service.UpdateContactStatus(contactID, newStatus)

	// Assert
	assert.NoError(suite.T(), err)
}

func (suite *ContactServiceTestSuite) TestUpdateContactStatus_InvalidStatus() {
	// Arrange
	contactID := uint(1)
	invalidStatus := "invalid-status"

	// Act
	err := suite.service.UpdateContactStatus(contactID, invalidStatus)

	// Assert
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid status")
}

func (suite *ContactServiceTestSuite) TestUpdateContactStatus_NotFound() {
	// Arrange
	contactID := uint(999)
	status := "contacted"

	suite.mockRepo.On("GetByID", contactID).Return(nil, ErrContactNotFound)

	// Act
	err := suite.service.UpdateContactStatus(contactID, status)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrContactNotFound, err)
}

// Test AssignContact
func (suite *ContactServiceTestSuite) TestAssignContact_Success() {
	// Arrange
	contactID := uint(1)
	userID := uint(10)

	existingContact := &models.Contact{
		ID:         contactID,
		AssignedTo: nil,
	}

	suite.mockRepo.On("GetByID", contactID).Return(existingContact, nil)
	suite.mockRepo.On("Assign", contactID, userID).Return(nil)

	// Act
	err := suite.service.AssignContact(contactID, userID)

	// Assert
	assert.NoError(suite.T(), err)
}

func (suite *ContactServiceTestSuite) TestAssignContact_NotFound() {
	// Arrange
	contactID := uint(999)
	userID := uint(10)

	suite.mockRepo.On("GetByID", contactID).Return(nil, ErrContactNotFound)

	// Act
	err := suite.service.AssignContact(contactID, userID)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrContactNotFound, err)
}

func (suite *ContactServiceTestSuite) TestAssignContact_InvalidUserID() {
	// Arrange
	contactID := uint(1)
	userID := uint(0)

	// Act
	err := suite.service.AssignContact(contactID, userID)

	// Assert
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid user ID")
}

// Test validation helper functions
func (suite *ContactServiceTestSuite) TestValidateContact() {
	// Test valid contact
	validContact := &models.Contact{
		Name:  "John Doe",
		Email: "john@example.com",
	}
	assert.NoError(suite.T(), suite.service.validateContact(validContact))

	// Test invalid contact - empty name
	invalidContact1 := &models.Contact{
		Name:  "",
		Email: "john@example.com",
	}
	assert.Error(suite.T(), suite.service.validateContact(invalidContact1))

	// Test invalid contact - empty email
	invalidContact2 := &models.Contact{
		Name:  "John Doe",
		Email: "",
	}
	assert.Error(suite.T(), suite.service.validateContact(invalidContact2))

	// Test invalid contact - invalid email format
	invalidContact3 := &models.Contact{
		Name:  "John Doe",
		Email: "invalid-email",
	}
	assert.Error(suite.T(), suite.service.validateContact(invalidContact3))
}

func (suite *ContactServiceTestSuite) TestIsValidStatus() {
	validStatuses := []string{"new", "contacted", "qualified", "customer", "inactive"}
	for _, status := range validStatuses {
		assert.True(suite.T(), suite.service.isValidStatus(status))
	}

	invalidStatuses := []string{"", "invalid", "pending", "closed"}
	for _, status := range invalidStatuses {
		assert.False(suite.T(), suite.service.isValidStatus(status))
	}
}

// Run the test suite
func TestContactServiceSuite(t *testing.T) {
	suite.Run(t, new(ContactServiceTestSuite))
}

// Benchmark tests
func BenchmarkCreateContact(b *testing.B) {
	mockRepo := new(MockContactRepository)
	service := &ContactService{repo: mockRepo}

	contact := &models.Contact{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	// Setup mocks for benchmark
	mockRepo.On("GetByEmail", mock.AnythingOfType("string")).Return(nil, ErrContactNotFound)
	mockRepo.On("Create", mock.AnythingOfType("*models.Contact")).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newContact := *contact
		newContact.Email = "john" + string(rune(i)) + "@example.com"
		_ = service.CreateContact(&newContact)
	}
}

func BenchmarkGetContact(b *testing.B) {
	mockRepo := new(MockContactRepository)
	service := &ContactService{repo: mockRepo}

	contact := &models.Contact{
		ID:    1,
		Name:  "John Doe",
		Email: "john@example.com",
	}

	mockRepo.On("GetByID", mock.AnythingOfType("uint")).Return(contact, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GetContact(1)
	}
}