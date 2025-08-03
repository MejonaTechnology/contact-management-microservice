package handlers

import (
	"contact-service/internal/models"
	"contact-service/internal/services"
	"contact-service/pkg/database"
	"contact-service/pkg/logger"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AnalyticsHandler handles analytics and metrics requests
type AnalyticsHandler struct {
	analyticsService *services.AnalyticsService
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler() *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: services.NewAnalyticsService(database.DB),
	}
}

// GetContactAnalytics godoc
// @Summary Get contact analytics
// @Description Get comprehensive contact analytics and metrics
// @Tags analytics
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Param user_ids query string false "Comma-separated user IDs to filter"
// @Param sources query string false "Comma-separated sources to filter"
// @Param granularity query string false "Data granularity (day, week, month)" default(day)
// @Success 200 {object} APIResponse{data=models.ContactMetricsResponse}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /analytics/contacts [get]
func (h *AnalyticsHandler) GetContactAnalytics(c *gin.Context) {
	request, err := h.parseAnalyticsRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request parameters", err.Error()))
		return
	}

	analytics, err := h.analyticsService.GetContactAnalytics(request)
	if err != nil {
		logger.Error("Failed to get contact analytics", err, map[string]interface{}{
			"start_date": request.StartDate,
			"end_date":   request.EndDate,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get contact analytics", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Contact analytics retrieved successfully", analytics))
}

// GetAppointmentAnalytics godoc
// @Summary Get appointment analytics
// @Description Get comprehensive appointment analytics and metrics
// @Tags analytics
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Param user_ids query string false "Comma-separated user IDs to filter"
// @Param granularity query string false "Data granularity (day, week, month)" default(day)
// @Success 200 {object} APIResponse{data=models.AppointmentMetricsResponse}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /analytics/appointments [get]
func (h *AnalyticsHandler) GetAppointmentAnalytics(c *gin.Context) {
	request, err := h.parseAnalyticsRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request parameters", err.Error()))
		return
	}

	analytics, err := h.analyticsService.GetAppointmentAnalytics(request)
	if err != nil {
		logger.Error("Failed to get appointment analytics", err, map[string]interface{}{
			"start_date": request.StartDate,
			"end_date":   request.EndDate,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get appointment analytics", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Appointment analytics retrieved successfully", analytics))
}

// GetUserPerformanceAnalytics godoc
// @Summary Get user performance analytics
// @Description Get comprehensive user performance metrics and analytics
// @Tags analytics
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Param user_ids query string false "Comma-separated user IDs to analyze"
// @Success 200 {object} APIResponse{data=models.UserPerformanceResponse}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /analytics/performance [get]
func (h *AnalyticsHandler) GetUserPerformanceAnalytics(c *gin.Context) {
	request, err := h.parseAnalyticsRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request parameters", err.Error()))
		return
	}

	analytics, err := h.analyticsService.GetUserPerformanceAnalytics(request)
	if err != nil {
		logger.Error("Failed to get user performance analytics", err, map[string]interface{}{
			"start_date": request.StartDate,
			"end_date":   request.EndDate,
			"user_ids":   request.UserIDs,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get user performance analytics", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("User performance analytics retrieved successfully", analytics))
}

// GetConversionMetrics godoc
// @Summary Get conversion metrics
// @Description Get conversion tracking metrics and analytics
// @Tags analytics
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Param sources query string false "Comma-separated sources to filter"
// @Param granularity query string false "Data granularity (day, week, month)" default(day)
// @Success 200 {object} APIResponse{data=models.ConversionMetricsResponse}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /analytics/conversion [get]
func (h *AnalyticsHandler) GetConversionMetrics(c *gin.Context) {
	request, err := h.parseAnalyticsRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request parameters", err.Error()))
		return
	}

	metrics, err := h.analyticsService.GetConversionMetrics(request)
	if err != nil {
		logger.Error("Failed to get conversion metrics", err, map[string]interface{}{
			"start_date": request.StartDate,
			"end_date":   request.EndDate,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get conversion metrics", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Conversion metrics retrieved successfully", metrics))
}

// GetResponseTimeMetrics godoc
// @Summary Get response time metrics
// @Description Get response time analytics and SLA compliance metrics
// @Tags analytics
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Param user_ids query string false "Comma-separated user IDs to filter"
// @Param granularity query string false "Data granularity (day, week, month)" default(day)
// @Success 200 {object} APIResponse{data=models.ResponseTimeMetricsResponse}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /analytics/response-times [get]
func (h *AnalyticsHandler) GetResponseTimeMetrics(c *gin.Context) {
	request, err := h.parseAnalyticsRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request parameters", err.Error()))
		return
	}

	metrics, err := h.analyticsService.GetResponseTimeMetrics(request)
	if err != nil {
		logger.Error("Failed to get response time metrics", err, map[string]interface{}{
			"start_date": request.StartDate,
			"end_date":   request.EndDate,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get response time metrics", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Response time metrics retrieved successfully", metrics))
}

// GetSourceAnalytics godoc
// @Summary Get source analytics
// @Description Get analytics for different contact sources
// @Tags analytics
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Param sources query string false "Comma-separated sources to analyze"
// @Success 200 {object} APIResponse{data=[]models.SourceMetric}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /analytics/sources [get]
func (h *AnalyticsHandler) GetSourceAnalytics(c *gin.Context) {
	request, err := h.parseAnalyticsRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request parameters", err.Error()))
		return
	}

	// Get contact analytics which includes source metrics
	contactAnalytics, err := h.analyticsService.GetContactAnalytics(request)
	if err != nil {
		logger.Error("Failed to get source analytics", err, map[string]interface{}{
			"start_date": request.StartDate,
			"end_date":   request.EndDate,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get source analytics", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Source analytics retrieved successfully", contactAnalytics.Analytics.TopSources))
}

// GetRealtimeMetrics godoc
// @Summary Get realtime metrics
// @Description Get real-time dashboard metrics and live data
// @Tags analytics
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=models.RealtimeMetrics}
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /analytics/realtime [get]
func (h *AnalyticsHandler) GetRealtimeMetrics(c *gin.Context) {
	metrics, err := h.analyticsService.GetRealtimeMetrics()
	if err != nil {
		logger.Error("Failed to get realtime metrics", err, nil)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get realtime metrics", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Realtime metrics retrieved successfully", metrics))
}

// GetDashboardSummary godoc
// @Summary Get dashboard summary
// @Description Get summary metrics for the main dashboard
// @Tags analytics
// @Accept json
// @Produce json
// @Param period query string false "Period for metrics (today, week, month, quarter, year)" default(month)
// @Success 200 {object} APIResponse{data=models.QuickStatsSnapshot}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /analytics/dashboard [get]
func (h *AnalyticsHandler) GetDashboardSummary(c *gin.Context) {
	period := c.Query("period")
	if period == "" {
		period = "month"
	}

	// Calculate date range based on period
	endDate := time.Now()
	var startDate time.Time

	switch period {
	case "today":
		startDate = time.Now().Truncate(24 * time.Hour)
	case "week":
		startDate = time.Now().AddDate(0, 0, -7)
	case "month":
		startDate = time.Now().AddDate(0, -1, 0)
	case "quarter":
		startDate = time.Now().AddDate(0, -3, 0)
	case "year":
		startDate = time.Now().AddDate(-1, 0, 0)
	default:
		startDate = time.Now().AddDate(0, -1, 0) // Default to month
	}

	request := &models.AnalyticsRequest{
		StartDate:   startDate,
		EndDate:     endDate,
		Granularity: "day",
	}

	// Get realtime metrics which includes quick stats
	realtimeMetrics, err := h.analyticsService.GetRealtimeMetrics()
	if err != nil {
		logger.Error("Failed to get dashboard summary", err, map[string]interface{}{
			"period": period,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get dashboard summary", err.Error()))
		return
	}

	// Enhance quick stats with period-specific data
	contactAnalytics, err := h.analyticsService.GetContactAnalytics(request)
	if err == nil {
		realtimeMetrics.QuickStats.ConversionRate = contactAnalytics.Analytics.ConversionRate
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Dashboard summary retrieved successfully", realtimeMetrics.QuickStats))
}

// GetBusinessIntelligence godoc
// @Summary Get business intelligence
// @Description Get high-level business intelligence metrics and KPIs
// @Tags analytics
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} APIResponse{data=models.BusinessIntelligenceResponse}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /analytics/business-intelligence [get]
func (h *AnalyticsHandler) GetBusinessIntelligence(c *gin.Context) {
	request, err := h.parseAnalyticsRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request parameters", err.Error()))
		return
	}

	// Build business intelligence from multiple analytics
	intelligence := &models.BusinessIntelligence{}
	intelligence.Period = fmt.Sprintf("%s to %s", request.StartDate.Format("2006-01-02"), request.EndDate.Format("2006-01-02"))

	// Get contact analytics
	contactAnalytics, err := h.analyticsService.GetContactAnalytics(request)
	if err == nil {
		intelligence.LeadConversionRate = contactAnalytics.Analytics.ConversionRate
		intelligence.AverageLeadValue = 1000.0 // TODO: Calculate from actual deal values
	}

	// Get user performance analytics
	userPerformance, err := h.analyticsService.GetUserPerformanceAnalytics(request)
	if err == nil {
		intelligence.TeamProductivity = float64(userPerformance.TeamAverage.ActivityScore) / 100.0 * 100
		
		// Build top performing users
		topUsers := make([]models.UserPerformanceMetric, 0)
		for _, user := range userPerformance.Users {
			if len(topUsers) < 5 { // Top 5 performers
				topUsers = append(topUsers, models.UserPerformanceMetric{
					UserID:         user.UserID,
					Username:       user.Username,
					FullName:       user.FullName,
					Score:          user.ActivityScore,
					ConversionRate: user.ConversionRate,
					Revenue:        0, // TODO: Calculate from deals
				})
			}
		}
		intelligence.TopPerformingUsers = topUsers
	}

	// Get conversion metrics
	conversionMetrics, err := h.analyticsService.GetConversionMetrics(request)
	if err == nil {
		intelligence.ConversionFunnel = conversionMetrics.ConversionFunnel
		
		// Build revenue by source (placeholder values)
		revenueBySource := make([]models.RevenueSourceMetric, 0)
		for _, source := range conversionMetrics.BySource {
			revenueBySource = append(revenueBySource, models.RevenueSourceMetric{
				Source:   source.Source,
				Revenue:  float64(source.Count) * 1000, // Placeholder calculation
				Count:    source.Count,
				AvgValue: 1000.0,
				Growth:   5.0, // Placeholder
			})
		}
		intelligence.RevenueBySource = revenueBySource
	}

	// Build KPI summary
	intelligence.KPISummary = models.KPISummary{
		TotalLeads:         contactAnalytics.Analytics.TotalContacts,
		QualifiedLeads:     contactAnalytics.Analytics.ActiveContacts,
		ConvertedLeads:     contactAnalytics.Analytics.ConvertedContacts,
		LeadConversionRate: intelligence.LeadConversionRate,
		AverageLeadValue:   intelligence.AverageLeadValue,
		TotalRevenue:       float64(contactAnalytics.Analytics.ConvertedContacts) * intelligence.AverageLeadValue,
		RevenueGrowth:      contactAnalytics.Analytics.GrowthRate,
		TeamProductivity:   intelligence.TeamProductivity,
	}

	response := &models.BusinessIntelligenceResponse{
		Period:       intelligence.Period,
		Intelligence: *intelligence,
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Business intelligence retrieved successfully", response))
}

// GetAnalyticsExport godoc
// @Summary Export analytics data
// @Description Export analytics data in various formats (CSV, Excel, PDF)
// @Tags analytics
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Param format query string true "Export format (csv, excel, pdf)"
// @Param type query string true "Analytics type (contacts, appointments, performance, conversion)"
// @Success 200 {object} APIResponse{data=map[string]interface{}}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /analytics/export [get]
func (h *AnalyticsHandler) GetAnalyticsExport(c *gin.Context) {
	format := c.Query("format")
	analyticsType := c.Query("type")

	if format == "" || analyticsType == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse("format and type parameters are required", ""))
		return
	}

	request, err := h.parseAnalyticsRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request parameters", err.Error()))
		return
	}

	// TODO: Implement actual export functionality
	// For now, return a placeholder response
	exportData := map[string]interface{}{
		"export_id":   "exp_" + strconv.FormatInt(time.Now().Unix(), 10),
		"format":      format,
		"type":        analyticsType,
		"period":      fmt.Sprintf("%s to %s", request.StartDate.Format("2006-01-02"), request.EndDate.Format("2006-01-02")),
		"status":      "processing",
		"download_url": fmt.Sprintf("/api/v1/analytics/export/download?id=exp_%d", time.Now().Unix()),
		"expires_at":  time.Now().Add(24 * time.Hour),
	}

	logger.Info("Analytics export requested", map[string]interface{}{
		"format":         format,
		"analytics_type": analyticsType,
		"start_date":     request.StartDate,
		"end_date":       request.EndDate,
	})

	c.JSON(http.StatusOK, NewSuccessResponse("Export initiated successfully", exportData))
}

// Helper methods

// parseAnalyticsRequest parses and validates analytics request parameters
func (h *AnalyticsHandler) parseAnalyticsRequest(c *gin.Context) (*models.AnalyticsRequest, error) {
	request := &models.AnalyticsRequest{}

	// Parse start date
	startDateStr := c.Query("start_date")
	if startDateStr == "" {
		// Default to 30 days ago
		request.StartDate = time.Now().AddDate(0, 0, -30)
	} else {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date format: %v", err)
		}
		request.StartDate = startDate
	}

	// Parse end date
	endDateStr := c.Query("end_date")
	if endDateStr == "" {
		// Default to today
		request.EndDate = time.Now()
	} else {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date format: %v", err)
		}
		request.EndDate = endDate
	}

	// Validate date range
	if request.EndDate.Before(request.StartDate) {
		return nil, fmt.Errorf("end_date must be after start_date")
	}

	// Parse user IDs
	userIDsStr := c.Query("user_ids")
	if userIDsStr != "" {
		userIDStrings := strings.Split(userIDsStr, ",")
		for _, userIDStr := range userIDStrings {
			userID, err := strconv.ParseUint(strings.TrimSpace(userIDStr), 10, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid user_id: %s", userIDStr)
			}
			request.UserIDs = append(request.UserIDs, uint(userID))
		}
	}

	// Parse sources
	sourcesStr := c.Query("sources")
	if sourcesStr != "" {
		request.Sources = strings.Split(sourcesStr, ",")
		for i, source := range request.Sources {
			request.Sources[i] = strings.TrimSpace(source)
		}
	}

	// Parse contact types
	contactTypesStr := c.Query("contact_types")
	if contactTypesStr != "" {
		request.ContactTypes = strings.Split(contactTypesStr, ",")
		for i, contactType := range request.ContactTypes {
			request.ContactTypes[i] = strings.TrimSpace(contactType)
		}
	}

	// Parse statuses
	statusesStr := c.Query("statuses")
	if statusesStr != "" {
		request.Statuses = strings.Split(statusesStr, ",")
		for i, status := range request.Statuses {
			request.Statuses[i] = strings.TrimSpace(status)
		}
	}

	// Parse granularity
	request.Granularity = c.Query("granularity")
	if request.Granularity == "" {
		request.Granularity = "day"
	}

	// Validate granularity
	validGranularities := []string{"day", "week", "month", "quarter", "year"}
	isValidGranularity := false
	for _, valid := range validGranularities {
		if request.Granularity == valid {
			isValidGranularity = true
			break
		}
	}
	if !isValidGranularity {
		return nil, fmt.Errorf("invalid granularity: must be one of %v", validGranularities)
	}

	// Parse metric types
	metricTypesStr := c.Query("metric_types")
	if metricTypesStr != "" {
		request.MetricTypes = strings.Split(metricTypesStr, ",")
		for i, metricType := range request.MetricTypes {
			request.MetricTypes[i] = strings.TrimSpace(metricType)
		}
	}

	return request, nil
}

