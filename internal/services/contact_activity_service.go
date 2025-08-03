package services

import (
	"contact-service/internal/models"
	"contact-service/pkg/database"
	"contact-service/pkg/logger"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ContactActivityService handles business logic for contact activity management
type ContactActivityService struct {
	db *gorm.DB
}

// NewContactActivityService creates a new contact activity service instance
func NewContactActivityService() *ContactActivityService {
	return &ContactActivityService{
		db: database.DB,
	}
}

// ActivityListOptions represents options for listing activities
type ActivityListOptions struct {
	Page         int
	PageSize     int
	ContactID    *uint
	ActivityType string
	Status       string
	PerformedBy  *uint
	AssignedTo   *uint
	DateFrom     *time.Time
	DateTo       *time.Time
	SortBy       string
	SortOrder    string
}

// CreateActivity creates a new contact activity
func (s *ContactActivityService) CreateActivity(req *models.ContactActivityRequest, performedBy uint) (*models.ContactActivity, error) {
	// Validate contact exists
	var contact models.Contact
	if err := s.db.First(&contact, req.ContactID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("contact not found")
		}
		return nil, fmt.Errorf("failed to validate contact: %v", err)
	}

	// Create activity
	activity := &models.ContactActivity{
		ContactID:    req.ContactID,
		ActivityType: req.ActivityType,
		Title:        req.Title,
		Description:  req.Description,
		Outcome:      req.Outcome,
		PerformedBy:  performedBy,
		Status:       models.ActivityStatusCompleted,
		Priority:     models.PriorityMedium,
		Direction:    models.DirectionOutbound,
		Channel:      models.ChannelEmail,
		Tags:         req.Tags,
		Metadata:     req.Metadata,
		Attachments:  req.Attachments,
		CreatedBy:    &performedBy,
	}

	// Set activity date
	if req.ActivityDate != nil {
		activity.ActivityDate = *req.ActivityDate
	} else {
		activity.ActivityDate = time.Now()
	}

	// Set optional fields
	if req.DurationMinutes != nil {
		activity.DurationMinutes = *req.DurationMinutes
	}
	if req.Status != nil {
		activity.Status = *req.Status
	}
	if req.Priority != nil {
		activity.Priority = *req.Priority
	}
	if req.Direction != nil {
		activity.Direction = *req.Direction
	}
	if req.Channel != nil {
		activity.Channel = *req.Channel
	}
	if req.ScheduledDate != nil {
		activity.ScheduledDate = req.ScheduledDate
	}
	if req.ReminderDate != nil {
		activity.ReminderDate = req.ReminderDate
	}
	if req.AssignedTo != nil {
		activity.AssignedTo = req.AssignedTo
	}
	if req.IsBillable != nil {
		activity.IsBillable = *req.IsBillable
	}
	if req.BillableAmount != nil {
		activity.BillableAmount = *req.BillableAmount
	}
	if req.Cost != nil {
		activity.Cost = *req.Cost
	}

	// Save to database
	if err := s.db.Create(activity).Error; err != nil {
		logger.Error("Failed to create contact activity", err, map[string]interface{}{
			"contact_id":    req.ContactID,
			"activity_type": req.ActivityType,
		})
		return nil, fmt.Errorf("failed to create activity: %v", err)
	}

	// Update contact's last activity date and interaction count
	if err := s.updateContactLastActivity(req.ContactID); err != nil {
		logger.Warn("Failed to update contact last activity", map[string]interface{}{
			"contact_id":  req.ContactID,
			"activity_id": activity.ID,
			"error":       err.Error(),
		})
	}

	logger.LogContactActivity(req.ContactID, string(req.ActivityType), map[string]interface{}{
		"activity_id":  activity.ID,
		"title":        req.Title,
		"performed_by": performedBy,
		"status":       activity.Status,
	})

	return activity, nil
}

// GetActivity retrieves an activity by ID
func (s *ContactActivityService) GetActivity(id uint) (*models.ContactActivity, error) {
	var activity models.ContactActivity
	
	if err := s.db.Preload("Contact").First(&activity, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("activity not found")
		}
		return nil, fmt.Errorf("failed to get activity: %v", err)
	}

	return &activity, nil
}

// UpdateActivity updates an existing activity
func (s *ContactActivityService) UpdateActivity(id uint, req *models.ContactActivityRequest, updatedBy uint) (*models.ContactActivity, error) {
	// Get existing activity
	activity, err := s.GetActivity(id)
	if err != nil {
		return nil, err
	}

	// Store original values
	originalStatus := activity.Status

	// Update fields
	activity.Title = req.Title
	activity.Description = req.Description
	activity.Outcome = req.Outcome
	activity.Tags = req.Tags
	activity.Metadata = req.Metadata
	activity.Attachments = req.Attachments
	activity.UpdatedBy = &updatedBy

	// Set activity date
	if req.ActivityDate != nil {
		activity.ActivityDate = *req.ActivityDate
	}

	// Set optional fields
	if req.DurationMinutes != nil {
		activity.DurationMinutes = *req.DurationMinutes
	}
	if req.Status != nil {
		activity.Status = *req.Status
	}
	if req.Priority != nil {
		activity.Priority = *req.Priority
	}
	if req.Direction != nil {
		activity.Direction = *req.Direction
	}
	if req.Channel != nil {
		activity.Channel = *req.Channel
	}
	if req.ScheduledDate != nil {
		activity.ScheduledDate = req.ScheduledDate
	}
	if req.ReminderDate != nil {
		activity.ReminderDate = req.ReminderDate
	}
	if req.AssignedTo != nil {
		activity.AssignedTo = req.AssignedTo
	}
	if req.IsBillable != nil {
		activity.IsBillable = *req.IsBillable
	}
	if req.BillableAmount != nil {
		activity.BillableAmount = *req.BillableAmount
	}
	if req.Cost != nil {
		activity.Cost = *req.Cost
	}

	// Set completion date if status changed to completed
	if originalStatus != models.ActivityStatusCompleted && activity.Status == models.ActivityStatusCompleted {
		now := time.Now()
		activity.CompletedDate = &now
	}

	// Save to database
	if err := s.db.Save(activity).Error; err != nil {
		logger.Error("Failed to update contact activity", err, map[string]interface{}{
			"activity_id": id,
		})
		return nil, fmt.Errorf("failed to update activity: %v", err)
	}

	// Log status change if applicable
	if originalStatus != activity.Status {
		logger.LogContactActivity(activity.ContactID, "activity_status_changed", map[string]interface{}{
			"activity_id": id,
			"old_status":  originalStatus,
			"new_status":  activity.Status,
			"updated_by":  updatedBy,
		})
	}

	return activity, nil
}

// DeleteActivity soft deletes an activity
func (s *ContactActivityService) DeleteActivity(id uint, deletedBy uint) error {
	activity, err := s.GetActivity(id)
	if err != nil {
		return err
	}

	now := time.Now()
	activity.DeletedAt = &now
	activity.UpdatedBy = &deletedBy

	if err := s.db.Save(activity).Error; err != nil {
		logger.Error("Failed to delete contact activity", err, map[string]interface{}{
			"activity_id": id,
		})
		return fmt.Errorf("failed to delete activity: %v", err)
	}

	logger.LogContactActivity(activity.ContactID, "activity_deleted", map[string]interface{}{
		"activity_id": id,
		"title":       activity.Title,
		"deleted_by":  deletedBy,
	})

	return nil
}

// ListActivities retrieves activities with filtering and pagination
func (s *ContactActivityService) ListActivities(opts *ActivityListOptions) ([]*models.ContactActivity, int64, error) {
	query := s.db.Model(&models.ContactActivity{}).
		Preload("Contact").
		Where("deleted_at IS NULL")

	// Apply filters
	if opts.ContactID != nil {
		query = query.Where("contact_id = ?", *opts.ContactID)
	}
	if opts.ActivityType != "" {
		query = query.Where("activity_type = ?", opts.ActivityType)
	}
	if opts.Status != "" {
		query = query.Where("status = ?", opts.Status)
	}
	if opts.PerformedBy != nil {
		query = query.Where("performed_by = ?", *opts.PerformedBy)
	}
	if opts.AssignedTo != nil {
		query = query.Where("assigned_to = ?", *opts.AssignedTo)
	}
	if opts.DateFrom != nil {
		query = query.Where("activity_date >= ?", *opts.DateFrom)
	}
	if opts.DateTo != nil {
		query = query.Where("activity_date <= ?", *opts.DateTo)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count activities: %v", err)
	}

	// Apply sorting
	sortBy := "activity_date"
	sortOrder := "DESC"
	if opts.SortBy != "" {
		sortBy = opts.SortBy
	}
	if opts.SortOrder != "" {
		sortOrder = opts.SortOrder
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.PageSize < 1 || opts.PageSize > 100 {
		opts.PageSize = 10
	}
	offset := (opts.Page - 1) * opts.PageSize
	query = query.Offset(offset).Limit(opts.PageSize)

	// Execute query
	var activities []*models.ContactActivity
	if err := query.Find(&activities).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list activities: %v", err)
	}

	return activities, total, nil
}

// GetContactActivities retrieves all activities for a specific contact
func (s *ContactActivityService) GetContactActivities(contactID uint, limit int) ([]*models.ContactActivity, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var activities []*models.ContactActivity
	if err := s.db.Where("contact_id = ? AND deleted_at IS NULL", contactID).
		Order("activity_date DESC").
		Limit(limit).
		Find(&activities).Error; err != nil {
		return nil, fmt.Errorf("failed to get contact activities: %v", err)
	}

	return activities, nil
}

// GetUpcomingActivities retrieves activities scheduled for the future
func (s *ContactActivityService) GetUpcomingActivities(userID *uint, limit int) ([]*models.ContactActivity, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := s.db.Model(&models.ContactActivity{}).
		Preload("Contact").
		Where("deleted_at IS NULL").
		Where("status IN (?)", []string{string(models.ActivityStatusPending), string(models.ActivityStatusInProgress)}).
		Where("scheduled_date IS NOT NULL AND scheduled_date > NOW()")

	if userID != nil {
		query = query.Where("assigned_to = ? OR performed_by = ?", *userID, *userID)
	}

	var activities []*models.ContactActivity
	if err := query.Order("scheduled_date ASC").
		Limit(limit).
		Find(&activities).Error; err != nil {
		return nil, fmt.Errorf("failed to get upcoming activities: %v", err)
	}

	return activities, nil
}

// GetOverdueActivities retrieves activities that are overdue
func (s *ContactActivityService) GetOverdueActivities(userID *uint, limit int) ([]*models.ContactActivity, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := s.db.Model(&models.ContactActivity{}).
		Preload("Contact").
		Where("deleted_at IS NULL").
		Where("status IN (?)", []string{string(models.ActivityStatusPending), string(models.ActivityStatusInProgress)}).
		Where("scheduled_date IS NOT NULL AND scheduled_date < NOW()")

	if userID != nil {
		query = query.Where("assigned_to = ? OR performed_by = ?", *userID, *userID)
	}

	var activities []*models.ContactActivity
	if err := query.Order("scheduled_date ASC").
		Limit(limit).
		Find(&activities).Error; err != nil {
		return nil, fmt.Errorf("failed to get overdue activities: %v", err)
	}

	return activities, nil
}

// GetActivityStats returns activity statistics
func (s *ContactActivityService) GetActivityStats(userID *uint, dateFrom, dateTo *time.Time) (map[string]interface{}, error) {
	query := s.db.Model(&models.ContactActivity{}).Where("deleted_at IS NULL")
	
	if userID != nil {
		query = query.Where("performed_by = ?", *userID)
	}
	if dateFrom != nil {
		query = query.Where("activity_date >= ?", *dateFrom)
	}
	if dateTo != nil {
		query = query.Where("activity_date <= ?", *dateTo)
	}

	// Get activity type distribution
	var typeStats []struct {
		ActivityType string `json:"activity_type"`
		Count        int64  `json:"count"`
	}
	
	if err := query.Select("activity_type, COUNT(*) as count").
		Group("activity_type").
		Find(&typeStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get activity type stats: %v", err)
	}

	// Get status distribution
	var statusStats []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	
	if err := query.Select("status, COUNT(*) as count").
		Group("status").
		Find(&statusStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get activity status stats: %v", err)
	}

	// Get total counts
	var totalActivities int64
	if err := query.Count(&totalActivities).Error; err != nil {
		return nil, fmt.Errorf("failed to get total activities: %v", err)
	}

	var completedActivities int64
	if err := query.Where("status = ?", models.ActivityStatusCompleted).
		Count(&completedActivities).Error; err != nil {
		return nil, fmt.Errorf("failed to get completed activities: %v", err)
	}

	stats := map[string]interface{}{
		"total_activities":     totalActivities,
		"completed_activities": completedActivities,
		"completion_rate":      0.0,
		"activity_types":       typeStats,
		"status_distribution":  statusStats,
	}

	if totalActivities > 0 {
		stats["completion_rate"] = (float64(completedActivities) / float64(totalActivities)) * 100
	}

	return stats, nil
}

// ScheduleFollowUp schedules a follow-up activity for a contact
func (s *ContactActivityService) ScheduleFollowUp(contactID uint, followUpDate time.Time, title, description string, assignedTo, scheduledBy uint) (*models.ContactActivity, error) {
	activity := &models.ContactActivity{
		ContactID:     contactID,
		ActivityType:  models.ActivityFollowUp,
		Title:         title,
		Description:   &description,
		Status:        models.ActivityStatusPending,
		Priority:      models.PriorityMedium,
		Direction:     models.DirectionOutbound,
		Channel:       models.ChannelEmail,
		ScheduledDate: &followUpDate,
		PerformedBy:   scheduledBy,
		AssignedTo:    &assignedTo,
		CreatedBy:     &scheduledBy,
	}

	if err := s.db.Create(activity).Error; err != nil {
		return nil, fmt.Errorf("failed to schedule follow-up: %v", err)
	}

	logger.LogContactActivity(contactID, "followup_scheduled", map[string]interface{}{
		"activity_id":    activity.ID,
		"scheduled_date": followUpDate,
		"assigned_to":    assignedTo,
		"scheduled_by":   scheduledBy,
	})

	return activity, nil
}

// Helper methods

func (s *ContactActivityService) updateContactLastActivity(contactID uint) error {
	now := time.Now()
	return s.db.Model(&models.Contact{}).
		Where("id = ?", contactID).
		Updates(map[string]interface{}{
			"last_activity_date":  now,
			"total_interactions": gorm.Expr("total_interactions + 1"),
		}).Error
}