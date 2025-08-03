package services

import (
	"contact-service/internal/models"
	"contact-service/pkg/logger"
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"
)

// AnalyticsService handles analytics and metrics calculations
type AnalyticsService struct {
	db *gorm.DB
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(db *gorm.DB) *AnalyticsService {
	return &AnalyticsService{db: db}
}

// GetContactAnalytics gets comprehensive contact analytics
func (s *AnalyticsService) GetContactAnalytics(request *models.AnalyticsRequest) (*models.ContactMetricsResponse, error) {
	analytics := &models.ContactAnalytics{
		ContactsByStatus:   make(map[string]int),
		ContactsBySource:   make(map[string]int),
		ContactsByType:     make(map[string]int),
		ContactsByAssignee: make(map[string]int),
	}

	// Get total contacts in date range
	var totalCount int64
	if err := s.db.Model(&models.Contact{}).
		Where("created_at BETWEEN ? AND ?", request.StartDate, request.EndDate).
		Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get total contacts: %v", err)
	}
	analytics.TotalContacts = int(totalCount)

	// Get new contacts (created in period)
	analytics.NewContacts = analytics.TotalContacts

	// Get active contacts (had activity in period)
	var activeCount int64
	if err := s.db.Model(&models.Contact{}).
		Where("last_contact_date BETWEEN ? AND ? OR updated_at BETWEEN ? AND ?", 
			request.StartDate, request.EndDate, request.StartDate, request.EndDate).
		Count(&activeCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get active contacts: %v", err)
	}
	analytics.ActiveContacts = int(activeCount)

	// Get converted contacts
	var convertedCount int64
	if err := s.db.Model(&models.Contact{}).
		Where("status = ? AND updated_at BETWEEN ? AND ?", "converted", request.StartDate, request.EndDate).
		Count(&convertedCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get converted contacts: %v", err)
	}
	analytics.ConvertedContacts = int(convertedCount)

	// Calculate conversion rate
	if analytics.TotalContacts > 0 {
		analytics.ConversionRate = float64(analytics.ConvertedContacts) / float64(analytics.TotalContacts) * 100
	}

	// Get contacts by status
	var statusResults []struct {
		Status string
		Count  int
	}
	if err := s.db.Model(&models.Contact{}).
		Select("status, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", request.StartDate, request.EndDate).
		Group("status").
		Scan(&statusResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get contacts by status: %v", err)
	}

	for _, result := range statusResults {
		analytics.ContactsByStatus[result.Status] = result.Count
	}

	// Get contacts by source
	var sourceResults []struct {
		Source string
		Count  int
	}
	if err := s.db.Model(&models.Contact{}).
		Select("source, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", request.StartDate, request.EndDate).
		Group("source").
		Scan(&sourceResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get contacts by source: %v", err)
	}

	for _, result := range sourceResults {
		analytics.ContactsBySource[result.Source] = result.Count
	}

	// Get contacts by type
	var typeResults []struct {
		Type  string
		Count int
	}
	if err := s.db.Model(&models.Contact{}).
		Select("type, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", request.StartDate, request.EndDate).
		Group("type").
		Scan(&typeResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get contacts by type: %v", err)
	}

	for _, result := range typeResults {
		analytics.ContactsByType[result.Type] = result.Count
	}

	// Get contacts by assignee
	var assigneeResults []struct {
		Username string
		Count    int
	}
	if err := s.db.Table("contacts c").
		Select("au.username, COUNT(*) as count").
		Joins("LEFT JOIN admin_users au ON c.assigned_to = au.id").
		Where("c.created_at BETWEEN ? AND ?", request.StartDate, request.EndDate).
		Group("au.username").
		Scan(&assigneeResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get contacts by assignee: %v", err)
	}

	for _, result := range assigneeResults {
		analytics.ContactsByAssignee[result.Username] = result.Count
	}

	// Calculate growth rate (compared to previous period)
	previousPeriod := request.StartDate.Add(-1 * request.EndDate.Sub(request.StartDate))
	var previousContacts int64
	if err := s.db.Model(&models.Contact{}).
		Where("created_at BETWEEN ? AND ?", previousPeriod, request.StartDate).
		Count(&previousContacts).Error; err == nil && previousContacts > 0 {
		analytics.GrowthRate = (float64(analytics.TotalContacts) - float64(previousContacts)) / float64(previousContacts) * 100
	}

	// Build top sources
	analytics.TopSources = s.buildTopSources(sourceResults, analytics.TotalContacts)

	// Build status distribution
	analytics.StatusDistribution = s.buildStatusDistribution(statusResults, analytics.TotalContacts)

	// Generate trends
	trends, err := s.generateContactTrends(request.StartDate, request.EndDate, request.Granularity)
	if err != nil {
		logger.Error("Failed to generate contact trends", err, nil)
		trends = []models.ContactTrendData{}
	}

	period := fmt.Sprintf("%s to %s", request.StartDate.Format("2006-01-02"), request.EndDate.Format("2006-01-02"))

	return &models.ContactMetricsResponse{
		Period:    period,
		Analytics: *analytics,
		Trends:    trends,
	}, nil
}

// GetAppointmentAnalytics gets comprehensive appointment analytics
func (s *AnalyticsService) GetAppointmentAnalytics(request *models.AnalyticsRequest) (*models.AppointmentMetricsResponse, error) {
	analytics := &models.AppointmentAnalytics{
		AppointmentsByType:   make(map[string]int),
		AppointmentsByStatus: make(map[string]int),
		AppointmentsByUser:   make(map[string]int),
	}

	// Get total appointments in date range
	var appointmentCount int64
	if err := s.db.Model(&models.Appointment{}).
		Where("created_at BETWEEN ? AND ? AND deleted_at IS NULL", request.StartDate, request.EndDate).
		Count(&appointmentCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get total appointments: %v", err)
	}
	analytics.TotalAppointments = int(appointmentCount)

	// Get completed appointments
	var completedCount int64
	if err := s.db.Model(&models.Appointment{}).
		Where("status = ? AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", 
			models.AppointmentCompleted, request.StartDate, request.EndDate).
		Count(&completedCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get completed appointments: %v", err)
	}
	analytics.CompletedAppointments = int(completedCount)

	// Get cancelled appointments
	var cancelledCount int64
	if err := s.db.Model(&models.Appointment{}).
		Where("status = ? AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", 
			models.AppointmentCancelled, request.StartDate, request.EndDate).
		Count(&cancelledCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get cancelled appointments: %v", err)
	}
	analytics.CancelledAppointments = int(cancelledCount)

	// Get upcoming appointments
	now := time.Now()
	var upcomingCount int64
	if err := s.db.Model(&models.Appointment{}).
		Where("start_time > ? AND status NOT IN ? AND deleted_at IS NULL", 
			now, []models.AppointmentStatus{models.AppointmentCancelled, models.AppointmentCompleted}).
		Count(&upcomingCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get upcoming appointments: %v", err)
	}
	analytics.UpcomingAppointments = int(upcomingCount)

	// Calculate rates
	if analytics.TotalAppointments > 0 {
		analytics.CompletionRate = float64(analytics.CompletedAppointments) / float64(analytics.TotalAppointments) * 100
		analytics.CancellationRate = float64(analytics.CancelledAppointments) / float64(analytics.TotalAppointments) * 100
	}

	// Get average rating for completed appointments
	var avgRating struct {
		Average float64
	}
	if err := s.db.Model(&models.Appointment{}).
		Select("AVG(rating) as average").
		Where("status = ? AND rating IS NOT NULL AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", 
			models.AppointmentCompleted, request.StartDate, request.EndDate).
		Scan(&avgRating).Error; err == nil {
		analytics.AverageRating = avgRating.Average
	}

	// Get appointments by type
	var typeResults []struct {
		Type  string
		Count int
	}
	if err := s.db.Model(&models.Appointment{}).
		Select("type, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ? AND deleted_at IS NULL", request.StartDate, request.EndDate).
		Group("type").
		Scan(&typeResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get appointments by type: %v", err)
	}

	for _, result := range typeResults {
		analytics.AppointmentsByType[result.Type] = result.Count
	}

	// Get appointments by status
	var statusResults []struct {
		Status string
		Count  int
	}
	if err := s.db.Model(&models.Appointment{}).
		Select("status, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ? AND deleted_at IS NULL", request.StartDate, request.EndDate).
		Group("status").
		Scan(&statusResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get appointments by status: %v", err)
	}

	for _, result := range statusResults {
		analytics.AppointmentsByStatus[result.Status] = result.Count
	}

	// Get appointments by user
	var userResults []struct {
		Username string
		Count    int
	}
	if err := s.db.Table("appointments a").
		Select("au.username, COUNT(*) as count").
		Joins("LEFT JOIN admin_users au ON a.assigned_to = au.id").
		Where("a.created_at BETWEEN ? AND ? AND a.deleted_at IS NULL", request.StartDate, request.EndDate).
		Group("au.username").
		Scan(&userResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get appointments by user: %v", err)
	}

	for _, result := range userResults {
		analytics.AppointmentsByUser[result.Username] = result.Count
	}

	// Build type performance metrics
	analytics.TypePerformance = s.buildAppointmentTypeMetrics(typeResults, request)

	// Generate appointment trends
	trends, err := s.generateAppointmentTrends(request.StartDate, request.EndDate, request.Granularity)
	if err != nil {
		logger.Error("Failed to generate appointment trends", err, nil)
		trends = []models.AppointmentTrendData{}
	}

	analytics.AppointmentTrends = trends

	period := fmt.Sprintf("%s to %s", request.StartDate.Format("2006-01-02"), request.EndDate.Format("2006-01-02"))

	return &models.AppointmentMetricsResponse{
		Period:    period,
		Analytics: *analytics,
		Trends:    trends,
	}, nil
}

// GetUserPerformanceAnalytics gets user performance analytics
func (s *AnalyticsService) GetUserPerformanceAnalytics(request *models.AnalyticsRequest) (*models.UserPerformanceResponse, error) {
	var users []models.UserPerformanceAnalytics
	var userIDs []uint

	// Get user IDs to analyze
	if len(request.UserIDs) > 0 {
		userIDs = request.UserIDs
	} else {
		// Get all active users
		if err := s.db.Model(&models.AdminUser{}).
			Where("is_active = ?", true).
			Pluck("id", &userIDs).Error; err != nil {
			return nil, fmt.Errorf("failed to get active users: %v", err)
		}
	}

	for _, userID := range userIDs {
		userMetrics, err := s.calculateUserPerformance(userID, request.StartDate, request.EndDate)
		if err != nil {
			logger.Error("Failed to calculate user performance", err, map[string]interface{}{
				"user_id": userID,
			})
			continue
		}
		users = append(users, *userMetrics)
	}

	// Sort users by performance score
	sort.Slice(users, func(i, j int) bool {
		return users[i].ActivityScore > users[j].ActivityScore
	})

	// Set performance ranks
	for i := range users {
		users[i].PerformanceRank = i + 1
	}

	// Calculate team averages
	teamAverage := s.calculateTeamAverage(users)

	var topPerformer models.UserPerformanceAnalytics
	if len(users) > 0 {
		topPerformer = users[0]
	}

	period := fmt.Sprintf("%s to %s", request.StartDate.Format("2006-01-02"), request.EndDate.Format("2006-01-02"))

	return &models.UserPerformanceResponse{
		Period:      period,
		Users:       users,
		TeamAverage: teamAverage,
		TopPerformer: topPerformer,
	}, nil
}

// GetConversionMetrics gets conversion tracking metrics
func (s *AnalyticsService) GetConversionMetrics(request *models.AnalyticsRequest) (*models.ConversionMetricsResponse, error) {
	// Get overall conversion rate
	var totalContacts, convertedContacts int64
	
	if err := s.db.Model(&models.Contact{}).
		Where("created_at BETWEEN ? AND ?", request.StartDate, request.EndDate).
		Count(&totalContacts).Error; err != nil {
		return nil, fmt.Errorf("failed to get total contacts: %v", err)
	}

	if err := s.db.Model(&models.Contact{}).
		Where("status = ? AND updated_at BETWEEN ? AND ?", "converted", request.StartDate, request.EndDate).
		Count(&convertedContacts).Error; err != nil {
		return nil, fmt.Errorf("failed to get converted contacts: %v", err)
	}

	overallRate := float64(0)
	if totalContacts > 0 {
		overallRate = float64(convertedContacts) / float64(totalContacts) * 100
	}

	// Build conversion funnel
	funnel := s.buildConversionFunnel(request.StartDate, request.EndDate)

	// Get conversion by source
	bySource := s.getConversionBySource(request.StartDate, request.EndDate)

	// Get conversion by user
	byUser := s.getConversionByUser(request.StartDate, request.EndDate)

	// Generate conversion trends
	trends, err := s.generateConversionTrends(request.StartDate, request.EndDate, request.Granularity)
	if err != nil {
		logger.Error("Failed to generate conversion trends", err, nil)
		trends = []models.ConversionTrendData{}
	}

	period := fmt.Sprintf("%s to %s", request.StartDate.Format("2006-01-02"), request.EndDate.Format("2006-01-02"))

	return &models.ConversionMetricsResponse{
		Period:           period,
		OverallRate:      overallRate,
		ConversionFunnel: funnel,
		BySource:         bySource,
		ByUser:           byUser,
		Trends:           trends,
	}, nil
}

// GetResponseTimeMetrics gets response time analytics
func (s *AnalyticsService) GetResponseTimeMetrics(request *models.AnalyticsRequest) (*models.ResponseTimeMetricsResponse, error) {
	// Calculate average response time
	var avgResponse struct {
		Average float64
	}
	
	if err := s.db.Raw(`
		SELECT AVG(TIMESTAMPDIFF(HOUR, created_at, first_response_date)) as average
		FROM contacts 
		WHERE created_at BETWEEN ? AND ? 
		AND first_response_date IS NOT NULL
	`, request.StartDate, request.EndDate).Scan(&avgResponse).Error; err != nil {
		return nil, fmt.Errorf("failed to get average response time: %v", err)
	}

	// Calculate median response time
	var medianResponse float64
	if err := s.db.Raw(`
		SELECT TIMESTAMPDIFF(HOUR, created_at, first_response_date) as response_hours
		FROM contacts 
		WHERE created_at BETWEEN ? AND ? 
		AND first_response_date IS NOT NULL
		ORDER BY response_hours
		LIMIT 1 OFFSET (
			SELECT FLOOR(COUNT(*)/2) 
			FROM contacts 
			WHERE created_at BETWEEN ? AND ? 
			AND first_response_date IS NOT NULL
		)
	`, request.StartDate, request.EndDate, request.StartDate, request.EndDate).
		Scan(&medianResponse).Error; err != nil {
		medianResponse = avgResponse.Average
	}

	// Get response metrics by user
	byUser := s.getResponseTimeByUser(request.StartDate, request.EndDate)

	// Get response metrics by contact type
	byContactType := s.getResponseTimeByType(request.StartDate, request.EndDate)

	// Generate response time trends
	trends := s.generateResponseTimeTrends(request.StartDate, request.EndDate, request.Granularity)

	// Calculate SLA compliance
	slaCompliance := s.calculateSLACompliance(request.StartDate, request.EndDate)

	period := fmt.Sprintf("%s to %s", request.StartDate.Format("2006-01-02"), request.EndDate.Format("2006-01-02"))

	return &models.ResponseTimeMetricsResponse{
		Period:          period,
		AverageResponse: avgResponse.Average,
		MedianResponse:  medianResponse,
		ByUser:          byUser,
		ByContactType:   byContactType,
		Trends:          trends,
		SLACompliance:   slaCompliance,
	}, nil
}

// GetRealtimeMetrics gets real-time dashboard metrics
func (s *AnalyticsService) GetRealtimeMetrics() (*models.RealtimeMetrics, error) {
	metrics := &models.RealtimeMetrics{}

	// Get active users count (logged in within last hour)
	lastHour := time.Now().Add(-1 * time.Hour)
	var activeUsersCount int64
	if err := s.db.Model(&models.AdminUser{}).
		Where("last_login >= ? AND is_active = ?", lastHour, true).
		Count(&activeUsersCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get active users: %v", err)
	}
	metrics.ActiveUsers = int(activeUsersCount)

	// Get online users
	var onlineUsers []models.OnlineUser
	if err := s.db.Model(&models.AdminUser{}).
		Select("id as user_id, username, CONCAT(first_name, ' ', last_name) as full_name, 'online' as status, last_login as last_activity").
		Where("last_login >= ? AND is_active = ?", lastHour, true).
		Scan(&onlineUsers).Error; err == nil {
		metrics.OnlineUsers = onlineUsers
	}

	// Get today's contacts
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)
	var todayContactsCount int64
	if err := s.db.Model(&models.Contact{}).
		Where("created_at BETWEEN ? AND ?", today, tomorrow).
		Count(&todayContactsCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get today's contacts: %v", err)
	}
	metrics.TodayContacts = int(todayContactsCount)

	// Get today's appointments
	var todayAppointmentsCount int64
	if err := s.db.Model(&models.Appointment{}).
		Where("DATE(start_time) = DATE(?) AND deleted_at IS NULL", today).
		Count(&todayAppointmentsCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get today's appointments: %v", err)
	}
	metrics.TodayAppointments = int(todayAppointmentsCount)

	// Get pending follow-ups
	var pendingFollowupsCount int64
	if err := s.db.Model(&models.Contact{}).
		Where("next_followup_date <= ? AND status NOT IN ?", 
			time.Now(), []string{"completed", "cancelled", "converted"}).
		Count(&pendingFollowupsCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get pending follow-ups: %v", err)
	}
	metrics.PendingFollowups = int(pendingFollowupsCount)

	// Get overdue appointments
	var overdueAppointmentsCount int64
	if err := s.db.Model(&models.Appointment{}).
		Where("start_time < ? AND status = ? AND deleted_at IS NULL", 
			time.Now(), models.AppointmentConfirmed).
		Count(&overdueAppointmentsCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get overdue appointments: %v", err)
	}
	metrics.OverdueAppointments = int(overdueAppointmentsCount)

	// Build quick stats
	metrics.QuickStats = s.buildQuickStats()

	return metrics, nil
}

// Helper methods

// buildTopSources builds top source metrics
func (s *AnalyticsService) buildTopSources(sourceResults []struct {
	Source string
	Count  int
}, totalContacts int) []models.SourceMetric {
	var sources []models.SourceMetric

	for _, result := range sourceResults {
		percentage := float64(0)
		if totalContacts > 0 {
			percentage = float64(result.Count) / float64(totalContacts) * 100
		}

		// Get conversion rate for this source
		var convertedCount int64
		s.db.Model(&models.Contact{}).
			Where("source = ? AND status = ?", result.Source, "converted").
			Count(&convertedCount)

		conversionRate := float64(0)
		if result.Count > 0 {
			conversionRate = float64(convertedCount) / float64(result.Count) * 100
		}

		sources = append(sources, models.SourceMetric{
			Source:         result.Source,
			Count:          result.Count,
			ConversionRate: conversionRate,
			Percentage:     percentage,
		})
	}

	// Sort by count descending
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].Count > sources[j].Count
	})

	// Return top 10
	if len(sources) > 10 {
		sources = sources[:10]
	}

	return sources
}

// buildStatusDistribution builds status distribution metrics
func (s *AnalyticsService) buildStatusDistribution(statusResults []struct {
	Status string
	Count  int
}, totalContacts int) []models.StatusDistribution {
	var distribution []models.StatusDistribution

	for _, result := range statusResults {
		percentage := float64(0)
		if totalContacts > 0 {
			percentage = float64(result.Count) / float64(totalContacts) * 100
		}

		distribution = append(distribution, models.StatusDistribution{
			Status:     result.Status,
			Count:      result.Count,
			Percentage: percentage,
		})
	}

	return distribution
}

// generateContactTrends generates contact trend data
func (s *AnalyticsService) generateContactTrends(startDate, endDate time.Time, granularity string) ([]models.ContactTrendData, error) {
	var trends []models.ContactTrendData
	
	// For now, return empty trends - implement based on granularity
	// TODO: Implement trend generation with proper date bucketing
	
	return trends, nil
}

// buildAppointmentTypeMetrics builds appointment type performance metrics
func (s *AnalyticsService) buildAppointmentTypeMetrics(typeResults []struct {
	Type  string
	Count int
}, request *models.AnalyticsRequest) []models.AppointmentTypeMetric {
	var metrics []models.AppointmentTypeMetric

	for _, result := range typeResults {
		// Get completion rate for this type
		var completedCount int64
		s.db.Model(&models.Appointment{}).
			Where("type = ? AND status = ? AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", 
				result.Type, models.AppointmentCompleted, request.StartDate, request.EndDate).
			Count(&completedCount)

		completionRate := float64(0)
		if result.Count > 0 {
			completionRate = float64(completedCount) / float64(result.Count) * 100
		}

		// Get average rating for this type
		var avgRating struct {
			Average float64
		}
		s.db.Model(&models.Appointment{}).
			Select("AVG(rating) as average").
			Where("type = ? AND status = ? AND rating IS NOT NULL AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", 
				result.Type, models.AppointmentCompleted, request.StartDate, request.EndDate).
			Scan(&avgRating)

		// Get average duration for this type
		var avgDuration struct {
			Average int
		}
		s.db.Model(&models.Appointment{}).
			Select("AVG(duration) as average").
			Where("type = ? AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", 
				result.Type, request.StartDate, request.EndDate).
			Scan(&avgDuration)

		metrics = append(metrics, models.AppointmentTypeMetric{
			Type:            result.Type,
			Count:           result.Count,
			CompletionRate:  completionRate,
			AverageRating:   avgRating.Average,
			AverageDuration: avgDuration.Average,
		})
	}

	return metrics
}

// generateAppointmentTrends generates appointment trend data
func (s *AnalyticsService) generateAppointmentTrends(startDate, endDate time.Time, granularity string) ([]models.AppointmentTrendData, error) {
	var trends []models.AppointmentTrendData
	
	// For now, return empty trends - implement based on granularity
	// TODO: Implement trend generation with proper date bucketing
	
	return trends, nil
}

// calculateUserPerformance calculates performance metrics for a specific user
func (s *AnalyticsService) calculateUserPerformance(userID uint, startDate, endDate time.Time) (*models.UserPerformanceAnalytics, error) {
	metrics := &models.UserPerformanceAnalytics{
		UserID:           userID,
		ContactsByStatus: make(map[string]int),
	}

	// Get user info
	var user models.AdminUser
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	metrics.Username = user.Email  // Use email as username
	metrics.FullName = user.Name

	// Get assigned contacts
	var assignedCount int64
	if err := s.db.Model(&models.Contact{}).
		Where("assigned_to = ? AND created_at BETWEEN ? AND ?", userID, startDate, endDate).
		Count(&assignedCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get assigned contacts: %v", err)
	}
	metrics.AssignedContacts = int(assignedCount)

	// Get converted contacts
	var convertedContactsCount int64
	if err := s.db.Model(&models.Contact{}).
		Where("assigned_to = ? AND status = ? AND updated_at BETWEEN ? AND ?", 
			userID, "converted", startDate, endDate).
		Count(&convertedContactsCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get converted contacts: %v", err)
	}
	metrics.ConvertedContacts = int(convertedContactsCount)

	// Calculate conversion rate
	if metrics.AssignedContacts > 0 {
		metrics.ConversionRate = float64(metrics.ConvertedContacts) / float64(metrics.AssignedContacts) * 100
	}

	// Get total appointments
	var totalAppointmentsCount int64
	if err := s.db.Model(&models.Appointment{}).
		Where("assigned_to = ? AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", 
			userID, startDate, endDate).
		Count(&totalAppointmentsCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get total appointments: %v", err)
	}
	metrics.TotalAppointments = int(totalAppointmentsCount)

	// Get completed appointments
	var completedAppointmentsCount int64
	if err := s.db.Model(&models.Appointment{}).
		Where("assigned_to = ? AND status = ? AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", 
			userID, models.AppointmentCompleted, startDate, endDate).
		Count(&completedAppointmentsCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get completed appointments: %v", err)
	}
	metrics.CompletedAppointments = int(completedAppointmentsCount)

	// Calculate completion rate
	if metrics.TotalAppointments > 0 {
		metrics.CompletionRate = float64(metrics.CompletedAppointments) / float64(metrics.TotalAppointments) * 100
	}

	// Get average rating
	var avgRating struct {
		Average float64
	}
	if err := s.db.Model(&models.Appointment{}).
		Select("AVG(rating) as average").
		Where("assigned_to = ? AND status = ? AND rating IS NOT NULL AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", 
			userID, models.AppointmentCompleted, startDate, endDate).
		Scan(&avgRating).Error; err == nil {
		metrics.AverageRating = avgRating.Average
	}

	// Calculate activity score (composite metric)
	metrics.ActivityScore = int(metrics.ConversionRate*0.4 + metrics.CompletionRate*0.3 + metrics.AverageRating*20*0.3)

	return metrics, nil
}

// calculateTeamAverage calculates team average metrics
func (s *AnalyticsService) calculateTeamAverage(users []models.UserPerformanceAnalytics) models.UserPerformanceAnalytics {
	if len(users) == 0 {
		return models.UserPerformanceAnalytics{}
	}

	var totalAssigned, totalConverted, totalAppointments, totalCompleted int
	var totalConversionRate, totalCompletionRate, totalRating float64

	for _, user := range users {
		totalAssigned += user.AssignedContacts
		totalConverted += user.ConvertedContacts
		totalAppointments += user.TotalAppointments
		totalCompleted += user.CompletedAppointments
		totalConversionRate += user.ConversionRate
		totalCompletionRate += user.CompletionRate
		totalRating += user.AverageRating
	}

	count := len(users)
	return models.UserPerformanceAnalytics{
		FullName:              "Team Average",
		AssignedContacts:      totalAssigned / count,
		ConvertedContacts:     totalConverted / count,
		TotalAppointments:     totalAppointments / count,
		CompletedAppointments: totalCompleted / count,
		ConversionRate:        totalConversionRate / float64(count),
		CompletionRate:        totalCompletionRate / float64(count),
		AverageRating:         totalRating / float64(count),
	}
}

// buildConversionFunnel builds conversion funnel data
func (s *AnalyticsService) buildConversionFunnel(startDate, endDate time.Time) models.ConversionFunnelData {
	stages := []models.FunnelStage{
		{Name: "Leads", Count: 0},
		{Name: "Qualified", Count: 0},
		{Name: "Contacted", Count: 0},
		{Name: "Interested", Count: 0},
		{Name: "Converted", Count: 0},
	}

	// Get counts for each stage
	var leadsCount int64
	s.db.Model(&models.Contact{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&leadsCount)
	stages[0].Count = int(leadsCount)

	var qualifiedCount int64
	s.db.Model(&models.Contact{}).
		Where("status = ? AND created_at BETWEEN ? AND ?", "qualified", startDate, endDate).
		Count(&qualifiedCount)
	stages[1].Count = int(qualifiedCount)

	var contactedCount int64
	s.db.Model(&models.Contact{}).
		Where("first_response_date IS NOT NULL AND created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&contactedCount)
	stages[2].Count = int(contactedCount)

	var interestedCount int64
	s.db.Model(&models.Contact{}).
		Where("status = ? AND created_at BETWEEN ? AND ?", "interested", startDate, endDate).
		Count(&interestedCount)
	stages[3].Count = int(interestedCount)

	var convertedFunnelCount int64
	s.db.Model(&models.Contact{}).
		Where("status = ? AND created_at BETWEEN ? AND ?", "converted", startDate, endDate).
		Count(&convertedFunnelCount)
	stages[4].Count = int(convertedFunnelCount)

	// Calculate conversion rates
	for i := 1; i < len(stages); i++ {
		if stages[i-1].Count > 0 {
			stages[i].ConversionRate = float64(stages[i].Count) / float64(stages[i-1].Count) * 100
			stages[i].DropOffRate = 100 - stages[i].ConversionRate
		}
	}

	return models.ConversionFunnelData{Stages: stages}
}

// getConversionBySource gets conversion metrics by source
func (s *AnalyticsService) getConversionBySource(startDate, endDate time.Time) []models.SourceMetric {
	var results []struct {
		Source      string
		Total       int
		Conversions int
	}

	s.db.Raw(`
		SELECT 
			source,
			COUNT(*) as total,
			SUM(CASE WHEN status = 'converted' THEN 1 ELSE 0 END) as conversions
		FROM contacts 
		WHERE created_at BETWEEN ? AND ?
		GROUP BY source
	`, startDate, endDate).Scan(&results)

	var metrics []models.SourceMetric
	for _, result := range results {
		conversionRate := float64(0)
		if result.Total > 0 {
			conversionRate = float64(result.Conversions) / float64(result.Total) * 100
		}

		metrics = append(metrics, models.SourceMetric{
			Source:         result.Source,
			Count:          result.Total,
			ConversionRate: conversionRate,
		})
	}

	return metrics
}

// getConversionByUser gets conversion metrics by user
func (s *AnalyticsService) getConversionByUser(startDate, endDate time.Time) []models.UserConversionMetric {
	var results []struct {
		UserID      uint
		Username    string
		FullName    string
		Total       int
		Conversions int
	}

	s.db.Raw(`
		SELECT 
			c.assigned_to as user_id,
			au.username,
			CONCAT(au.first_name, ' ', au.last_name) as full_name,
			COUNT(*) as total,
			SUM(CASE WHEN c.status = 'converted' THEN 1 ELSE 0 END) as conversions
		FROM contacts c
		LEFT JOIN admin_users au ON c.assigned_to = au.id
		WHERE c.created_at BETWEEN ? AND ?
		GROUP BY c.assigned_to, au.username, au.first_name, au.last_name
	`, startDate, endDate).Scan(&results)

	var metrics []models.UserConversionMetric
	for _, result := range results {
		conversionRate := float64(0)
		if result.Total > 0 {
			conversionRate = float64(result.Conversions) / float64(result.Total) * 100
		}

		metrics = append(metrics, models.UserConversionMetric{
			UserID:         result.UserID,
			Username:       result.Username,
			FullName:       result.FullName,
			TotalContacts:  result.Total,
			Conversions:    result.Conversions,
			ConversionRate: conversionRate,
		})
	}

	return metrics
}

// generateConversionTrends generates conversion trend data
func (s *AnalyticsService) generateConversionTrends(startDate, endDate time.Time, granularity string) ([]models.ConversionTrendData, error) {
	var trends []models.ConversionTrendData
	
	// For now, return empty trends - implement based on granularity
	// TODO: Implement trend generation with proper date bucketing
	
	return trends, nil
}

// getResponseTimeByUser gets response time metrics by user
func (s *AnalyticsService) getResponseTimeByUser(startDate, endDate time.Time) []models.UserResponseMetric {
	var metrics []models.UserResponseMetric
	
	// TODO: Implement response time calculation by user
	
	return metrics
}

// getResponseTimeByType gets response time metrics by contact type
func (s *AnalyticsService) getResponseTimeByType(startDate, endDate time.Time) []models.TypeResponseMetric {
	var metrics []models.TypeResponseMetric
	
	// TODO: Implement response time calculation by type
	
	return metrics
}

// generateResponseTimeTrends generates response time trend data
func (s *AnalyticsService) generateResponseTimeTrends(startDate, endDate time.Time, granularity string) []models.ResponseTrendData {
	var trends []models.ResponseTrendData
	
	// TODO: Implement response time trend generation
	
	return trends
}

// calculateSLACompliance calculates SLA compliance metrics
func (s *AnalyticsService) calculateSLACompliance(startDate, endDate time.Time) models.SLAComplianceMetrics {
	compliance := models.SLAComplianceMetrics{
		ByPriority: make(map[string]float64),
	}
	
	// TODO: Implement SLA compliance calculation
	
	return compliance
}

// buildQuickStats builds quick statistics snapshot
func (s *AnalyticsService) buildQuickStats() models.QuickStatsSnapshot {
	stats := models.QuickStatsSnapshot{}
	
	// Get total contacts
	var totalContactsCount int64
	s.db.Model(&models.Contact{}).Count(&totalContactsCount)
	stats.TotalContacts = int(totalContactsCount)
	
	// Get today's new contacts
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)
	var todayNewContactsCount int64
	s.db.Model(&models.Contact{}).
		Where("created_at BETWEEN ? AND ?", today, tomorrow).
		Count(&todayNewContactsCount)
	stats.TodayNewContacts = int(todayNewContactsCount)
	
	// Calculate week growth
	weekAgo := time.Now().AddDate(0, 0, -7)
	var lastWeekContacts int64
	s.db.Model(&models.Contact{}).
		Where("created_at BETWEEN ? AND ?", weekAgo, today).
		Count(&lastWeekContacts)
	
	var thisWeekContacts int64
	s.db.Model(&models.Contact{}).
		Where("created_at >= ?", weekAgo).
		Count(&thisWeekContacts)
	
	if lastWeekContacts > 0 {
		stats.WeekGrowth = (float64(thisWeekContacts) - float64(lastWeekContacts)) / float64(lastWeekContacts) * 100
	}
	
	// Get conversion rate
	var totalContacts, convertedContacts int64
	s.db.Model(&models.Contact{}).Count(&totalContacts)
	s.db.Model(&models.Contact{}).Where("status = ?", "converted").Count(&convertedContacts)
	
	if totalContacts > 0 {
		stats.ConversionRate = float64(convertedContacts) / float64(totalContacts) * 100
	}
	
	return stats
}