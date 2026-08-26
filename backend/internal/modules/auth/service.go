package auth

import (
	"errors"
	"strings"

	"github.com/lfhonda/metappstwo.git/backend/internal/utils"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailAlreadyExists = errors.New("email already registered")
)

type AuthService struct {
	db *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{
		db: db,
	}
}

type RegisterInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

type UserResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

func (s *AuthService) Login(input LoginInput) (*AuthResponse, error) {
	email := normalizeEmail(input.Email)

	var user User

	err := s.db.
		Where("email = ?", email).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, err
	}

	if !utils.CompareHashPassword(user.Password, input.Password) {
		return nil, ErrInvalidCredentials
	}

	token, err := utils.GenerateToken(
		user.ID,
		user.Email,
		user.Role,
	)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		User: UserResponse{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
			Role:  user.Role,
		},
		Token: token,
	}, nil
}

func (s *AuthService) RegisterStudent(input RegisterInput) (*UserResponse, error) {
	return s.registerUser(input, "student")
}

func (s *AuthService) RegisterTeacher(input RegisterInput) (*UserResponse, error) {
	return s.registerUser(input, "teacher")
}

func (s *AuthService) registerUser(input RegisterInput, role string) (*UserResponse, error) {
	email := normalizeEmail(input.Email)
	name := strings.TrimSpace(input.Name)

	var existingUser User

	err := s.db.
		Where("email = ?", email).
		First(&existingUser).Error

	if err == nil {
		return nil, ErrEmailAlreadyExists
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	passwordHash, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := User{
		Email:    email,
		Password: passwordHash,
		Name:     name,
		Role:     role,
	}

	if err := s.db.Create(&user).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return nil, ErrEmailAlreadyExists
		}

		return nil, err
	}

	return &UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Role:  user.Role,
	}, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
