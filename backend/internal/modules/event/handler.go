package event

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// ============================================================
// REQUESTS
// ============================================================

type CreateEventRequest struct {
	Title           string    `json:"title" binding:"required"`
	Description     string    `json:"description" binding:"required"`
	Place           string    `json:"place" binding:"required"`
	Category        string    `json:"category" binding:"required"`
	Date            time.Time `json:"date" binding:"required"`
	EndDate         time.Time `json:"end_date" binding:"required"`
	MaxParticipants int       `json:"max_participants" binding:"required,gt=0"`
}

type UpdateEventRequest struct {
	Title           string    `json:"title" binding:"required"`
	Description     string    `json:"description" binding:"required"`
	Place           string    `json:"place" binding:"required"`
	Category        string    `json:"category" binding:"required"`
	Date            time.Time `json:"date" binding:"required"`
	EndDate         time.Time `json:"end_date" binding:"required"`
	MaxParticipants int       `json:"max_participants" binding:"required,gt=0"`
}

type CheckInRequest struct {
	Token string `json:"token" binding:"required"`
}

// ============================================================
// HELPERS
// ============================================================

func getIDParam(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New("invalid id")
	}

	return uint(id), nil
}

func getUserID(c *gin.Context) (uint, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	userID, ok := value.(uint)
	if !ok {
		return 0, false
	}

	return userID, true
}

func frontendURL() string {
	url := os.Getenv("FRONTEND_URL")

	if url == "" {
		return "http://localhost:3000"
	}

	return url
}

// ============================================================
// CREATE EVENT
// POST /events
// Teacher only
// ============================================================

func (h *Handler) Create(c *gin.Context) {
	var req CreateEventRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	event := &Event{
		Title:           req.Title,
		Description:     req.Description,
		Place:           req.Place,
		Category:        req.Category,
		Date:            req.Date,
		EndDate:         req.EndDate,
		MaxParticipants: req.MaxParticipants,
	}

	if err := h.service.CreateEvent(event); err != nil {
		switch {
		case errors.Is(err, ErrInvalidEventDate):
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Bad Request",
				"message": "End date must be after start date",
			})

		case errors.Is(err, ErrInvalidMaxPlaces):
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Bad Request",
				"message": "Max participants must be greater than zero",
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "Failed to create event",
			})
		}

		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Event created successfully",
		"data":    event,
	})
}

// ============================================================
// LIST EVENTS
// GET /events
// Authenticated users
// Only future events are returned by the service
// ============================================================

func (h *Handler) List(c *gin.Context) {
	events, err := h.service.ListEvents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to list events",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": events,
	})
}

// ============================================================
// GET EVENT
// GET /events/:id
// Authenticated users
// ============================================================

func (h *Handler) Get(c *gin.Context) {
	eventID, err := getIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid event ID",
		})
		return
	}

	event, err := h.service.GetEventDetails(eventID)
	if err != nil {
		if errors.Is(err, ErrEventNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": "Event not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to get event",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": event,
	})
}

// ============================================================
// UPDATE EVENT
// PUT /events/:id
// Teacher only
// ============================================================

func (h *Handler) Update(c *gin.Context) {
	eventID, err := getIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid event ID",
		})
		return
	}

	var req UpdateEventRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	event := &Event{
		Title:           req.Title,
		Description:     req.Description,
		Place:           req.Place,
		Category:        req.Category,
		Date:            req.Date,
		EndDate:         req.EndDate,
		MaxParticipants: req.MaxParticipants,
	}

	updatedEvent, err := h.service.UpdateEvent(
		eventID,
		event,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrEventNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": "Event not found",
			})

		case errors.Is(err, ErrInvalidEventDate):
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Bad Request",
				"message": "End date must be after start date",
			})

		case errors.Is(err, ErrInvalidMaxPlaces):
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Bad Request",
				"message": "Max participants must be greater than zero",
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "Failed to update event",
			})
		}

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Event updated successfully",
		"data":    updatedEvent,
	})
}

// ============================================================
// DELETE EVENT
// DELETE /events/:id
// Teacher only
//
// The service also deletes:
// - registrations
// - attendances
// - certificates
// ============================================================

func (h *Handler) Delete(c *gin.Context) {
	eventID, err := getIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid event ID",
		})
		return
	}

	if err := h.service.DeleteEvent(eventID); err != nil {
		if errors.Is(err, ErrEventNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": "Event not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to delete event",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Event deleted successfully",
	})
}

// ============================================================
// REGISTER
// POST /events/:id/register
// Student only
// ============================================================

func (h *Handler) Register(c *gin.Context) {
	eventID, err := getIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid event ID",
		})
		return
	}

	studentID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	registration, err := h.service.RegisterStudent(
		eventID,
		studentID,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrEventNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": "Event not found",
			})

		case errors.Is(err, ErrAlreadyRegistered):
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Conflict",
				"message": "Student is already registered for this event",
			})

		case errors.Is(err, ErrEventFull):
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Conflict",
				"message": "Event is full",
			})

		case errors.Is(err, ErrCheckInClosed):
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Conflict",
				"message": "Registration is no longer available",
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "Failed to register student",
			})
		}

		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Student registered successfully",
		"data":    registration,
	})
}

// ============================================================
// CANCEL REGISTRATION
// DELETE /events/:id/register
// Student only
// ============================================================

func (h *Handler) CancelRegistration(c *gin.Context) {
	eventID, err := getIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid event ID",
		})
		return
	}

	studentID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	err = h.service.CancelRegistration(
		eventID,
		studentID,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrRegistrationNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": "Registration not found",
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "Failed to cancel registration",
			})
		}

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Registration cancelled successfully",
	})
}

// ============================================================
// CHECK-IN SCREEN
// GET /events/:id/check-in
// Teacher only
// ============================================================

func (h *Handler) CheckInScreen(c *gin.Context) {
	eventID, err := getIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid event ID",
		})
		return
	}

	info, err := h.service.GetCheckInInfo(
		eventID,
		frontendURL(),
	)

	if err != nil {
		if errors.Is(err, ErrEventNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": "Event not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to load check-in information",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": info,
	})
}

// ============================================================
// CHECK-IN
// POST /events/:id/check-in
// Student only
// ============================================================

func (h *Handler) CheckIn(c *gin.Context) {
	eventID, err := getIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid event ID",
		})
		return
	}

	studentID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	var req CheckInRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Token is required",
		})
		return
	}

	result, err := h.service.CheckIn(
		eventID,
		studentID,
		req.Token,
		time.Now(),
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrEventNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": "Event not found",
			})

		case errors.Is(err, ErrInvalidCheckInToken):
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Invalid check-in token",
			})

		case errors.Is(err, ErrCheckInClosed):
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Conflict",
				"message": "Check-in is not available at this time",
			})

		case errors.Is(err, ErrNotRegistered):
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Student is not registered for this event",
			})

		case errors.Is(err, ErrAlreadyCheckedIn):
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Conflict",
				"message": "Student already checked in",
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "Failed to check in",
			})
		}

		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Attendance confirmed successfully",
		"data": gin.H{
			"attendance": result.Attendance,
			"certificate": gin.H{
				"id":   result.Certificate.ID,
				"code": result.Certificate.Code,
			},
		},
	})
}

// ============================================================
// CERTIFICATE PDF
// GET /certificates/:id/pdf
// Student only
// ============================================================

func (h *Handler) CertificatePDF(c *gin.Context) {
	certificateID, err := getIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid certificate ID",
		})
		return
	}

	studentID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	certificate, err := h.service.GetCertificate(
		certificateID,
	)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": "Certificate not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to get certificate",
		})
		return
	}

	// O aluno só pode baixar seus próprios certificados.
	if certificate.StudentID != studentID {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "You cannot access this certificate",
		})
		return
	}

	var student struct {
		Name string
	}

	if err := h.service.db.
		Table("users").
		Select("name").
		Where("id = ?", certificate.StudentID).
		Scan(&student).
		Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to load student",
		})
		return
	}

	var event Event

	if err := h.service.db.
		First(&event, certificate.EventID).
		Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": "Event not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to load event",
		})
		return
	}

	var attendance EventAttendance

	if err := h.service.db.
		Where(
			"event_id = ? AND student_id = ?",
			certificate.EventID,
			certificate.StudentID,
		).
		First(&attendance).
		Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Attendance not found",
		})
		return
	}

	pdf, err := GenerateCertificatePDF(
		CertificateData{
			StudentName: student.Name,
			EventTitle:  event.Title,
			Description: event.Description,
			Category:    event.Category,
			Place:       event.Place,
			Date:        event.Date.Format("02/01/2006"),
			EndDate:     event.EndDate.Format("02/01/2006"),
			CheckedInAt: attendance.CheckedInAt.Format("02/01/2006 15:04"),
			Code:        certificate.Code,
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to generate certificate",
		})
		return
	}

	c.Header(
		"Content-Disposition",
		"attachment; filename=\"certificate.pdf\"",
	)

	c.Data(
		http.StatusOK,
		"application/pdf",
		pdf,
	)
}

// ============================================================
// VERIFY CERTIFICATE
// GET /certificates/verify/:code
// Public
// ============================================================

func (h *Handler) VerifyCertificate(c *gin.Context) {
	code := c.Param("code")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Certificate code is required",
		})
		return
	}

	certificate, err := h.service.GetCertificateByCode(code)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"valid": false,
				"error": "Not Found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to verify certificate",
		})
		return
	}

	var student struct {
		Name string `json:"name"`
	}

	if err := h.service.db.
		Table("users").
		Select("name").
		Where("id = ?", certificate.StudentID).
		Scan(&student).
		Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to load student",
		})
		return
	}

	var event Event

	if err := h.service.db.
		First(&event, certificate.EventID).
		Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to load event",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid": true,
		"data": gin.H{
			"certificate_code": certificate.Code,

			"student": student,

			"event": gin.H{
				"title":       event.Title,
				"description": event.Description,
				"category":    event.Category,
				"place":       event.Place,
				"date":        event.Date,
				"end_date":    event.EndDate,
			},

			"issued_at": certificate.IssuedAt,
		},
	})
}
