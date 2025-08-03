package repository

import (
	"fmt"

	"gorm.io/gorm"

	"contact-service/internal/models"
)

// contactRepository implements ContactRepository interface
type contactRepository struct {
	db *gorm.DB
}

// NewContactRepository creates a new contact repository
func NewContactRepository(db *gorm.DB) ContactRepository {
	return &contactRepository{db: db}
}

// Create creates a new contact
func (r *contactRepository) Create(contact *models.Contact) error {
	return r.db.Create(contact).Error
}

// GetByID retrieves a contact by ID
func (r *contactRepository) GetByID(id uint) (*models.Contact, error) {
	var contact models.Contact
	err := r.db.First(&contact, id).Error
	if err != nil {
		return nil, err
	}
	return &contact, nil
}

// GetByEmail retrieves a contact by email
func (r *contactRepository) GetByEmail(email string) (*models.Contact, error) {
	var contact models.Contact
	err := r.db.Where("email = ?", email).First(&contact).Error
	if err != nil {
		return nil, err
	}
	return &contact, nil
}

// Update updates a contact
func (r *contactRepository) Update(contact *models.Contact) error {
	return r.db.Save(contact).Error
}

// Delete deletes a contact
func (r *contactRepository) Delete(id uint) error {
	return r.db.Delete(&models.Contact{}, id).Error
}

// List retrieves contacts with pagination and filtering
func (r *contactRepository) List(params ContactListParams) ([]models.Contact, int64, error) {
	var contacts []models.Contact
	var total int64

	query := r.db.Model(&models.Contact{})

	// Apply filters
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.TypeID != 0 {
		query = query.Where("type_id = ?", params.TypeID)
	}
	if params.SourceID != 0 {
		query = query.Where("source_id = ?", params.SourceID)
	}
	if params.Search != "" {
		query = query.Where("name LIKE ? OR email LIKE ? OR company LIKE ?",
			"%"+params.Search+"%", "%"+params.Search+"%", "%"+params.Search+"%")
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	if params.Sort != "" {
		order := "DESC"
		if params.Order == "asc" {
			order = "ASC"
		}
		query = query.Order(fmt.Sprintf("%s %s", params.Sort, order))
	} else {
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	if params.Limit > 0 {
		offset := (params.Page - 1) * params.Limit
		query = query.Offset(offset).Limit(params.Limit)
	}

	// Execute query
	err := query.Find(&contacts).Error
	return contacts, total, err
}

// Search searches contacts with a query string
func (r *contactRepository) Search(searchQuery string, params ContactListParams) ([]models.Contact, int64, error) {
	params.Search = searchQuery
	return r.List(params)
}

// UpdateStatus updates a contact's status
func (r *contactRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&models.Contact{}).Where("id = ?", id).Update("status", status).Error
}

// Assign assigns a contact to a user
func (r *contactRepository) Assign(id uint, userID uint) error {
	return r.db.Model(&models.Contact{}).Where("id = ?", id).Update("assigned_to", userID).Error
}

// GetAssignedContacts retrieves all contacts assigned to a user
func (r *contactRepository) GetAssignedContacts(userID uint) ([]models.Contact, error) {
	var contacts []models.Contact
	err := r.db.Where("assigned_to = ?", userID).Find(&contacts).Error
	return contacts, err
}