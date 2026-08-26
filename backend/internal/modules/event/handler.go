package event

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type EventHandler struct {
	service *EventService
}

func NewEventHandler(service *EventService) *EventHandler {
	return &EventHandler{
		service: service,
	}
}

func (h *EventHandler) Create(c *gin.Context) {
	var input CreateEventInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
		return
	}

	event, err := h.service.Create(input)
	if err != nil {
		if errors.Is(err, ErrInvalidMaxParticipants) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Bad Request",
				"message": "Maximum participants must be greater than zero",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to create event",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Event created successfully",
		"data":    event,
	})
}

func (h *EventHandler) List(c *gin.Context) {
	events, err := h.service.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to fetch events",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": events,
	})
}

func (h *EventHandler) GetByID(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		return
	}

	event, err := h.service.GetByID(id)
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
			"message": "Failed to fetch event",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": event,
	})
}

func (h *EventHandler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		return
	}

	var input UpdateEventInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
		return
	}

	event, err := h.service.Update(id, input)
	if err != nil {
		switch {
		case errors.Is(err, ErrEventNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": "Event not found",
			})

		case errors.Is(err, ErrEventFull):
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Bad Request",
				"message": "Maximum participants cannot be lower than current registrations",
			})

		case errors.Is(err, ErrInvalidMaxParticipants):
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Bad Request",
				"message": "Maximum participants must be greater than zero",
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
		"data":    event,
	})
}

func (h *EventHandler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		return
	}

	if err := h.service.Delete(id); err != nil {
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

func (h *EventHandler) Register(c *gin.Context) {
	eventID, err := parseID(c)
	if err != nil {
		return
	}

	studentID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	studentIDUint, ok := studentID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Invalid user ID",
		})
		return
	}

	if err := h.service.RegisterStudent(eventID, studentIDUint); err != nil {
		switch {
		case errors.Is(err, ErrEventNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": "Event not found or event has already occurred",
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

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "Failed to register for event",
			})
		}

		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Successfully registered for event",
	})
}

func (h *EventHandler) CancelRegistration(c *gin.Context) {
	eventID, err := parseID(c)
	if err != nil {
		return
	}

	studentID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	studentIDUint, ok := studentID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Invalid user ID",
		})
		return
	}

	if err := h.service.CancelRegistration(eventID, studentIDUint); err != nil {
		if errors.Is(err, ErrRegistrationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": "Registration not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to cancel registration",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Registration cancelled successfully",
	})
}

func parseID(c *gin.Context) (uint, error) {
	idParam := c.Param("id")

	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid event ID",
		})

		return 0, errors.New("invalid event id")
	}

	return uint(id), nil
}

var _ = gorm.ErrRecordNotFound
