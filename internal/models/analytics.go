package models

import (
	"time"
)

// Analytics and Metrics Models

// ContactAnalytics represents contact-related analytics
type ContactAnalytics struct {
	TotalContacts      int                    `json:"total_contacts"`
	NewContacts        int                    `json:"new_contacts"`
	ActiveContacts     int                    `json:"active_contacts"`
	ConvertedContacts  int                    `json:"converted_contacts"`
	ContactsByStatus   map[string]int         `json:"contacts_by_status"`
	ContactsBySource   map[string]int         `json:"contacts_by_source"`
	ContactsByType     map[string]int         `json:"contacts_by_type"`
	ContactsByAssignee map[string]int         `json:"contacts_by_assignee"`
	GrowthRate         float64                `json:"growth_rate"`
	ConversionRate     float64                `json:"conversion_rate"`
	TopSources         []SourceMetric         `json:"top_sources"`
	StatusDistribution []StatusDistribution   `json:"status_distribution"`
}

// AppointmentAnalytics represents appointment-related analytics
type AppointmentAnalytics struct {
	TotalAppointments     int                         `json:"total_appointments"`
	CompletedAppointments int                         `json:"completed_appointments"`
	CancelledAppointments int                         `json:"cancelled_appointments"`
	UpcomingAppointments  int                         `json:"upcoming_appointments"`
	AppointmentsByType    map[string]int              `json:"appointments_by_type"`
	AppointmentsByStatus  map[string]int              `json:"appointments_by_status"`
	AppointmentsByUser    map[string]int              `json:"appointments_by_user"`
	CompletionRate        float64                     `json:"completion_rate"`
	CancellationRate      float64                     `json:"cancellation_rate"`
	AverageRating         float64                     `json:"average_rating"`
	AppointmentTrends     []AppointmentTrendData      `json:"appointment_trends"`
	TypePerformance       []AppointmentTypeMetric     `json:"type_performance"`
}

// UserPerformanceAnalytics represents user performance metrics
type UserPerformanceAnalytics struct {
	UserID              uint                    `json:"user_id"`
	Username            string                  `json:"username"`
	FullName            string                  `json:"full_name"`
	AssignedContacts    int                     `json:"assigned_contacts"`
	ConvertedContacts   int                     `json:"converted_contacts"`
	TotalAppointments   int                     `json:"total_appointments"`
	CompletedAppointments int                   `json:"completed_appointments"`
	ConversionRate      float64                 `json:"conversion_rate"`
	CompletionRate      float64                 `json:"completion_rate"`
	AverageRating       float64                 `json:"average_rating"`
	ResponseTime        float64                 `json:"response_time"` // Average in hours
	ActivityScore       int                     `json:"activity_score"`
	PerformanceRank     int                     `json:"performance_rank"`
	ContactsByStatus    map[string]int          `json:"contacts_by_status"`
	MonthlyTrends       []MonthlyPerformance    `json:"monthly_trends"`
}

// BusinessIntelligence represents high-level business metrics
type BusinessIntelligence struct {
	Period              string                  `json:"period"`
	TotalRevenue        float64                 `json:"total_revenue"`
	PotentialRevenue    float64                 `json:"potential_revenue"`
	RevenueGrowth       float64                 `json:"revenue_growth"`
	LeadConversionRate  float64                 `json:"lead_conversion_rate"`
	CustomerRetention   float64                 `json:"customer_retention"`
	AverageLeadValue    float64                 `json:"average_lead_value"`
	SalesVelocity       float64                 `json:"sales_velocity"`
	TeamProductivity    float64                 `json:"team_productivity"`
	TopPerformingUsers  []UserPerformanceMetric `json:"top_performing_users"`
	RevenueBySource     []RevenueSourceMetric   `json:"revenue_by_source"`
	ConversionFunnel    ConversionFunnelData    `json:"conversion_funnel"`
	KPISummary          KPISummary              `json:"kpi_summary"`
}

// Supporting metric structures

// SourceMetric represents metrics for a contact source
type SourceMetric struct {
	Source          string  `json:"source"`
	Count           int     `json:"count"`
	ConversionRate  float64 `json:"conversion_rate"`
	AverageValue    float64 `json:"average_value"`
	Percentage      float64 `json:"percentage"`
}

// StatusDistribution represents contact status distribution
type StatusDistribution struct {
	Status     string  `json:"status"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
	Change     float64 `json:"change"` // Percentage change from previous period
}

// AppointmentTrendData represents appointment trends over time
type AppointmentTrendData struct {
	Date              time.Time `json:"date"`
	TotalAppointments int       `json:"total_appointments"`
	CompletedCount    int       `json:"completed_count"`
	CancelledCount    int       `json:"cancelled_count"`
	CompletionRate    float64   `json:"completion_rate"`
}

// AppointmentTypeMetric represents performance metrics by appointment type
type AppointmentTypeMetric struct {
	Type           string  `json:"type"`
	Count          int     `json:"count"`
	CompletionRate float64 `json:"completion_rate"`
	AverageRating  float64 `json:"average_rating"`
	AverageDuration int    `json:"average_duration"`
	ConversionRate float64 `json:"conversion_rate"`
}

// MonthlyPerformance represents monthly performance data for users
type MonthlyPerformance struct {
	Month             time.Time `json:"month"`
	ContactsHandled   int       `json:"contacts_handled"`
	ContactsConverted int       `json:"contacts_converted"`
	AppointmentsHeld  int       `json:"appointments_held"`
	ConversionRate    float64   `json:"conversion_rate"`
	Revenue           float64   `json:"revenue"`
}

// UserPerformanceMetric represents condensed user performance metrics
type UserPerformanceMetric struct {
	UserID         uint    `json:"user_id"`
	Username       string  `json:"username"`
	FullName       string  `json:"full_name"`
	Score          int     `json:"score"`
	ConversionRate float64 `json:"conversion_rate"`
	Revenue        float64 `json:"revenue"`
}

// RevenueSourceMetric represents revenue metrics by source
type RevenueSourceMetric struct {
	Source    string  `json:"source"`
	Revenue   float64 `json:"revenue"`
	Count     int     `json:"count"`
	AvgValue  float64 `json:"avg_value"`
	Growth    float64 `json:"growth"`
}

// ConversionFunnelData represents the conversion funnel metrics
type ConversionFunnelData struct {
	Stages []FunnelStage `json:"stages"`
}

// FunnelStage represents a stage in the conversion funnel
type FunnelStage struct {
	Name             string  `json:"name"`
	Count            int     `json:"count"`
	ConversionRate   float64 `json:"conversion_rate"`
	DropOffRate      float64 `json:"drop_off_rate"`
	AverageTimeInStage float64 `json:"average_time_in_stage"` // Days
}

// KPISummary represents key performance indicators
type KPISummary struct {
	TotalLeads           int     `json:"total_leads"`
	QualifiedLeads       int     `json:"qualified_leads"`
	ConvertedLeads       int     `json:"converted_leads"`
	LeadConversionRate   float64 `json:"lead_conversion_rate"`
	AverageLeadValue     float64 `json:"average_lead_value"`
	TotalRevenue         float64 `json:"total_revenue"`
	RevenueGrowth        float64 `json:"revenue_growth"`
	AverageResponseTime  float64 `json:"average_response_time"`
	TeamProductivity     float64 `json:"team_productivity"`
	CustomerSatisfaction float64 `json:"customer_satisfaction"`
}

// Request/Response types for analytics

// AnalyticsRequest represents a request for analytics data
type AnalyticsRequest struct {
	StartDate    time.Time `json:"start_date" binding:"required"`
	EndDate      time.Time `json:"end_date" binding:"required"`
	UserIDs      []uint    `json:"user_ids"`
	Sources      []string  `json:"sources"`
	ContactTypes []string  `json:"contact_types"`
	Statuses     []string  `json:"statuses"`
	Granularity  string    `json:"granularity"` // day, week, month, quarter, year
	MetricTypes  []string  `json:"metric_types"` // contacts, appointments, revenue, performance
}

// ContactMetricsResponse represents contact metrics response
type ContactMetricsResponse struct {
	Period    string            `json:"period"`
	Analytics ContactAnalytics  `json:"analytics"`
	Trends    []ContactTrendData `json:"trends"`
}

// ContactTrendData represents contact trends over time
type ContactTrendData struct {
	Date         time.Time `json:"date"`
	NewContacts  int       `json:"new_contacts"`
	TotalContacts int      `json:"total_contacts"`
	Conversions  int       `json:"conversions"`
	GrowthRate   float64   `json:"growth_rate"`
}

// AppointmentMetricsResponse represents appointment metrics response
type AppointmentMetricsResponse struct {
	Period    string               `json:"period"`
	Analytics AppointmentAnalytics `json:"analytics"`
	Trends    []AppointmentTrendData `json:"trends"`
}

// UserPerformanceResponse represents user performance metrics response
type UserPerformanceResponse struct {
	Period      string                      `json:"period"`
	Users       []UserPerformanceAnalytics  `json:"users"`
	TeamAverage UserPerformanceAnalytics    `json:"team_average"`
	TopPerformer UserPerformanceAnalytics   `json:"top_performer"`
}

// BusinessIntelligenceResponse represents business intelligence response
type BusinessIntelligenceResponse struct {
	Period       string               `json:"period"`
	Intelligence BusinessIntelligence `json:"intelligence"`
}

// ConversionMetricsResponse represents conversion tracking metrics
type ConversionMetricsResponse struct {
	Period           string               `json:"period"`
	OverallRate      float64              `json:"overall_rate"`
	ConversionFunnel ConversionFunnelData `json:"conversion_funnel"`
	BySource         []SourceMetric       `json:"by_source"`
	ByUser           []UserConversionMetric `json:"by_user"`
	Trends           []ConversionTrendData `json:"trends"`
}

// UserConversionMetric represents conversion metrics by user
type UserConversionMetric struct {
	UserID         uint    `json:"user_id"`
	Username       string  `json:"username"`
	FullName       string  `json:"full_name"`
	TotalContacts  int     `json:"total_contacts"`
	Conversions    int     `json:"conversions"`
	ConversionRate float64 `json:"conversion_rate"`
	Revenue        float64 `json:"revenue"`
}

// ConversionTrendData represents conversion trends over time
type ConversionTrendData struct {
	Date           time.Time `json:"date"`
	TotalContacts  int       `json:"total_contacts"`
	Conversions    int       `json:"conversions"`
	ConversionRate float64   `json:"conversion_rate"`
	Revenue        float64   `json:"revenue"`
}

// ResponseTimeMetricsResponse represents response time analytics
type ResponseTimeMetricsResponse struct {
	Period          string                 `json:"period"`
	AverageResponse float64                `json:"average_response"` // Hours
	MedianResponse  float64                `json:"median_response"`  // Hours
	ByUser          []UserResponseMetric   `json:"by_user"`
	ByContactType   []TypeResponseMetric   `json:"by_contact_type"`
	Trends          []ResponseTrendData    `json:"trends"`
	SLACompliance   SLAComplianceMetrics   `json:"sla_compliance"`
}

// UserResponseMetric represents response time metrics by user
type UserResponseMetric struct {
	UserID          uint    `json:"user_id"`
	Username        string  `json:"username"`
	FullName        string  `json:"full_name"`
	AverageResponse float64 `json:"average_response"`
	MedianResponse  float64 `json:"median_response"`
	TotalContacts   int     `json:"total_contacts"`
	Within1Hour     int     `json:"within_1_hour"`
	Within4Hours    int     `json:"within_4_hours"`
	Within24Hours   int     `json:"within_24_hours"`
	SLACompliance   float64 `json:"sla_compliance"`
}

// TypeResponseMetric represents response time metrics by contact type
type TypeResponseMetric struct {
	ContactType     string  `json:"contact_type"`
	AverageResponse float64 `json:"average_response"`
	MedianResponse  float64 `json:"median_response"`
	Count           int     `json:"count"`
	SLACompliance   float64 `json:"sla_compliance"`
}

// ResponseTrendData represents response time trends over time
type ResponseTrendData struct {
	Date            time.Time `json:"date"`
	AverageResponse float64   `json:"average_response"`
	MedianResponse  float64   `json:"median_response"`
	SLACompliance   float64   `json:"sla_compliance"`
	Volume          int       `json:"volume"`
}

// SLAComplianceMetrics represents SLA compliance metrics
type SLAComplianceMetrics struct {
	Overall      float64                   `json:"overall"`
	Within1Hour  float64                   `json:"within_1_hour"`
	Within4Hours float64                   `json:"within_4_hours"`
	Within24Hours float64                  `json:"within_24_hours"`
	ByPriority   map[string]float64        `json:"by_priority"`
	ByUser       []UserSLACompliance       `json:"by_user"`
}

// UserSLACompliance represents SLA compliance by user
type UserSLACompliance struct {
	UserID     uint    `json:"user_id"`
	Username   string  `json:"username"`
	FullName   string  `json:"full_name"`
	Compliance float64 `json:"compliance"`
	Total      int     `json:"total"`
	Met        int     `json:"met"`
}

// RealtimeMetrics represents real-time dashboard metrics
type RealtimeMetrics struct {
	ActiveUsers        int                    `json:"active_users"`
	OnlineUsers        []OnlineUser           `json:"online_users"`
	TodayContacts      int                    `json:"today_contacts"`
	TodayAppointments  int                    `json:"today_appointments"`
	PendingFollowups   int                    `json:"pending_followups"`
	OverdueAppointments int                   `json:"overdue_appointments"`
	RecentActivity     []ActivityFeedItem     `json:"recent_activity"`
	AlertsAndNotifications []SystemAlert       `json:"alerts_and_notifications"`
	QuickStats         QuickStatsSnapshot     `json:"quick_stats"`
}

// OnlineUser represents an online user
type OnlineUser struct {
	UserID       uint      `json:"user_id"`
	Username     string    `json:"username"`
	FullName     string    `json:"full_name"`
	Status       string    `json:"status"`
	LastActivity time.Time `json:"last_activity"`
	ActiveTasks  int       `json:"active_tasks"`
}

// ActivityFeedItem represents an activity feed item
type ActivityFeedItem struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	Username    string    `json:"username"`
	Action      string    `json:"action"`
	EntityType  string    `json:"entity_type"`
	EntityID    uint      `json:"entity_id"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
	Metadata    JSONMap   `json:"metadata"`
}


// QuickStatsSnapshot represents quick statistics snapshot
type QuickStatsSnapshot struct {
	TotalContacts      int     `json:"total_contacts"`
	TodayNewContacts   int     `json:"today_new_contacts"`
	WeekGrowth         float64 `json:"week_growth"`
	ConversionRate     float64 `json:"conversion_rate"`
	AverageResponseTime float64 `json:"average_response_time"`
	ActiveDeals        int     `json:"active_deals"`
	TotalRevenue       float64 `json:"total_revenue"`
	MonthlyTarget      float64 `json:"monthly_target"`
	TargetProgress     float64 `json:"target_progress"`
}

// Database models for analytics tables

// ActivityLog represents user activity logging
type ActivityLog struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       *uint     `json:"user_id" gorm:"index"`
	Action       string    `json:"action" gorm:"size:100;not null;index"`
	EntityType   string    `json:"entity_type" gorm:"size:50;not null;index"`
	EntityID     *uint     `json:"entity_id" gorm:"index"`
	Description  string    `json:"description" gorm:"type:text;not null"`
	IPAddress    *string   `json:"ip_address" gorm:"size:45"`
	UserAgent    *string   `json:"user_agent" gorm:"type:text"`
	RequestMethod *string  `json:"request_method" gorm:"size:10"`
	RequestURL   *string   `json:"request_url" gorm:"size:500"`
	Metadata     JSONMap   `json:"metadata" gorm:"type:json"`
	ExecutionTime *int     `json:"execution_time"` // Milliseconds
	ResponseStatus *int    `json:"response_status"`
	CreatedAt    time.Time `json:"created_at" gorm:"index"`
	
	// Relationships
	User *AdminUser `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for ActivityLog
func (ActivityLog) TableName() string {
	return "activity_logs"
}

// SystemAlert represents system alerts and notifications
type SystemAlert struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	Type        string     `json:"type" gorm:"size:50;not null;index"` // info, warning, error, critical
	Priority    string     `json:"priority" gorm:"size:20;not null;default:medium;index"` // low, medium, high, urgent
	Title       string     `json:"title" gorm:"size:255;not null"`
	Message     string     `json:"message" gorm:"type:text;not null"`
	UserID      *uint      `json:"user_id" gorm:"index"` // NULL for global alerts
	Role        *string    `json:"role" gorm:"size:50"` // NULL for user-specific alerts
	IsRead      bool       `json:"is_read" gorm:"default:false;index"`
	IsDismissed bool       `json:"is_dismissed" gorm:"default:false"`
	IsActive    bool       `json:"is_active" gorm:"default:true;index"`
	ActionURL   *string    `json:"action_url" gorm:"size:500"`
	ActionLabel *string    `json:"action_label" gorm:"size:100"`
	ExpiresAt   *time.Time `json:"expires_at" gorm:"index"`
	CreatedAt   time.Time  `json:"created_at" gorm:"index"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CreatedBy   *uint      `json:"created_by"`
	
	// Relationships
	User      *AdminUser `json:"user,omitempty" gorm:"foreignKey:UserID"`
	CreatedByUser *AdminUser `json:"created_by_user,omitempty" gorm:"foreignKey:CreatedBy"`
}

// TableName specifies the table name for SystemAlert
func (SystemAlert) TableName() string {
	return "system_alerts"
}

// AnalyticsCache represents cached analytics data
type AnalyticsCache struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	CacheKey        string    `json:"cache_key" gorm:"size:255;not null;unique;index"`
	MetricType      string    `json:"metric_type" gorm:"size:50;not null;index"`
	StartDate       time.Time `json:"start_date" gorm:"not null;type:date;index:idx_dates"`
	EndDate         time.Time `json:"end_date" gorm:"not null;type:date;index:idx_dates"`
	Granularity     string    `json:"granularity" gorm:"size:20;not null"`
	Filters         JSONMap   `json:"filters" gorm:"type:json"`
	Data            JSONMap   `json:"data" gorm:"type:json;not null"`
	CalculationTime *int      `json:"calculation_time"` // Milliseconds
	RecordCount     *int      `json:"record_count"`
	CreatedAt       time.Time `json:"created_at"`
	ExpiresAt       time.Time `json:"expires_at" gorm:"not null;index"`
	AccessCount     int       `json:"access_count" gorm:"default:0"`
	LastAccessed    time.Time `json:"last_accessed" gorm:"default:CURRENT_TIMESTAMP;index"`
}

// TableName specifies the table name for AnalyticsCache
func (AnalyticsCache) TableName() string {
	return "analytics_cache"
}

// UserSession represents user session tracking
type UserSession struct {
	ID               uint       `json:"id" gorm:"primaryKey"`
	UserID           uint       `json:"user_id" gorm:"not null;index"`
	SessionToken     string     `json:"session_token" gorm:"size:255;not null;unique;index"`
	DeviceInfo       *string    `json:"device_info" gorm:"type:text"`
	BrowserInfo      *string    `json:"browser_info" gorm:"type:text"`
	IPAddress        *string    `json:"ip_address" gorm:"size:45"`
	LocationInfo     JSONMap    `json:"location_info" gorm:"type:json"`
	LoginAt          time.Time  `json:"login_at" gorm:"default:CURRENT_TIMESTAMP"`
	LastActivity     time.Time  `json:"last_activity" gorm:"default:CURRENT_TIMESTAMP;index"`
	LogoutAt         *time.Time `json:"logout_at"`
	IsActive         bool       `json:"is_active" gorm:"default:true;index"`
	PageViews        int        `json:"page_views" gorm:"default:0"`
	ActionsPerformed int        `json:"actions_performed" gorm:"default:0"`
	SessionDuration  *int       `json:"session_duration"` // Seconds
	
	// Relationships
	User *AdminUser `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for UserSession
func (UserSession) TableName() string {
	return "user_sessions"
}

// PerformanceMetric represents system performance metrics
type PerformanceMetric struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	MetricName string    `json:"metric_name" gorm:"size:100;not null;index"`
	MetricType string    `json:"metric_type" gorm:"size:50;not null;index"` // response_time, query_time, memory_usage
	EntityType *string   `json:"entity_type" gorm:"size:50;index:idx_entity"`
	EntityID   *uint     `json:"entity_id" gorm:"index:idx_entity"`
	Value      float64   `json:"value" gorm:"type:decimal(10,3);not null"`
	Unit       string    `json:"unit" gorm:"size:20;not null"` // ms, seconds, mb, count, percentage
	UserID     *uint     `json:"user_id" gorm:"index"`
	RequestID  *string   `json:"request_id" gorm:"size:100"`
	Endpoint   *string   `json:"endpoint" gorm:"size:200;index"`
	Metadata   JSONMap   `json:"metadata" gorm:"type:json"`
	Tags       JSONArray `json:"tags" gorm:"type:json"`
	RecordedAt time.Time `json:"recorded_at" gorm:"default:CURRENT_TIMESTAMP;index"`
	
	// Relationships
	User *AdminUser `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for PerformanceMetric
func (PerformanceMetric) TableName() string {
	return "performance_metrics"
}

// BusinessMetric represents business KPIs and metrics
type BusinessMetric struct {
	ID                 uint       `json:"id" gorm:"primaryKey"`
	MetricName         string     `json:"metric_name" gorm:"size:100;not null;index"`
	MetricCategory     string     `json:"metric_category" gorm:"size:50;not null;index"` // revenue, conversion, productivity, satisfaction
	PeriodType         string     `json:"period_type" gorm:"size:20;not null;index:idx_period"` // daily, weekly, monthly, quarterly, yearly
	PeriodDate         time.Time  `json:"period_date" gorm:"not null;type:date;index:idx_period"`
	Value              float64    `json:"value" gorm:"type:decimal(15,2);not null"`
	TargetValue        *float64   `json:"target_value" gorm:"type:decimal(15,2)"`
	PreviousValue      *float64   `json:"previous_value" gorm:"type:decimal(15,2)"`
	ChangeAmount       *float64   `json:"change_amount" gorm:"type:decimal(15,2)"`
	ChangePercentage   *float64   `json:"change_percentage" gorm:"type:decimal(5,2)"`
	IsTargetMet        bool       `json:"is_target_met" gorm:"default:false"`
	Department         *string    `json:"department" gorm:"size:50;index"`
	UserID             *uint      `json:"user_id" gorm:"index"`
	Source             *string    `json:"source" gorm:"size:50"`
	CalculationMethod  *string    `json:"calculation_method" gorm:"type:text"`
	DataSources        JSONArray  `json:"data_sources" gorm:"type:json"`
	Notes              *string    `json:"notes" gorm:"type:text"`
	CalculatedAt       time.Time  `json:"calculated_at" gorm:"default:CURRENT_TIMESTAMP"`
	CalculatedBy       *uint      `json:"calculated_by"`
	
	// Relationships
	User           *AdminUser `json:"user,omitempty" gorm:"foreignKey:UserID"`
	CalculatedByUser *AdminUser `json:"calculated_by_user,omitempty" gorm:"foreignKey:CalculatedBy"`
}

// TableName specifies the table name for BusinessMetric
func (BusinessMetric) TableName() string {
	return "business_metrics"
}

// ExportJob represents analytics export job tracking
type ExportJob struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	ExportID     string     `json:"export_id" gorm:"size:50;not null;unique;index"`
	JobType      string     `json:"job_type" gorm:"size:50;not null;index"` // analytics, contacts, appointments, reports
	ExportFormat string     `json:"export_format" gorm:"size:20;not null"` // csv, excel, pdf, json
	StartDate    *time.Time `json:"start_date" gorm:"type:date"`
	EndDate      *time.Time `json:"end_date" gorm:"type:date"`
	Filters      JSONMap    `json:"filters" gorm:"type:json"`
	UserIDs      JSONArray  `json:"user_ids" gorm:"type:json"`
	Status       string     `json:"status" gorm:"size:20;default:pending;index"` // pending, processing, completed, failed
	Progress     int        `json:"progress" gorm:"default:0"` // 0-100 percentage
	FilePath     *string    `json:"file_path" gorm:"size:500"`
	FileSize     *int       `json:"file_size"` // Bytes
	RecordCount  *int       `json:"record_count"`
	DownloadURL  *string    `json:"download_url" gorm:"size:500"`
	ErrorMessage *string    `json:"error_message" gorm:"type:text"`
	RetryCount   int        `json:"retry_count" gorm:"default:0"`
	MaxRetries   int        `json:"max_retries" gorm:"default:3"`
	CreatedAt    time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP;index"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	ExpiresAt    *time.Time `json:"expires_at" gorm:"index"`
	CreatedBy    uint       `json:"created_by" gorm:"not null;index"`
	DownloadedAt *time.Time `json:"downloaded_at"`
	DownloadCount int       `json:"download_count" gorm:"default:0"`
	
	// Relationships
	CreatedByUser *AdminUser `json:"created_by_user,omitempty" gorm:"foreignKey:CreatedBy"`
}

// TableName specifies the table name for ExportJob
func (ExportJob) TableName() string {
	return "export_jobs"
}