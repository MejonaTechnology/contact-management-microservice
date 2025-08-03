package repository

import (
	"fmt"

	"gorm.io/gorm"

	"contact-service/internal/models"
)

// userRepository implements UserRepository interface
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(id uint) (*models.AdminUser, error) {
	var user models.AdminUser
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(email string) (*models.AdminUser, error) {
	var user models.AdminUser
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create creates a new user
func (r *userRepository) Create(user *models.AdminUser) error {
	return r.db.Create(user).Error
}

// Update updates a user
func (r *userRepository) Update(user *models.AdminUser) error {
	return r.db.Save(user).Error
}

// Delete deletes a user
func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&models.AdminUser{}, id).Error
}

// List retrieves users with pagination and filtering
func (r *userRepository) List(params UserListParams) ([]models.AdminUser, int64, error) {
	var users []models.AdminUser
	var total int64

	query := r.db.Model(&models.AdminUser{})

	// Apply filters
	if params.Role != "" {
		query = query.Where("role = ?", params.Role)
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
		query = query.Order("name ASC")
	}

	// Apply pagination
	if params.Limit > 0 {
		offset := (params.Page - 1) * params.Limit
		query = query.Offset(offset).Limit(params.Limit)
	}

	// Execute query
	err := query.Find(&users).Error
	return users, total, err
}