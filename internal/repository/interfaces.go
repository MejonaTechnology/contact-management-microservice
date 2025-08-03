package repository

import "contact-service/internal/models"

// ContactRepository defines the interface for contact data operations
type ContactRepository interface {
	Create(contact *models.Contact) error
	GetByID(id uint) (*models.Contact, error)
	GetByEmail(email string) (*models.Contact, error)
	Update(contact *models.Contact) error
	Delete(id uint) error
	List(params ContactListParams) ([]models.Contact, int64, error)
	Search(query string, params ContactListParams) ([]models.Contact, int64, error)
	UpdateStatus(id uint, status string) error
	Assign(id uint, userID uint) error
	GetAssignedContacts(userID uint) ([]models.Contact, error)
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	GetByID(id uint) (*models.AdminUser, error)
	GetByEmail(email string) (*models.AdminUser, error)
	Create(user *models.AdminUser) error
	Update(user *models.AdminUser) error
	Delete(id uint) error
	List(params UserListParams) ([]models.AdminUser, int64, error)
}

// ContactListParams represents parameters for listing contacts
type ContactListParams struct {
	Page     int
	Limit    int
	Sort     string
	Order    string
	Status   string
	TypeID   uint
	SourceID uint
	Search   string
}

// UserListParams represents parameters for listing users
type UserListParams struct {
	Page   int
	Limit  int
	Sort   string
	Order  string
	Role   string
	Active bool
}