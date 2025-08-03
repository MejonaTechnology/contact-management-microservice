package handlers

import (
	"contact-service/internal/models"
	"contact-service/internal/services"
	"contact-service/pkg/database"
	"contact-service/pkg/logger"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SchedulingHandler handles appointment scheduling requests
type SchedulingHandler struct {
	schedulingService *services.SchedulingService
}

// NewSchedulingHandler creates a new scheduling handler
func NewSchedulingHandler() *SchedulingHandler {
	return &SchedulingHandler{
		schedulingService: services.NewSchedulingService(database.DB),
	}
}

// CreateAppointment godoc
// @Summary Create appointment
// @Description Create a new appointment
// @Tags appointments
// @Accept json
// @Produce json
// @Param appointment body models.AppointmentRequest true "Appointment data"
// @Success 201 {object} APIResponse{data=models.AppointmentResponse}
// @Failure 400 {object} APIResponse
// @Failure 409 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /appointments [post]
func (h *SchedulingHandler) CreateAppointment(c *gin.Context) {
	var req models.AppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request data", err.Error()))
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	appointment, err := h.schedulingService.CreateAppointment(&req, *userID)
	if err != nil {
		logger.Error("Failed to create appointment", err, map[string]interface{}{
			"contact_id":  req.ContactID,
			"assigned_to": req.AssignedTo,
			"scheduled_date": req.ScheduledDate,
			"scheduled_time": req.ScheduledTime,
			"user_id":     *userID,
		})
		
		// Check for conflict errors
		if strings.Contains(err.Error(), "conflict") {
			c.JSON(http.StatusConflict, NewErrorResponse("Scheduling conflict", err.Error()))
			return
		}
		
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to create appointment", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, NewSuccessResponse("Appointment created successfully", appointment))
}

// GetAppointment godoc
// @Summary Get appointment
// @Description Get a specific appointment by ID
// @Tags appointments
// @Accept json
// @Produce json
// @Param id path int true "Appointment ID"
// @Success 200 {object} APIResponse{data=models.AppointmentResponse}
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /appointments/{id} [get]
func (h *SchedulingHandler) GetAppointment(c *gin.Context) {
	appointmentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid appointment ID", err.Error()))
		return
	}

	appointment, err := h.schedulingService.GetAppointment(uint(appointmentID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, NewErrorResponse("Appointment not found", ""))
			return
		}
		
		logger.Error("Failed to get appointment", err, map[string]interface{}{
			"appointment_id": appointmentID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get appointment", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Appointment retrieved successfully", appointment))
}

// UpdateAppointment godoc
// @Summary Update appointment
// @Description Update an existing appointment
// @Tags appointments
// @Accept json
// @Produce json
// @Param id path int true "Appointment ID"
// @Param appointment body models.AppointmentRequest true "Updated appointment data"
// @Success 200 {object} APIResponse{data=models.AppointmentResponse}
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 409 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /appointments/{id} [put]
func (h *SchedulingHandler) UpdateAppointment(c *gin.Context) {
	appointmentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid appointment ID", err.Error()))
		return
	}

	var req models.AppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request data", err.Error()))
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	appointment, err := h.schedulingService.UpdateAppointment(uint(appointmentID), &req, *userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, NewErrorResponse("Appointment not found", ""))
			return
		}
		if strings.Contains(err.Error(), "conflict") {
			c.JSON(http.StatusConflict, NewErrorResponse("Scheduling conflict", err.Error()))
			return
		}
		
		logger.Error("Failed to update appointment", err, map[string]interface{}{
			"appointment_id": appointmentID,
			"user_id":        *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to update appointment", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Appointment updated successfully", appointment))
}

// UpdateAppointmentStatus godoc
// @Summary Update appointment status
// @Description Update the status of an appointment (confirm, complete, cancel, etc.)
// @Tags appointments
// @Accept json
// @Produce json
// @Param id path int true "Appointment ID"
// @Param status body models.AppointmentUpdateRequest true "Status update data"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /appointments/{id}/status [put]
func (h *SchedulingHandler) UpdateAppointmentStatus(c *gin.Context) {
	appointmentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid appointment ID", err.Error()))
		return
	}

	var req models.AppointmentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request data", err.Error()))
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	err = h.schedulingService.UpdateAppointmentStatus(uint(appointmentID), &req, *userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, NewErrorResponse("Appointment not found", ""))
			return
		}
		
		logger.Error("Failed to update appointment status", err, map[string]interface{}{
			"appointment_id": appointmentID,
			"new_status":     req.Status,
			"user_id":        *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to update appointment status", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Appointment status updated successfully", nil))
}

// RescheduleAppointment godoc
// @Summary Reschedule appointment
// @Description Reschedule an appointment to a new time
// @Tags appointments
// @Accept json
// @Produce json
// @Param id path int true "Appointment ID"
// @Param reschedule body models.RescheduleRequest true "Reschedule data"
// @Success 200 {object} APIResponse{data=models.AppointmentResponse}
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 409 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /appointments/{id}/reschedule [post]
func (h *SchedulingHandler) RescheduleAppointment(c *gin.Context) {
	appointmentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid appointment ID", err.Error()))
		return
	}

	var req models.RescheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request data", err.Error()))
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	appointment, err := h.schedulingService.RescheduleAppointment(uint(appointmentID), &req, *userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, NewErrorResponse("Appointment not found", ""))
			return
		}
		if strings.Contains(err.Error(), "conflict") {
			c.JSON(http.StatusConflict, NewErrorResponse("Scheduling conflict", err.Error()))
			return
		}
		
		logger.Error("Failed to reschedule appointment", err, map[string]interface{}{
			"appointment_id":   appointmentID,
			"new_start_time":   req.NewStartTime,
			"user_id":          *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to reschedule appointment", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Appointment rescheduled successfully", appointment))
}

// CancelAppointment godoc
// @Summary Cancel appointment
// @Description Cancel an appointment
// @Tags appointments
// @Accept json
// @Produce json
// @Param id path int true "Appointment ID"
// @Param cancel body map[string]string true "Cancellation data"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /appointments/{id}/cancel [post]
func (h *SchedulingHandler) CancelAppointment(c *gin.Context) {
	appointmentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid appointment ID", err.Error()))
		return
	}

	var requestData map[string]string
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request data", err.Error()))
		return
	}

	reason := requestData["reason"]
	if reason == "" {
		reason = "Cancelled by user"
	}

	// Get user ID from context (set by auth middleware)
	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	err = h.schedulingService.CancelAppointment(uint(appointmentID), reason, *userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, NewErrorResponse("Appointment not found", ""))
			return
		}
		
		logger.Error("Failed to cancel appointment", err, map[string]interface{}{
			"appointment_id": appointmentID,
			"reason":         reason,
			"user_id":        *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to cancel appointment", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Appointment cancelled successfully", nil))
}

// GetUserAppointments godoc
// @Summary Get user appointments
// @Description Get appointments for a specific user with optional filters
// @Tags appointments
// @Accept json
// @Produce json
// @Param user_id query int false "User ID (defaults to current user)"
// @Param start_date query string false "Start date filter (ISO 8601)"
// @Param end_date query string false "End date filter (ISO 8601)"
// @Param status query string false "Status filter"
// @Success 200 {object} APIResponse{data=[]models.AppointmentResponse}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /appointments/user [get]
func (h *SchedulingHandler) GetUserAppointments(c *gin.Context) {
	// Get user ID from context or query parameter
	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	// Check if a different user ID is specified (admin only)
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if specifiedUserID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			// TODO: Add role check for admin access
			userID = new(uint)
			*userID = uint(specifiedUserID)
		}
	}

	// Parse date filters
	var startDate, endDate time.Time
	var err error

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err = time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid start_date format", "Use ISO 8601 format"))
			return
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err = time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid end_date format", "Use ISO 8601 format"))
			return
		}
	}

	status := c.Query("status")

	appointments, err := h.schedulingService.GetUserAppointments(*userID, startDate, endDate, status)
	if err != nil {
		logger.Error("Failed to get user appointments", err, map[string]interface{}{
			"user_id": *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get appointments", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("User appointments retrieved successfully", appointments))
}

// GetContactAppointments godoc
// @Summary Get contact appointments
// @Description Get all appointments for a specific contact
// @Tags appointments
// @Accept json
// @Produce json
// @Param id path int true "Contact ID"
// @Success 200 {object} APIResponse{data=[]models.AppointmentResponse}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /contacts/{id}/appointments [get]
func (h *SchedulingHandler) GetContactAppointments(c *gin.Context) {
	contactID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid contact ID", err.Error()))
		return
	}

	appointments, err := h.schedulingService.GetContactAppointments(uint(contactID))
	if err != nil {
		logger.Error("Failed to get contact appointments", err, map[string]interface{}{
			"contact_id": contactID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get contact appointments", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Contact appointments retrieved successfully", appointments))
}

// FindAvailableSlots godoc
// @Summary Find available time slots
// @Description Find available time slots for scheduling based on user availability
// @Tags availability
// @Accept json
// @Produce json
// @Param request body models.AvailabilitySlotRequest true "Availability search criteria"
// @Success 200 {object} APIResponse{data=[]models.AvailabilitySlot}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /appointments/available-slots [post]
func (h *SchedulingHandler) FindAvailableSlots(c *gin.Context) {
	var req models.AvailabilitySlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request data", err.Error()))
		return
	}

	slots, err := h.schedulingService.FindAvailableSlots(&req)
	if err != nil {
		logger.Error("Failed to find available slots", err, map[string]interface{}{
			"user_id":    req.UserID,
			"start_date": req.StartDate,
			"end_date":   req.EndDate,
			"duration":   req.Duration,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to find available slots", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Available slots retrieved successfully", slots))
}

// GetUserAvailability godoc
// @Summary Get user availability
// @Description Get availability information for a user on a specific date
// @Tags availability
// @Accept json
// @Produce json
// @Param user_id query int true "User ID"
// @Param date query string true "Date (YYYY-MM-DD)"
// @Success 200 {object} APIResponse{data=models.AvailabilityResponse}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /appointments/availability [get]
func (h *SchedulingHandler) GetUserAvailability(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse("user_id is required", ""))
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid user_id", err.Error()))
		return
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse("date is required", "Use YYYY-MM-DD format"))
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid date format", "Use YYYY-MM-DD format"))
		return
	}

	availability, err := h.schedulingService.GetUserAvailability(uint(userID), date)
	if err != nil {
		logger.Error("Failed to get user availability", err, map[string]interface{}{
			"user_id": userID,
			"date":    date,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get user availability", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("User availability retrieved successfully", availability))
}

// GetMyAppointments godoc
// @Summary Get current user's appointments
// @Description Get appointments for the authenticated user
// @Tags appointments
// @Accept json
// @Produce json
// @Param start_date query string false "Start date filter (ISO 8601)"
// @Param end_date query string false "End date filter (ISO 8601)"
// @Param status query string false "Status filter"
// @Success 200 {object} APIResponse{data=[]models.AppointmentResponse}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /appointments/my [get]
func (h *SchedulingHandler) GetMyAppointments(c *gin.Context) {
	// Get user ID from context
	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	// Parse date filters
	var startDate, endDate time.Time
	var err error

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err = time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid start_date format", "Use ISO 8601 format"))
			return
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err = time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid end_date format", "Use ISO 8601 format"))
			return
		}
	}

	status := c.Query("status")

	appointments, err := h.schedulingService.GetUserAppointments(*userID, startDate, endDate, status)
	if err != nil {
		logger.Error("Failed to get user appointments", err, map[string]interface{}{
			"user_id": *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get appointments", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Your appointments retrieved successfully", appointments))
}

// GetTodaysAppointments godoc
// @Summary Get today's appointments
// @Description Get all appointments for the current user scheduled for today
// @Tags appointments
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=[]models.AppointmentResponse}
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /appointments/today [get]
func (h *SchedulingHandler) GetTodaysAppointments(c *gin.Context) {
	// Get user ID from context
	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	// Set date range for today
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	appointments, err := h.schedulingService.GetUserAppointments(*userID, startOfDay, endOfDay, "")
	if err != nil {
		logger.Error("Failed to get today's appointments", err, map[string]interface{}{
			"user_id": *userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get today's appointments", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Today's appointments retrieved successfully", appointments))
}

// GetUpcomingAppointments godoc
// @Summary Get upcoming appointments
// @Description Get upcoming appointments for the current user (next 7 days)
// @Tags appointments
// @Accept json
// @Produce json
// @Param days query int false "Number of days to look ahead" default(7)
// @Success 200 {object} APIResponse{data=[]models.AppointmentResponse}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /appointments/upcoming [get]
func (h *SchedulingHandler) GetUpcomingAppointments(c *gin.Context) {
	// Get user ID from context
	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedResponse())
		return
	}

	// Parse days parameter
	days := 7
	if daysStr := c.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	// Set date range
	now := time.Now()
	startDate := now
	endDate := now.Add(time.Duration(days) * 24 * time.Hour)

	appointments, err := h.schedulingService.GetUserAppointments(*userID, startDate, endDate, "")
	if err != nil {
		logger.Error("Failed to get upcoming appointments", err, map[string]interface{}{
			"user_id": *userID,
			"days":    days,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get upcoming appointments", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Upcoming appointments retrieved successfully", appointments))
}

// GetAppointments retrieves appointments with filtering
func GetAppointments(c *gin.Context) {
	_ = NewSchedulingHandler()
	
	// Parse query parameters
	_, _ = parsePaginationParams(c)
	_ = c.Query("status")
	userIDParam := c.Query("user_id")
	
	var userID *uint
	if userIDParam != "" {
		id, err := strconv.ParseUint(userIDParam, 10, 32)
		if err == nil {
			uid := uint(id)
			userID = &uid
		}
	}
	_ = userID
	
	// Get appointments - this would need to be implemented in the service
	// For now, return a placeholder response
	c.JSON(http.StatusOK, NewSuccessResponse("Appointments retrieved successfully", []interface{}{}))
}

// ConfirmAppointment confirms an appointment
func ConfirmAppointment(c *gin.Context) {
	handler := NewSchedulingHandler()
	
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid appointment ID", "ID must be a valid number"))
		return
	}

	userID := getUserIDFromContext(c)
	
	// Update appointment status to confirmed
	updateRequest := &models.AppointmentUpdateRequest{
		Status: string(models.AppointmentConfirmed),
	}
	
	err = handler.schedulingService.UpdateAppointmentStatus(uint(id), updateRequest, *userID)
	if err != nil {
		logger.Error("Failed to confirm appointment", err, map[string]interface{}{
			"appointment_id": id,
			"user_id":        userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to confirm appointment", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Appointment confirmed successfully", nil))
}

// RescheduleAppointment reschedules an appointment
func RescheduleAppointment(c *gin.Context) {
	handler := NewSchedulingHandler()
	
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid appointment ID", "ID must be a valid number"))
		return
	}

	var req models.RescheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid input", err.Error()))
		return
	}

	userID := getUserIDFromContext(c)
	
	appointment, err := handler.schedulingService.RescheduleAppointment(uint(id), &req, *userID)
	if err != nil {
		logger.Error("Failed to reschedule appointment", err, map[string]interface{}{
			"appointment_id": id,
			"user_id":        userID,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to reschedule appointment", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Appointment rescheduled successfully", appointment))
}

// CancelAppointment cancels an appointment
func CancelAppointment(c *gin.Context) {
	handler := NewSchedulingHandler()
	
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid appointment ID", "ID must be a valid number"))
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid input", err.Error()))
		return
	}

	userID := getUserIDFromContext(c)
	
	err = handler.schedulingService.CancelAppointment(uint(id), req.Reason, *userID)
	if err != nil {
		logger.Error("Failed to cancel appointment", err, map[string]interface{}{
			"appointment_id": id,
			"user_id":        userID,
			"reason":         req.Reason,
		})
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to cancel appointment", err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Appointment cancelled successfully", nil))
}

