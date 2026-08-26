package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	service *AuthService
}

func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
		return
	}

	response, err := h.service.Login(input)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid email or password",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to authenticate user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"data":    response,
	})
}

func (h *AuthHandler) RegisterTeacher(c *gin.Context) {
	var input RegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
		return
	}

	user, err := h.service.RegisterTeacher(input)
	if err != nil {
		h.handleRegistrationError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Teacher registered successfully",
		"data":    user,
	})
}

func (h *AuthHandler) RegisterStudent(c *gin.Context) {
	var input RegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
		return
	}

	user, err := h.service.RegisterStudent(input)
	if err != nil {
		h.handleRegistrationError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Student registered successfully",
		"data":    user,
	})
}

func (h *AuthHandler) handleRegistrationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrEmailAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Conflict",
			"message": "Email already registered",
		})

	case errors.Is(err, gorm.ErrDuplicatedKey):
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Conflict",
			"message": "Email already registered",
		})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to register user",
		})
	}
}
