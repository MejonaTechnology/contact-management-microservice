package services

import (
	"contact-service/internal/models"
	"contact-service/pkg/logger"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// AssignmentService handles contact assignment and routing logic
type AssignmentService struct {
	db *gorm.DB
}

// NewAssignmentService creates a new assignment service
func NewAssignmentService(db *gorm.DB) *AssignmentService {
	return &AssignmentService{db: db}
}

// AssignContactAutomatically assigns a contact automatically based on rules
func (s *AssignmentService) AssignContactAutomatically(contactID uint, contextData map[string]interface{}) (*models.ContactAssignment, error) {
	// Get the contact
	var contact models.Contact
	if err := s.db.Preload("ContactType").Preload("ContactSource").First(&contact, contactID).Error; err != nil {
		return nil, fmt.Errorf("contact not found: %v", err)
	}

	// Get active assignment rules ordered by priority
	var rules []models.AssignmentRule
	if err := s.db.Where("status = ? AND deleted_at IS NULL", models.AssignmentRuleActive).
		Order("priority DESC, created_at ASC").Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("failed to get assignment rules: %v", err)
	}

	// Try each rule until one matches
	for _, rule := range rules {
		if s.evaluateRule(&rule, &contact, contextData) {
			assigneeID, err := s.selectAssignee(&rule, &contact)
			if err != nil {
				logger.Error("Failed to select assignee for rule", err, map[string]interface{}{
					"rule_id":    rule.ID,
					"contact_id": contactID,
				})
				continue
			}

			// Create assignment
			assignment := &models.ContactAssignment{
				ContactID:        contactID,
				AssignedToID:     assigneeID,
				RuleID:           &rule.ID,
				AssignmentType:   "automatic",
				AssignmentReason: s.generateAssignmentReason(&rule, &contact),
				Priority:         contact.Priority,
				Status:           "active",
			}

			if err := s.db.Create(assignment).Error; err != nil {
				return nil, fmt.Errorf("failed to create assignment: %v", err)
			}

			// Update contact assignment fields
			now := time.Now()
			if err := s.db.Model(&contact).Updates(map[string]interface{}{
				"assigned_to": assigneeID,
				"assigned_at": now,
			}).Error; err != nil {
				logger.Error("Failed to update contact assignment", err, map[string]interface{}{
					"contact_id":      contactID,
					"assigned_to_id":  assigneeID,
				})
			}

			// Update rule statistics
			s.updateRuleStatistics(&rule)

			// Update user workload
			s.updateUserWorkload(assigneeID)

			// Log assignment history
			s.logAssignmentHistory(contactID, nil, &assigneeID, nil, &rule.ID, "assigned", 
				fmt.Sprintf("Automatically assigned by rule: %s", rule.Name))

			logger.Info("Contact assigned automatically", map[string]interface{}{
				"contact_id":      contactID,
				"assigned_to_id":  assigneeID,
				"rule_id":         rule.ID,
				"rule_name":       rule.Name,
			})

			return assignment, nil
		}
	}

	// No rule matched, use fallback assignment
	return s.fallbackAssignment(&contact)
}

// AssignContactManually assigns a contact manually to a specific user
func (s *AssignmentService) AssignContactManually(request *models.ContactAssignmentRequest, assignedByID uint) (*models.ContactAssignment, error) {
	// Validate assignee exists and is active
	var assignee models.AdminUser
	if err := s.db.Where("id = ? AND is_active = ?", request.AssignedToID, true).First(&assignee).Error; err != nil {
		return nil, fmt.Errorf("assignee not found or inactive: %v", err)
	}

	// Get current assignment if exists
	var currentAssignment models.ContactAssignment
	currentExists := s.db.Where("contact_id = ? AND status = ?", request.ContactID, "active").
		First(&currentAssignment).Error == nil

	var fromUserID *uint
	if currentExists {
		fromUserID = &currentAssignment.AssignedToID
		// Mark current assignment as reassigned
		s.db.Model(&currentAssignment).Updates(map[string]interface{}{
			"status": "reassigned",
		})
	}

	// Create new assignment
	assignment := &models.ContactAssignment{
		ContactID:        request.ContactID,
		AssignedToID:     request.AssignedToID,
		AssignedByID:     &assignedByID,
		AssignmentType:   "manual",
		AssignmentReason: request.AssignmentReason,
		Priority:         request.Priority,
		Status:           "active",
	}

	if err := s.db.Create(assignment).Error; err != nil {
		return nil, fmt.Errorf("failed to create manual assignment: %v", err)
	}

	// Update contact assignment fields
	now := time.Now()
	if err := s.db.Model(&models.Contact{}).Where("id = ?", request.ContactID).Updates(map[string]interface{}{
		"assigned_to": request.AssignedToID,
		"assigned_at": now,
	}).Error; err != nil {
		logger.Error("Failed to update contact assignment", err, map[string]interface{}{
			"contact_id":     request.ContactID,
			"assigned_to_id": request.AssignedToID,
		})
	}

	// Update user workloads
	if fromUserID != nil {
		s.updateUserWorkload(*fromUserID)
	}
	s.updateUserWorkload(request.AssignedToID)

	// Log assignment history
	changeType := "assigned"
	if currentExists {
		changeType = "reassigned"
	}
	s.logAssignmentHistory(request.ContactID, fromUserID, &request.AssignedToID, &assignedByID, nil, 
		changeType, request.AssignmentReason)

	logger.Info("Contact assigned manually", map[string]interface{}{
		"contact_id":      request.ContactID,
		"assigned_to_id":  request.AssignedToID,
		"assigned_by_id":  assignedByID,
		"from_user_id":    fromUserID,
	})

	return assignment, nil
}

// BulkAssignContacts assigns multiple contacts to a user
func (s *AssignmentService) BulkAssignContacts(request *models.BulkAssignmentRequest, assignedByID uint) error {
	// Validate assignee exists and is active
	var assignee models.AdminUser
	if err := s.db.Where("id = ? AND is_active = ?", request.AssignedToID, true).First(&assignee).Error; err != nil {
		return fmt.Errorf("assignee not found or inactive: %v", err)
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	successCount := 0
	var lastError error

	for _, contactID := range request.ContactIDs {
		// Create assignment request for each contact
		assignmentRequest := &models.ContactAssignmentRequest{
			ContactID:        contactID,
			AssignedToID:     request.AssignedToID,
			AssignmentReason: request.AssignmentReason,
			Priority:         request.Priority,
		}

		// Use a new service instance with the transaction
		txService := &AssignmentService{db: tx}
		_, err := txService.AssignContactManually(assignmentRequest, assignedByID)
		if err != nil {
			logger.Error("Failed to assign contact in bulk operation", err, map[string]interface{}{
				"contact_id":     contactID,
				"assigned_to_id": request.AssignedToID,
			})
			lastError = err
			continue
		}
		successCount++
	}

	if successCount == 0 {
		tx.Rollback()
		return fmt.Errorf("failed to assign any contacts: %v", lastError)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit bulk assignment: %v", err)
	}

	logger.Info("Bulk contact assignment completed", map[string]interface{}{
		"total_contacts":   len(request.ContactIDs),
		"successful":       successCount,
		"failed":          len(request.ContactIDs) - successCount,
		"assigned_to_id":  request.AssignedToID,
		"assigned_by_id":  assignedByID,
	})

	return nil
}

// UnassignContact removes assignment from a contact
func (s *AssignmentService) UnassignContact(contactID uint, unassignedByID uint, reason string) error {
	// Get current assignment
	var assignment models.ContactAssignment
	if err := s.db.Where("contact_id = ? AND status = ?", contactID, "active").First(&assignment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("no active assignment found for contact")
		}
		return fmt.Errorf("failed to get current assignment: %v", err)
	}

	// Mark assignment as unassigned
	if err := s.db.Model(&assignment).Updates(map[string]interface{}{
		"status": "unassigned",
	}).Error; err != nil {
		return fmt.Errorf("failed to update assignment status: %v", err)
	}

	// Update contact to remove assignment
	if err := s.db.Model(&models.Contact{}).Where("id = ?", contactID).Updates(map[string]interface{}{
		"assigned_to": nil,
		"assigned_at": nil,
	}).Error; err != nil {
		logger.Error("Failed to update contact assignment", err, map[string]interface{}{
			"contact_id": contactID,
		})
	}

	// Update user workload
	s.updateUserWorkload(assignment.AssignedToID)

	// Log assignment history
	s.logAssignmentHistory(contactID, &assignment.AssignedToID, nil, &unassignedByID, nil, 
		"unassigned", reason)

	logger.Info("Contact unassigned", map[string]interface{}{
		"contact_id":       contactID,
		"from_user_id":     assignment.AssignedToID,
		"unassigned_by_id": unassignedByID,
	})

	return nil
}

// GetUserWorkload gets the current workload for a user
func (s *AssignmentService) GetUserWorkload(userID uint) (*models.UserWorkloadResponse, error) {
	var workload models.UserWorkload
	if err := s.db.Preload("User").Where("user_id = ?", userID).First(&workload).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create default workload
			workload = models.UserWorkload{
				UserID:       userID,
				IsAvailable:  true,
			}
			s.db.Create(&workload)
			s.db.Preload("User").Where("user_id = ?", userID).First(&workload)
		} else {
			return nil, fmt.Errorf("failed to get user workload: %v", err)
		}
	}

	response := &models.UserWorkloadResponse{
		UserID:                 workload.UserID,
		ActiveContacts:         workload.ActiveContacts,
		PendingContacts:        workload.PendingContacts,
		OverdueContacts:        workload.OverdueContacts,
		TotalContacts:          workload.TotalContacts,
		TodayAssignments:       workload.TodayAssignments,
		TodayResponses:         workload.TodayResponses,
		TodayCompletions:       workload.TodayCompletions,
		AvgResponseTimeHours:   workload.AvgResponseTimeHours,
		AvgResolutionTimeHours: workload.AvgResolutionTimeHours,
		ConversionRate:         workload.ConversionRate,
		IsAvailable:            workload.IsAvailable,
		MaxDailyAssignments:    workload.MaxDailyAssignments,
		MaxActiveContacts:      workload.MaxActiveContacts,
		Skills:                 workload.Skills,
		Territories:            workload.Territories,
		ContactTypes:           workload.ContactTypes,
		LastCalculatedAt:       workload.LastCalculatedAt,
		WorkloadScore:          s.calculateWorkloadScore(&workload),
		AvailabilityScore:      s.calculateAvailabilityScore(&workload),
	}

	if workload.User != nil {
		response.User = workload.User.ToResponse()
	}

	return response, nil
}

// Private helper methods

// evaluateRule checks if a contact matches the conditions of an assignment rule
func (s *AssignmentService) evaluateRule(rule *models.AssignmentRule, contact *models.Contact, contextData map[string]interface{}) bool {
	// Check business hours if enabled
	if rule.BusinessHoursEnabled {
		if !s.isWithinBusinessHours(rule) {
			return false
		}
	}

	// Check rate limits
	if !s.checkRateLimits(rule) {
		return false
	}

	// Evaluate conditions
	for _, condition := range rule.Conditions {
		if !s.evaluateCondition(condition, contact, contextData) {
			return false
		}
	}

	return true
}

// evaluateCondition evaluates a single condition against a contact
func (s *AssignmentService) evaluateCondition(condition models.AssignmentCondition, contact *models.Contact, contextData map[string]interface{}) bool {
	fieldValue := s.getFieldValue(condition.Field, contact, contextData)
	
	switch condition.Operator {
	case "equals":
		return s.compareValues(fieldValue, condition.Value, "equals")
	case "not_equals":
		return !s.compareValues(fieldValue, condition.Value, "equals")
	case "contains":
		return s.compareValues(fieldValue, condition.Value, "contains")
	case "not_contains":
		return !s.compareValues(fieldValue, condition.Value, "contains")
	case "greater_than":
		return s.compareValues(fieldValue, condition.Value, "greater_than")
	case "less_than":
		return s.compareValues(fieldValue, condition.Value, "less_than")
	case "in":
		return s.compareValues(fieldValue, condition.Value, "in")
	case "not_in":
		return !s.compareValues(fieldValue, condition.Value, "in")
	default:
		return false
	}
}

// getFieldValue extracts field value from contact or context data
func (s *AssignmentService) getFieldValue(fieldName string, contact *models.Contact, contextData map[string]interface{}) interface{} {
	// Check context data first
	if val, exists := contextData[fieldName]; exists {
		return val
	}

	// Map contact fields
	switch fieldName {
	case "contact_type_id":
		return contact.ContactTypeID
	case "contact_source_id":
		return contact.ContactSourceID
	case "priority":
		return contact.Priority
	case "lead_score":
		return contact.LeadScore
	case "estimated_value":
		return contact.EstimatedValue
	case "country":
		return contact.Country
	case "state":
		return contact.State
	case "city":
		return contact.City
	case "company":
		return contact.Company
	case "utm_source":
		return contact.UTMSource
	case "utm_medium":
		return contact.UTMMedium
	case "utm_campaign":
		return contact.UTMCampaign
	default:
		return nil
	}
}

// compareValues compares two values based on operator
func (s *AssignmentService) compareValues(fieldValue, conditionValue interface{}, operator string) bool {
	switch operator {
	case "equals":
		return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", conditionValue)
	case "contains":
		fieldStr := strings.ToLower(fmt.Sprintf("%v", fieldValue))
		conditionStr := strings.ToLower(fmt.Sprintf("%v", conditionValue))
		return strings.Contains(fieldStr, conditionStr)
	case "greater_than":
		fieldNum := s.toFloat64(fieldValue)
		conditionNum := s.toFloat64(conditionValue)
		return fieldNum > conditionNum
	case "less_than":
		fieldNum := s.toFloat64(fieldValue)
		conditionNum := s.toFloat64(conditionValue)
		return fieldNum < conditionNum
	case "in":
		conditionArray, ok := conditionValue.([]interface{})
		if !ok {
			return false
		}
		fieldStr := fmt.Sprintf("%v", fieldValue)
		for _, val := range conditionArray {
			if fmt.Sprintf("%v", val) == fieldStr {
				return true
			}
		}
		return false
	}
	return false
}

// toFloat64 converts interface{} to float64
func (s *AssignmentService) toFloat64(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0
}

// selectAssignee selects the best assignee based on rule type
func (s *AssignmentService) selectAssignee(rule *models.AssignmentRule, contact *models.Contact) (uint, error) {
	assigneeIDs := s.extractUserIDs(rule.AssigneeIDs)
	if len(assigneeIDs) == 0 {
		if rule.FallbackUserID != nil {
			return *rule.FallbackUserID, nil
		}
		return 0, errors.New("no assignees available")
	}

	switch rule.Type {
	case models.AssignmentRuleRoundRobin:
		return s.selectRoundRobin(assigneeIDs, rule)
	case models.AssignmentRuleLoadBased:
		return s.selectLoadBased(assigneeIDs)
	case models.AssignmentRuleSkillBased:
		return s.selectSkillBased(assigneeIDs, contact)
	case models.AssignmentRuleGeographyBased:
		return s.selectGeographyBased(assigneeIDs, contact)
	case models.AssignmentRuleValueBased:
		return s.selectValueBased(assigneeIDs, contact)
	default:
		// Default to round robin
		return s.selectRoundRobin(assigneeIDs, rule)
	}
}

// selectRoundRobin selects assignee using round-robin algorithm
func (s *AssignmentService) selectRoundRobin(assigneeIDs []uint, rule *models.AssignmentRule) (uint, error) {
	if len(assigneeIDs) == 0 {
		return 0, errors.New("no assignees available")
	}

	// Get last assignment from this rule
	var lastAssignment models.ContactAssignment
	if s.db.Where("rule_id = ?", rule.ID).Order("created_at DESC").First(&lastAssignment).Error == nil {
		// Find current assignee index
		for i, id := range assigneeIDs {
			if id == lastAssignment.AssignedToID {
				// Return next assignee (circular)
				nextIndex := (i + 1) % len(assigneeIDs)
				return assigneeIDs[nextIndex], nil
			}
		}
	}

	// No previous assignment or assignee not found, return first
	return assigneeIDs[0], nil
}

// selectLoadBased selects assignee with lowest workload
func (s *AssignmentService) selectLoadBased(assigneeIDs []uint) (uint, error) {
	type workloadScore struct {
		UserID uint
		Score  float64
	}

	var scores []workloadScore
	
	for _, userID := range assigneeIDs {
		var workload models.UserWorkload
		if s.db.Where("user_id = ?", userID).First(&workload).Error != nil {
			// User not found in workload, create default
			workload = models.UserWorkload{
				UserID:      userID,
				IsAvailable: true,
			}
		}

		if !workload.IsAvailable {
			continue
		}

		// Calculate workload score (lower is better)
		score := s.calculateWorkloadScore(&workload)
		scores = append(scores, workloadScore{UserID: userID, Score: score})
	}

	if len(scores) == 0 {
		if len(assigneeIDs) > 0 {
			return assigneeIDs[0], nil // Fallback to first assignee
		}
		return 0, errors.New("no available assignees")
	}

	// Sort by score (ascending - lower is better)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score < scores[j].Score
	})

	return scores[0].UserID, nil
}

// selectSkillBased selects assignee based on skills matching
func (s *AssignmentService) selectSkillBased(assigneeIDs []uint, contact *models.Contact) (uint, error) {
	// For now, use load-based as fallback
	// This could be enhanced to match skills from contact type or custom fields
	return s.selectLoadBased(assigneeIDs)
}

// selectGeographyBased selects assignee based on geographical matching
func (s *AssignmentService) selectGeographyBased(assigneeIDs []uint, contact *models.Contact) (uint, error) {
	// Look for assignees with matching territories
	for _, userID := range assigneeIDs {
		var workload models.UserWorkload
		if s.db.Where("user_id = ?", userID).First(&workload).Error == nil {
			if s.matchesTerritory(&workload, contact) {
				return userID, nil
			}
		}
	}

	// No geographical match, use load-based
	return s.selectLoadBased(assigneeIDs)
}

// selectValueBased selects assignee based on contact value
func (s *AssignmentService) selectValueBased(assigneeIDs []uint, contact *models.Contact) (uint, error) {
	// High-value contacts go to more experienced users (higher conversion rates)
	if contact.EstimatedValue > 50000 || contact.LeadScore > 80 {
		// Find assignee with highest conversion rate
		var bestUser uint
		var bestRate float64 = -1

		for _, userID := range assigneeIDs {
			var workload models.UserWorkload
			if s.db.Where("user_id = ?", userID).First(&workload).Error == nil {
				if workload.IsAvailable && workload.ConversionRate > bestRate {
					bestRate = workload.ConversionRate
					bestUser = userID
				}
			}
		}

		if bestUser != 0 {
			return bestUser, nil
		}
	}

	// Use load-based for other contacts
	return s.selectLoadBased(assigneeIDs)
}

// Helper methods

// extractUserIDs extracts user IDs from JSONArray
func (s *AssignmentService) extractUserIDs(jsonArray models.JSONArray) []uint {
	var userIDs []uint
	for _, val := range jsonArray {
		if id, ok := val.(float64); ok {
			userIDs = append(userIDs, uint(id))
		} else if id, ok := val.(int); ok {
			userIDs = append(userIDs, uint(id))
		}
	}
	return userIDs
}

// isWithinBusinessHours checks if current time is within business hours
func (s *AssignmentService) isWithinBusinessHours(rule *models.AssignmentRule) bool {
	// This is a simplified version - would need proper timezone handling
	now := time.Now()
	weekday := strings.ToLower(now.Weekday().String()[:3])

	// Check working days
	if rule.WorkingDays != nil {
		dayFound := false
		for _, day := range rule.WorkingDays {
			if dayStr, ok := day.(string); ok && dayStr == weekday {
				dayFound = true
				break
			}
		}
		if !dayFound {
			return false
		}
	}

	// Check business hours
	if rule.BusinessHoursStart != nil && rule.BusinessHoursEnd != nil {
		currentTime := now.Format("15:04")
		return currentTime >= *rule.BusinessHoursStart && currentTime <= *rule.BusinessHoursEnd
	}

	return true
}

// checkRateLimits checks if assignment is within rate limits
func (s *AssignmentService) checkRateLimits(rule *models.AssignmentRule) bool {
	now := time.Now()

	// Check hourly limit
	if rule.MaxAssignmentsPerHour != nil {
		hourAgo := now.Add(-time.Hour)
		var count int64
		s.db.Model(&models.ContactAssignment{}).
			Where("rule_id = ? AND created_at > ?", rule.ID, hourAgo).
			Count(&count)
		if int(count) >= *rule.MaxAssignmentsPerHour {
			return false
		}
	}

	// Check daily limit
	if rule.MaxAssignmentsPerDay != nil {
		dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		var count int64
		s.db.Model(&models.ContactAssignment{}).
			Where("rule_id = ? AND created_at > ?", rule.ID, dayStart).
			Count(&count)
		if int(count) >= *rule.MaxAssignmentsPerDay {
			return false
		}
	}

	return true
}

// matchesTerritory checks if contact matches user's territory
func (s *AssignmentService) matchesTerritory(workload *models.UserWorkload, contact *models.Contact) bool {
	if workload.Territories == nil {
		return false
	}

	for _, territory := range workload.Territories {
		if territoryStr, ok := territory.(string); ok {
			// Simple string matching for now
			if strings.EqualFold(territoryStr, contact.Country) ||
				(contact.State != nil && strings.EqualFold(territoryStr, *contact.State)) ||
				(contact.City != nil && strings.EqualFold(territoryStr, *contact.City)) {
				return true
			}
		}
	}

	return false
}

// calculateWorkloadScore calculates a workload score for assignment purposes
func (s *AssignmentService) calculateWorkloadScore(workload *models.UserWorkload) float64 {
	score := float64(workload.ActiveContacts) * 1.0
	score += float64(workload.PendingContacts) * 0.8
	score += float64(workload.OverdueContacts) * 2.0
	
	// Factor in response time (higher response time = higher score = less desirable)
	if workload.AvgResponseTimeHours > 24 {
		score += 5.0
	} else if workload.AvgResponseTimeHours > 8 {
		score += 2.0
	}

	return score
}

// calculateAvailabilityScore calculates availability score
func (s *AssignmentService) calculateAvailabilityScore(workload *models.UserWorkload) float64 {
	if !workload.IsAvailable {
		return 0.0
	}

	score := 1.0

	// Check daily limits
	if workload.MaxDailyAssignments != nil && workload.TodayAssignments >= *workload.MaxDailyAssignments {
		score *= 0.5
	}

	// Check active contact limits
	if workload.MaxActiveContacts != nil && workload.ActiveContacts >= *workload.MaxActiveContacts {
		score *= 0.3
	}

	// Factor in conversion rate
	score *= (1.0 + workload.ConversionRate)

	return math.Min(score, 1.0)
}

// generateAssignmentReason generates a human-readable assignment reason
func (s *AssignmentService) generateAssignmentReason(rule *models.AssignmentRule, contact *models.Contact) string {
	switch rule.Type {
	case models.AssignmentRuleRoundRobin:
		return fmt.Sprintf("Round-robin assignment via rule: %s", rule.Name)
	case models.AssignmentRuleLoadBased:
		return fmt.Sprintf("Load-based assignment via rule: %s", rule.Name)
	case models.AssignmentRuleSkillBased:
		return fmt.Sprintf("Skill-based assignment via rule: %s", rule.Name)
	case models.AssignmentRuleGeographyBased:
		return fmt.Sprintf("Geography-based assignment via rule: %s", rule.Name)
	case models.AssignmentRuleValueBased:
		return fmt.Sprintf("Value-based assignment via rule: %s", rule.Name)
	default:
		return fmt.Sprintf("Automatic assignment via rule: %s", rule.Name)
	}
}

// fallbackAssignment handles assignment when no rules match
func (s *AssignmentService) fallbackAssignment(contact *models.Contact) (*models.ContactAssignment, error) {
	// Find available users with hr_manager or admin role
	var users []models.AdminUser
	if err := s.db.Where("role IN ? AND is_active = ?", []string{"admin", "hr_manager"}, true).
		Find(&users).Error; err != nil || len(users) == 0 {
		return nil, fmt.Errorf("no available users for fallback assignment")
	}

	// Use load-based selection
	userIDs := make([]uint, len(users))
	for i, user := range users {
		userIDs[i] = user.ID
	}

	service := &AssignmentService{db: s.db}
	assigneeID, err := service.selectLoadBased(userIDs)
	if err != nil {
		assigneeID = users[0].ID // Ultimate fallback
	}

	// Create assignment
	assignment := &models.ContactAssignment{
		ContactID:        contact.ID,
		AssignedToID:     assigneeID,
		AssignmentType:   "automatic",
		AssignmentReason: "Fallback assignment - no matching rules",
		Priority:         contact.Priority,
		Status:           "active",
	}

	if err := s.db.Create(assignment).Error; err != nil {
		return nil, fmt.Errorf("failed to create fallback assignment: %v", err)
	}

	// Update contact
	now := time.Now()
	s.db.Model(contact).Updates(map[string]interface{}{
		"assigned_to": assigneeID,
		"assigned_at": now,
	})

	// Update user workload
	s.updateUserWorkload(assigneeID)

	// Log assignment history
	s.logAssignmentHistory(contact.ID, nil, &assigneeID, nil, nil, "assigned", 
		"Fallback assignment - no matching rules")

	return assignment, nil
}

// updateRuleStatistics updates assignment rule statistics
func (s *AssignmentService) updateRuleStatistics(rule *models.AssignmentRule) {
	now := time.Now()
	s.db.Model(rule).Updates(map[string]interface{}{
		"total_assignments":     gorm.Expr("total_assignments + 1"),
		"successful_assignments": gorm.Expr("successful_assignments + 1"),
		"last_assignment_at":    now,
	})
}

// updateUserWorkload recalculates and updates user workload
func (s *AssignmentService) updateUserWorkload(userID uint) {
	// This would typically be called asynchronously
	// For now, we'll do a simple update
	now := time.Now()
	
	var workload models.UserWorkload
	if s.db.Where("user_id = ?", userID).First(&workload).Error != nil {
		workload = models.UserWorkload{
			UserID:      userID,
			IsAvailable: true,
		}
		s.db.Create(&workload)
	}

	// Count active contacts
	var activeCount int64
	s.db.Model(&models.Contact{}).Where("assigned_to = ? AND status NOT IN ?", 
		userID, []string{"closed_won", "closed_lost"}).Count(&activeCount)

	// Count today's assignments
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var todayCount int64
	s.db.Model(&models.ContactAssignment{}).
		Where("assigned_to_id = ? AND created_at >= ?", userID, todayStart).
		Count(&todayCount)

	// Update workload
	s.db.Model(&workload).Updates(map[string]interface{}{
		"active_contacts":     activeCount,
		"today_assignments":   todayCount,
		"last_calculated_at":  now,
	})
}

// logAssignmentHistory logs assignment changes
func (s *AssignmentService) logAssignmentHistory(contactID uint, fromUserID, toUserID *uint, changedByID, ruleID *uint, changeType, reason string) {
	history := &models.AssignmentHistory{
		ContactID:    contactID,
		FromUserID:   fromUserID,
		ToUserID:     *toUserID,
		ChangedByID:  changedByID,
		RuleID:       ruleID,
		ChangeType:   changeType,
		ChangeReason: reason,
	}

	s.db.Create(history)
}