package services

import (
	"contact-service/internal/models"
	"contact-service/pkg/logger"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// LifecycleService handles contact lifecycle management, lead scoring, and status transitions
type LifecycleService struct {
	db *gorm.DB
}

// NewLifecycleService creates a new lifecycle service
func NewLifecycleService(db *gorm.DB) *LifecycleService {
	return &LifecycleService{db: db}
}

// ScoreContact calculates and updates the lead score for a contact
func (s *LifecycleService) ScoreContact(contactID uint, forceRescore bool, reason string, scoredByUserID *uint) (*models.ContactLifecycleResponse, error) {
	// Get the contact with related data
	var contact models.Contact
	if err := s.db.Preload("ContactType").Preload("ContactSource").
		First(&contact, contactID).Error; err != nil {
		return nil, fmt.Errorf("contact not found: %v", err)
	}

	// Get or create lifecycle record
	var lifecycle models.ContactLifecycle
	if err := s.db.Where("contact_id = ?", contactID).First(&lifecycle).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new lifecycle record
			lifecycle = models.ContactLifecycle{
				ContactID:       contactID,
				CurrentStatus:   contact.Status,
				CurrentStage:    models.StageUnknown,
				CurrentScore:    0,
				StageEnteredAt:  time.Now(),
				StatusEnteredAt: time.Now(),
				LastScoredAt:    time.Now(),
				ScoreHistory:    make(models.JSONArray, 0),
				ScoringFactors:  make(models.JSONMap),
				StageVelocity:   make(models.JSONMap),
			}
		} else {
			return nil, fmt.Errorf("failed to get lifecycle record: %v", err)
		}
	}

	// Check if we need to rescore
	if !forceRescore && s.isRecentlyScored(&lifecycle) {
		return s.buildLifecycleResponse(&lifecycle, &contact)
	}

	// Get active scoring rules
	var rules []models.LeadScoringRule
	if err := s.db.Where("is_active = ? AND deleted_at IS NULL", true).
		Order("priority DESC, created_at ASC").Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("failed to get scoring rules: %v", err)
	}

	// Calculate scores
	totalScore := 0
	demoScore := 0
	behavioralScore := 0
	engagementScore := 0
	firmoScore := 0
	scoringFactors := make(models.JSONMap)
	appliedRules := make([]string, 0)

	for _, rule := range rules {
		if s.isRuleApplicable(&rule, &contact) {
			ruleScore := s.calculateRuleScore(&rule, &contact)
			totalScore += ruleScore

			// Categorize scores
			switch rule.Category {
			case "demographic":
				demoScore += ruleScore
			case "behavioral":
				behavioralScore += ruleScore
			case "engagement":
				engagementScore += ruleScore
			case "firmographic":
				firmoScore += ruleScore
			}

			if ruleScore > 0 {
				appliedRules = append(appliedRules, rule.Name)
				scoringFactors[rule.Name] = map[string]interface{}{
					"score":    ruleScore,
					"category": rule.Category,
					"reason":   s.getRuleScoreReason(&rule, &contact),
				}

				// Update rule statistics
				s.updateRuleStatistics(&rule)
			}
		}
	}

	// Ensure score doesn't exceed maximum
	if totalScore > 100 {
		totalScore = 100
	}

	// Save score history
	scoreSnapshot := models.ScoreSnapshot{
		ContactID:       contactID,
		Timestamp:       time.Now(),
		TotalScore:      totalScore,
		DemoScore:       demoScore,
		BehavioralScore: behavioralScore,
		EngagementScore: engagementScore,
		FirmoScore:      firmoScore,
		Factors:         scoringFactors,
	}

	// Update score history (keep last 10 entries)
	scoreHistory := lifecycle.ScoreHistory
	if scoreHistory == nil {
		scoreHistory = make(models.JSONArray, 0)
	}
	scoreHistory = append(scoreHistory, scoreSnapshot)
	if len(scoreHistory) > 10 {
		scoreHistory = scoreHistory[1:]
	}

	// Update lifecycle record
	previousScore := lifecycle.CurrentScore
	now := time.Now()
	
	updates := map[string]interface{}{
		"current_score":      totalScore,
		"demographic_score":  demoScore,
		"behavioral_score":   behavioralScore,
		"engagement_score":   engagementScore,
		"firmographic_score": firmoScore,
		"score_history":      scoreHistory,
		"scoring_factors":    scoringFactors,
		"last_scored_at":     now,
	}

	// Update lifecycle stage based on score
	newStage := s.determineLifecycleStage(totalScore, &contact)
	if newStage != lifecycle.CurrentStage {
		updates["current_stage"] = newStage
		updates["stage_entered_at"] = now
		
		// Record stage change event
		s.recordLifecycleEvent(contactID, lifecycle.ID, "stage_change", 
			fmt.Sprintf("Stage changed from %s to %s", lifecycle.CurrentStage, newStage),
			string(lifecycle.CurrentStage), string(newStage), nil, scoredByUserID, "automatic", "scoring")
	}

	// Calculate days in current status/stage
	updates["days_in_current_status"] = int(time.Since(lifecycle.StatusEnteredAt).Hours() / 24)
	updates["days_in_current_stage"] = int(time.Since(lifecycle.StageEnteredAt).Hours() / 24)
	updates["total_lifecycle_days"] = int(time.Since(lifecycle.CreatedAt).Hours() / 24)

	// Update the lifecycle record
	if lifecycle.ID == 0 {
		lifecycle.CurrentScore = totalScore
		lifecycle.DemographicScore = demoScore
		lifecycle.BehavioralScore = behavioralScore
		lifecycle.EngagementScore = engagementScore
		lifecycle.FirmographicScore = firmoScore
		lifecycle.ScoreHistory = scoreHistory
		lifecycle.ScoringFactors = scoringFactors
		lifecycle.LastScoredAt = now
		lifecycle.CurrentStage = newStage
		if err := s.db.Create(&lifecycle).Error; err != nil {
			return nil, fmt.Errorf("failed to create lifecycle record: %v", err)
		}
	} else {
		if err := s.db.Model(&lifecycle).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update lifecycle record: %v", err)
		}
		// Reload the record
		s.db.First(&lifecycle, lifecycle.ID)
	}

	// Update contact's lead score
	if err := s.db.Model(&contact).Update("lead_score", totalScore).Error; err != nil {
		logger.Error("Failed to update contact lead score", err, map[string]interface{}{
			"contact_id": contactID,
			"new_score":  totalScore,
		})
	}

	// Record scoring event
	changeAmount := totalScore - previousScore
	s.recordLifecycleEvent(contactID, lifecycle.ID, "score_change",
		fmt.Sprintf("Lead score updated from %d to %d (%+d)", previousScore, totalScore, changeAmount),
		strconv.Itoa(previousScore), strconv.Itoa(totalScore), &changeAmount, scoredByUserID, "automatic", "scoring")

	// Check for automatic status transitions
	s.checkAutomaticStatusTransitions(&contact, &lifecycle)

	logger.Info("Contact scored successfully", map[string]interface{}{
		"contact_id":     contactID,
		"previous_score": previousScore,
		"new_score":      totalScore,
		"change":         changeAmount,
		"applied_rules":  len(appliedRules),
		"reason":         reason,
	})

	return s.buildLifecycleResponse(&lifecycle, &contact)
}

// ChangeContactStatus manually changes a contact's status
func (s *LifecycleService) ChangeContactStatus(request *models.StatusChangeRequest, changedByUserID uint) error {
	// Get the contact
	var contact models.Contact
	if err := s.db.First(&contact, request.ContactID).Error; err != nil {
		return fmt.Errorf("contact not found: %v", err)
	}

	// Check if status change is valid
	if contact.Status == request.NewStatus {
		return fmt.Errorf("contact is already in %s status", request.NewStatus)
	}

	// Check transition rules if not forced
	if !request.ForceChange {
		if !s.isStatusTransitionAllowed(contact.Status, request.NewStatus) {
			return fmt.Errorf("transition from %s to %s is not allowed", contact.Status, request.NewStatus)
		}
	}

	// Get lifecycle record
	var lifecycle models.ContactLifecycle
	if err := s.db.Where("contact_id = ?", request.ContactID).First(&lifecycle).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create lifecycle record if it doesn't exist
			lifecycle = models.ContactLifecycle{
				ContactID:       request.ContactID,
				CurrentStatus:   contact.Status,
				CurrentStage:    models.StageUnknown,
				CurrentScore:    contact.LeadScore,
				StageEnteredAt:  time.Now(),
				StatusEnteredAt: time.Now(),
				LastScoredAt:    time.Now(),
			}
			s.db.Create(&lifecycle)
		} else {
			return fmt.Errorf("failed to get lifecycle record: %v", err)
		}
	}

	// Update contact status
	previousStatus := contact.Status
	now := time.Now()
	
	if err := s.db.Model(&contact).Updates(map[string]interface{}{
		"status":              request.NewStatus,
		"last_activity_date": now,
	}).Error; err != nil {
		return fmt.Errorf("failed to update contact status: %v", err)
	}

	// Update lifecycle record
	updates := map[string]interface{}{
		"current_status":     request.NewStatus,
		"status_entered_at":  now,
		"days_in_current_status": 0,
	}

	// Update milestones based on new status
	switch request.NewStatus {
	case models.StatusQualified:
		if lifecycle.QualificationAt == nil {
			updates["qualification_at"] = now
		}
	case models.StatusProposal, models.StatusNegotiation:
		if lifecycle.OpportunityAt == nil {
			updates["opportunity_at"] = now
		}
	case models.StatusClosedWon:
		if lifecycle.ConversionAt == nil {
			updates["conversion_at"] = now
		}
	}

	if err := s.db.Model(&lifecycle).Updates(updates).Error; err != nil {
		logger.Error("Failed to update lifecycle on status change", err, map[string]interface{}{
			"contact_id": request.ContactID,
			"new_status": request.NewStatus,
		})
	}

	// Record status change event
	s.recordLifecycleEvent(request.ContactID, lifecycle.ID, "status_change",
		fmt.Sprintf("Status changed from %s to %s: %s", previousStatus, request.NewStatus, request.Reason),
		string(previousStatus), string(request.NewStatus), nil, &changedByUserID, "manual", "user_action")

	logger.Info("Contact status changed", map[string]interface{}{
		"contact_id":      request.ContactID,
		"previous_status": previousStatus,
		"new_status":      request.NewStatus,
		"changed_by":      changedByUserID,
		"reason":          request.Reason,
	})

	return nil
}

// BulkChangeContactStatus changes status for multiple contacts
func (s *LifecycleService) BulkChangeContactStatus(request *models.BulkStatusChangeRequest, changedByUserID uint) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	successCount := 0
	var lastError error

	for _, contactID := range request.ContactIDs {
		statusRequest := &models.StatusChangeRequest{
			ContactID:   contactID,
			NewStatus:   request.NewStatus,
			Reason:      request.Reason,
			ForceChange: request.ForceChange,
		}

		// Use a new service instance with the transaction
		txService := &LifecycleService{db: tx}
		if err := txService.ChangeContactStatus(statusRequest, changedByUserID); err != nil {
			logger.Error("Failed to change status in bulk operation", err, map[string]interface{}{
				"contact_id": contactID,
				"new_status": request.NewStatus,
			})
			lastError = err
			continue
		}
		successCount++
	}

	if successCount == 0 {
		tx.Rollback()
		return fmt.Errorf("failed to change status for any contacts: %v", lastError)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit bulk status change: %v", err)
	}

	logger.Info("Bulk status change completed", map[string]interface{}{
		"total_contacts": len(request.ContactIDs),
		"successful":     successCount,
		"failed":        len(request.ContactIDs) - successCount,
		"new_status":    request.NewStatus,
		"changed_by":    changedByUserID,
	})

	return nil
}

// GetContactLifecycle gets the lifecycle information for a contact
func (s *LifecycleService) GetContactLifecycle(contactID uint) (*models.ContactLifecycleResponse, error) {
	var contact models.Contact
	if err := s.db.Preload("ContactType").Preload("ContactSource").
		First(&contact, contactID).Error; err != nil {
		return nil, fmt.Errorf("contact not found: %v", err)
	}

	var lifecycle models.ContactLifecycle
	if err := s.db.Where("contact_id = ?", contactID).First(&lifecycle).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create and return basic lifecycle
			lifecycle = models.ContactLifecycle{
				ContactID:       contactID,
				CurrentStatus:   contact.Status,
				CurrentStage:    models.StageUnknown,
				CurrentScore:    contact.LeadScore,
				StageEnteredAt:  contact.CreatedAt,
				StatusEnteredAt: contact.CreatedAt,
				LastScoredAt:    contact.CreatedAt,
			}
		} else {
			return nil, fmt.Errorf("failed to get lifecycle record: %v", err)
		}
	}

	return s.buildLifecycleResponse(&lifecycle, &contact)
}

// GetLifecycleEvents gets the lifecycle events for a contact
func (s *LifecycleService) GetLifecycleEvents(contactID uint, limit int) ([]models.LifecycleEvent, error) {
	var events []models.LifecycleEvent
	query := s.db.Where("contact_id = ?", contactID).
		Preload("TriggeredByUser").
		Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&events).Error; err != nil {
		return nil, fmt.Errorf("failed to get lifecycle events: %v", err)
	}

	return events, nil
}

// AnalyzeContactScoring provides detailed scoring analysis for a contact
func (s *LifecycleService) AnalyzeContactScoring(contactID uint) (*models.ScoringAnalysisResponse, error) {
	// Get contact and lifecycle
	var contact models.Contact
	if err := s.db.Preload("ContactType").Preload("ContactSource").
		First(&contact, contactID).Error; err != nil {
		return nil, fmt.Errorf("contact not found: %v", err)
	}

	var lifecycle models.ContactLifecycle
	if err := s.db.Where("contact_id = ?", contactID).First(&lifecycle).Error; err != nil {
		return nil, fmt.Errorf("lifecycle record not found: %v", err)
	}

	// Get active scoring rules for analysis
	var rules []models.LeadScoringRule
	if err := s.db.Where("is_active = ? AND deleted_at IS NULL", true).
		Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("failed to get scoring rules: %v", err)
	}

	// Calculate max possible score
	maxPossibleScore := 0
	appliedRules := make([]string, 0)
	
	for _, rule := range rules {
		if s.isRuleApplicable(&rule, &contact) {
			maxPossibleScore += rule.MaxScore
			appliedRules = append(appliedRules, rule.Name)
		}
	}

	// Build category breakdown
	categoryBreakdown := map[string]int{
		"demographic":  lifecycle.DemographicScore,
		"behavioral":   lifecycle.BehavioralScore,
		"engagement":   lifecycle.EngagementScore,
		"firmographic": lifecycle.FirmographicScore,
	}

	// Calculate score percentage
	scorePercentage := 0.0
	if maxPossibleScore > 0 {
		scorePercentage = float64(lifecycle.CurrentScore) / float64(maxPossibleScore) * 100
	}

	// Generate recommendations
	recommendations := s.generateScoringRecommendations(&lifecycle, &contact)

	response := &models.ScoringAnalysisResponse{
		ContactID:         contactID,
		TotalScore:        lifecycle.CurrentScore,
		MaxPossibleScore:  maxPossibleScore,
		ScorePercentage:   scorePercentage,
		Grade:             s.calculateScoreGrade(lifecycle.CurrentScore),
		CategoryBreakdown: categoryBreakdown,
		AppliedRules:      appliedRules,
		ScoringFactors:    lifecycle.ScoringFactors,
		Recommendations:   recommendations,
		LastUpdated:       lifecycle.LastScoredAt,
	}

	return response, nil
}

// checkAutomaticStatusTransitions checks and applies automatic status transitions
func (s *LifecycleService) checkAutomaticStatusTransitions(contact *models.Contact, lifecycle *models.ContactLifecycle) {
	// Get active transition rules
	var rules []models.StatusTransitionRule
	if err := s.db.Where("is_active = ? AND from_status = ? AND deleted_at IS NULL", 
		true, contact.Status).Order("priority DESC").Find(&rules).Error; err != nil {
		logger.Error("Failed to get transition rules", err, map[string]interface{}{
			"contact_id": contact.ID,
		})
		return
	}

	for _, rule := range rules {
		if s.shouldTriggerTransition(&rule, contact, lifecycle) {
			s.executeStatusTransition(&rule, contact, lifecycle)
			break // Only execute one transition at a time
		}
	}
}

// Helper methods

// isRecentlyScored checks if the contact was scored recently
func (s *LifecycleService) isRecentlyScored(lifecycle *models.ContactLifecycle) bool {
	// Consider it recent if scored within the last hour
	return time.Since(lifecycle.LastScoredAt) < time.Hour
}

// isRuleApplicable checks if a scoring rule applies to a contact
func (s *LifecycleService) isRuleApplicable(rule *models.LeadScoringRule, contact *models.Contact) bool {
	// Check if rule has applicability conditions
	if rule.ApplicableWhen == nil || len(rule.ApplicableWhen) == 0 {
		return true
	}

	// Evaluate conditions (similar to assignment rule evaluation)
	for _, condition := range rule.ApplicableWhen {
		if !s.evaluateCondition(condition, contact) {
			return false
		}
	}

	return true
}

// calculateRuleScore calculates the score contribution of a rule for a contact
func (s *LifecycleService) calculateRuleScore(rule *models.LeadScoringRule, contact *models.Contact) int {
	totalScore := rule.BaseScore

	for _, criteria := range rule.Criteria {
		if s.evaluateScoringCriteria(criteria, contact) {
			score := int(float64(criteria.Score) * criteria.Weight)
			totalScore += score
		}
	}

	// Ensure score doesn't exceed rule's maximum
	if totalScore > rule.MaxScore {
		totalScore = rule.MaxScore
	}

	return totalScore
}

// evaluateCondition evaluates a condition against a contact
func (s *LifecycleService) evaluateCondition(condition models.AssignmentCondition, contact *models.Contact) bool {
	fieldValue := s.getContactFieldValue(condition.Field, contact)
	
	switch condition.Operator {
	case "equals":
		return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", condition.Value)
	case "contains":
		fieldStr := strings.ToLower(fmt.Sprintf("%v", fieldValue))
		conditionStr := strings.ToLower(fmt.Sprintf("%v", condition.Value))
		return strings.Contains(fieldStr, conditionStr)
	case "greater_than":
		return s.toFloat64(fieldValue) > s.toFloat64(condition.Value)
	case "less_than":
		return s.toFloat64(fieldValue) < s.toFloat64(condition.Value)
	default:
		return false
	}
}

// evaluateScoringCriteria evaluates scoring criteria against a contact
func (s *LifecycleService) evaluateScoringCriteria(criteria models.ScoringCriteria, contact *models.Contact) bool {
	fieldValue := s.getContactFieldValue(criteria.Field, contact)
	
	switch criteria.Operator {
	case "equals":
		return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", criteria.Value)
	case "contains":
		fieldStr := strings.ToLower(fmt.Sprintf("%v", fieldValue))
		criteriaStr := strings.ToLower(fmt.Sprintf("%v", criteria.Value))
		return strings.Contains(fieldStr, criteriaStr)
	case "greater_than":
		return s.toFloat64(fieldValue) > s.toFloat64(criteria.Value)
	case "less_than":
		return s.toFloat64(fieldValue) < s.toFloat64(criteria.Value)
	default:
		return false
	}
}

// getContactFieldValue gets field value from contact
func (s *LifecycleService) getContactFieldValue(fieldName string, contact *models.Contact) interface{} {
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
	case "email_opened":
		return contact.EmailOpened
	case "email_clicked":
		return contact.EmailClicked
	case "total_interactions":
		return contact.TotalInteractions
	default:
		return ""
	}
}

// toFloat64 converts interface{} to float64
func (s *LifecycleService) toFloat64(val interface{}) float64 {
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

// determineLifecycleStage determines lifecycle stage based on score
func (s *LifecycleService) determineLifecycleStage(score int, contact *models.Contact) models.LifecycleStage {
	switch {
	case score >= 80:
		return models.StageSalesQL
	case score >= 60:
		return models.StageMarketingQL
	case score >= 40:
		return models.StageProspect
	case score >= 20:
		return models.StageSuspect
	default:
		return models.StageUnknown
	}
}

// getRuleScoreReason generates a human-readable reason for rule scoring
func (s *LifecycleService) getRuleScoreReason(rule *models.LeadScoringRule, contact *models.Contact) string {
	reasons := make([]string, 0)
	
	for _, criteria := range rule.Criteria {
		if s.evaluateScoringCriteria(criteria, contact) {
			reasons = append(reasons, criteria.Description)
		}
	}
	
	if len(reasons) > 0 {
		return strings.Join(reasons, ", ")
	}
	
	return fmt.Sprintf("Base score from %s rule", rule.Name)
}

// updateRuleStatistics updates scoring rule statistics
func (s *LifecycleService) updateRuleStatistics(rule *models.LeadScoringRule) {
	now := time.Now()
	s.db.Model(rule).Updates(map[string]interface{}{
		"times_applied":   gorm.Expr("times_applied + 1"),
		"last_applied_at": now,
	})
}

// isStatusTransitionAllowed checks if a status transition is allowed
func (s *LifecycleService) isStatusTransitionAllowed(fromStatus, toStatus models.ContactStatus) bool {
	// Define allowed transitions (simplified business logic)
	allowedTransitions := map[models.ContactStatus][]models.ContactStatus{
		models.StatusNew:          {models.StatusContacted, models.StatusQualified, models.StatusClosedLost},
		models.StatusContacted:    {models.StatusQualified, models.StatusNurturing, models.StatusClosedLost},
		models.StatusQualified:    {models.StatusProposal, models.StatusNurturing, models.StatusClosedLost},
		models.StatusProposal:     {models.StatusNegotiation, models.StatusQualified, models.StatusClosedLost},
		models.StatusNegotiation:  {models.StatusClosedWon, models.StatusClosedLost, models.StatusOnHold},
		models.StatusNurturing:    {models.StatusQualified, models.StatusContacted, models.StatusClosedLost},
		models.StatusOnHold:       {models.StatusNegotiation, models.StatusQualified, models.StatusClosedLost},
	}

	if allowedStatuses, exists := allowedTransitions[fromStatus]; exists {
		for _, status := range allowedStatuses {
			if status == toStatus {
				return true
			}
		}
	}

	return false
}

// shouldTriggerTransition checks if a transition rule should be triggered
func (s *LifecycleService) shouldTriggerTransition(rule *models.StatusTransitionRule, contact *models.Contact, lifecycle *models.ContactLifecycle) bool {
	// Check score requirement
	if rule.RequiredScore > 0 && lifecycle.CurrentScore < rule.RequiredScore {
		return false
	}

	// Check days in status
	if rule.DaysInStatus > 0 && lifecycle.DaysInCurrentStatus < rule.DaysInStatus {
		return false
	}

	// Check conditions
	for _, condition := range rule.Conditions {
		if !s.evaluateCondition(condition, contact) {
			return false
		}
	}

	return true
}

// executeStatusTransition executes an automatic status transition
func (s *LifecycleService) executeStatusTransition(rule *models.StatusTransitionRule, contact *models.Contact, lifecycle *models.ContactLifecycle) {
	previousStatus := contact.Status
	now := time.Now()

	// Update contact status
	if err := s.db.Model(contact).Updates(map[string]interface{}{
		"status":              rule.ToStatus,
		"last_activity_date": now,
	}).Error; err != nil {
		logger.Error("Failed to execute automatic status transition", err, map[string]interface{}{
			"contact_id": contact.ID,
			"rule_id":    rule.ID,
			"to_status":  rule.ToStatus,
		})
		return
	}

	// Update lifecycle
	updates := map[string]interface{}{
		"current_status":     rule.ToStatus,
		"status_entered_at":  now,
		"days_in_current_status": 0,
	}

	s.db.Model(lifecycle).Updates(updates)

	// Record transition event
	s.recordLifecycleEvent(contact.ID, lifecycle.ID, "status_change",
		fmt.Sprintf("Automatic status transition from %s to %s via rule: %s", previousStatus, rule.ToStatus, rule.Name),
		string(previousStatus), string(rule.ToStatus), nil, nil, "automatic", fmt.Sprintf("rule:%d", rule.ID))

	// Update rule statistics
	s.db.Model(rule).Updates(map[string]interface{}{
		"times_triggered":   gorm.Expr("times_triggered + 1"),
		"last_triggered_at": now,
	})

	logger.Info("Automatic status transition executed", map[string]interface{}{
		"contact_id":      contact.ID,
		"rule_id":         rule.ID,
		"previous_status": previousStatus,
		"new_status":      rule.ToStatus,
	})
}

// recordLifecycleEvent records a lifecycle event
func (s *LifecycleService) recordLifecycleEvent(contactID, lifecycleID uint, eventType, description, prevValue, newValue string, changeAmount *int, triggeredBy *uint, triggerType, triggerSource string) {
	event := &models.LifecycleEvent{
		ContactID:        contactID,
		LifecycleID:      lifecycleID,
		EventType:        eventType,
		EventName:        description,
		EventDescription: description,
		PreviousValue:    &prevValue,
		NewValue:         newValue,
		ChangeAmount:     changeAmount,
		TriggerType:      triggerType,
		TriggerSource:    triggerSource,
		TriggeredBy:      triggeredBy,
	}

	if err := s.db.Create(event).Error; err != nil {
		logger.Error("Failed to record lifecycle event", err, map[string]interface{}{
			"contact_id":   contactID,
			"event_type":   eventType,
			"description":  description,
		})
	}
}

// buildLifecycleResponse builds a lifecycle response with computed fields
func (s *LifecycleService) buildLifecycleResponse(lifecycle *models.ContactLifecycle, contact *models.Contact) (*models.ContactLifecycleResponse, error) {
	var contactResponse *models.ContactResponse
	if contact != nil {
		contactResponse = &models.ContactResponse{
			ID:          contact.ID,
			FirstName:   contact.FirstName,
			LastName:    contact.LastName,
			FullName:    contact.GetFullName(),
			DisplayName: contact.GetDisplayName(),
			Email:       contact.Email,
			Phone:       contact.Phone,
			Company:     contact.Company,
			Status:      contact.Status,
			Priority:    contact.Priority,
			LeadScore:   contact.LeadScore,
		}
	}

	response := &models.ContactLifecycleResponse{
		ID:                  lifecycle.ID,
		ContactID:           lifecycle.ContactID,
		Contact:             contactResponse,
		CurrentStage:        lifecycle.CurrentStage,
		CurrentStatus:       lifecycle.CurrentStatus,
		CurrentScore:        lifecycle.CurrentScore,
		StageEnteredAt:      lifecycle.StageEnteredAt,
		StatusEnteredAt:     lifecycle.StatusEnteredAt,
		LastScoredAt:        lifecycle.LastScoredAt,
		DemographicScore:    lifecycle.DemographicScore,
		BehavioralScore:     lifecycle.BehavioralScore,
		EngagementScore:     lifecycle.EngagementScore,
		FirmographicScore:   lifecycle.FirmographicScore,
		ScoreHistory:        lifecycle.ScoreHistory,
		ScoringFactors:      lifecycle.ScoringFactors,
		FirstEngagementAt:   lifecycle.FirstEngagementAt,
		QualificationAt:     lifecycle.QualificationAt,
		OpportunityAt:       lifecycle.OpportunityAt,
		ConversionAt:        lifecycle.ConversionAt,
		DaysInCurrentStage:  int(time.Since(lifecycle.StageEnteredAt).Hours() / 24),
		DaysInCurrentStatus: int(time.Since(lifecycle.StatusEnteredAt).Hours() / 24),
		TotalLifecycleDays:  int(time.Since(lifecycle.CreatedAt).Hours() / 24),
		StageVelocity:       lifecycle.StageVelocity,
		ConversionRate:      lifecycle.ConversionRate,
		CreatedAt:           lifecycle.CreatedAt,
		UpdatedAt:           lifecycle.UpdatedAt,
		ScoreGrade:          s.calculateScoreGrade(lifecycle.CurrentScore),
		QualificationStatus: s.determineQualificationStatus(lifecycle.CurrentScore, lifecycle.CurrentStatus),
		NextSuggestedAction: s.suggestNextAction(lifecycle.CurrentScore, lifecycle.CurrentStatus, lifecycle.CurrentStage),
	}

	return response, nil
}

// calculateScoreGrade calculates letter grade based on score
func (s *LifecycleService) calculateScoreGrade(score int) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}

// determineQualificationStatus determines qualification status
func (s *LifecycleService) determineQualificationStatus(score int, status models.ContactStatus) string {
	if score >= 60 || status == models.StatusQualified {
		return "qualified"
	} else if score >= 40 {
		return "pending"
	}
	return "unqualified"
}

// suggestNextAction suggests the next action for a contact
func (s *LifecycleService) suggestNextAction(score int, status models.ContactStatus, stage models.LifecycleStage) string {
	switch {
	case score >= 80 && status == models.StatusNew:
		return "Qualify lead and move to sales"
	case score >= 60 && status == models.StatusContacted:
		return "Schedule demo or consultation"
	case score >= 40 && status == models.StatusNew:
		return "Initiate contact via phone or email"
	case status == models.StatusQualified:
		return "Send proposal or pricing information"
	case status == models.StatusProposal:
		return "Follow up on proposal status"
	default:
		return "Continue nurturing with targeted content"
	}
}

// generateScoringRecommendations generates recommendations to improve lead score
func (s *LifecycleService) generateScoringRecommendations(lifecycle *models.ContactLifecycle, contact *models.Contact) []string {
	recommendations := make([]string, 0)

	// Check engagement
	if lifecycle.EngagementScore < 20 {
		recommendations = append(recommendations, "Increase engagement through email campaigns and content marketing")
	}

	// Check company information
	if contact.Company == nil || *contact.Company == "" {
		recommendations = append(recommendations, "Collect company information to improve firmographic scoring")
	}

	// Check contact information completeness
	if contact.Phone == nil {
		recommendations = append(recommendations, "Obtain phone number for better contact scoring")
	}

	// Check behavioral indicators
	if lifecycle.BehavioralScore < 15 {
		recommendations = append(recommendations, "Track website visits and content downloads to improve behavioral scoring")
	}

	// Score-based recommendations
	if lifecycle.CurrentScore < 40 {
		recommendations = append(recommendations, "Focus on lead nurturing to increase overall score")
	} else if lifecycle.CurrentScore >= 60 {
		recommendations = append(recommendations, "Consider moving to sales qualification process")
	}

	return recommendations
}