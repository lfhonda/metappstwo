package event

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrEventNotFound          = errors.New("event not found")
	ErrEventFull              = errors.New("event is full")
	ErrAlreadyRegistered      = errors.New("student is already registered")
	ErrRegistrationNotFound   = errors.New("registration not found")
	ErrInvalidMaxParticipants = errors.New("max participants must be greater than zero")
)

type EventService struct {
	db *gorm.DB
}

func NewEventService(db *gorm.DB) *EventService {
	return &EventService{
		db: db,
	}
}

type CreateEventInput struct {
	Title           string    `json:"title" binding:"required"`
	Description     string    `json:"description" binding:"required"`
	Place           string    `json:"place" binding:"required"`
	Category        string    `json:"category" binding:"required"`
	Date            time.Time `json:"date" binding:"required"`
	MaxParticipants int       `json:"max_participants" binding:"required,min=1"`
}

type UpdateEventInput struct {
	Title           string    `json:"title" binding:"required"`
	Description     string    `json:"description" binding:"required"`
	Place           string    `json:"place" binding:"required"`
	Category        string    `json:"category" binding:"required"`
	Date            time.Time `json:"date" binding:"required"`
	MaxParticipants int       `json:"max_participants" binding:"required,min=1"`
}

type EventResponse struct {
	ID              uint      `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Place           string    `json:"place"`
	Category        string    `json:"category"`
	Date            time.Time `json:"date"`
	MaxParticipants int       `json:"max_participants"`
	AvailableSlots  int       `json:"available_slots"`
}

func (s *EventService) Create(input CreateEventInput) (*EventResponse, error) {
	if input.MaxParticipants <= 0 {
		return nil, ErrInvalidMaxParticipants
	}

	event := Event{
		Title:           input.Title,
		Description:     input.Description,
		Place:           input.Place,
		Category:        input.Category,
		Date:            input.Date,
		MaxParticipants: input.MaxParticipants,
	}

	if err := s.db.Create(&event).Error; err != nil {
		return nil, err
	}

	return s.buildEventResponse(event)
}

func (s *EventService) List() ([]EventResponse, error) {
	var events []Event

	err := s.db.
		Where("date > ?", time.Now()).
		Order("date ASC").
		Find(&events).Error

	if err != nil {
		return nil, err
	}

	responses := make([]EventResponse, 0, len(events))

	for _, event := range events {
		response, err := s.buildEventResponse(event)
		if err != nil {
			return nil, err
		}

		responses = append(responses, *response)
	}

	return responses, nil
}

func (s *EventService) GetByID(id uint) (*EventResponse, error) {
	var event Event

	err := s.db.First(&event, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEventNotFound
		}

		return nil, err
	}

	return s.buildEventResponse(event)
}

func (s *EventService) Update(id uint, input UpdateEventInput) (*EventResponse, error) {
	if input.MaxParticipants <= 0 {
		return nil, ErrInvalidMaxParticipants
	}

	var event Event

	err := s.db.First(&event, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEventNotFound
		}

		return nil, err
	}

	var registrations int64

	if err := s.db.
		Model(&EventRegistration{}).
		Where("event_id = ?", id).
		Count(&registrations).Error; err != nil {
		return nil, err
	}

	// Não permite reduzir a capacidade abaixo da quantidade
	// atual de inscritos.
	if input.MaxParticipants < int(registrations) {
		return nil, ErrEventFull
	}

	event.Title = input.Title
	event.Description = input.Description
	event.Place = input.Place
	event.Category = input.Category
	event.Date = input.Date
	event.MaxParticipants = input.MaxParticipants

	if err := s.db.Save(&event).Error; err != nil {
		return nil, err
	}

	return s.buildEventResponse(event)
}

func (s *EventService) Delete(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var event Event

		err := tx.First(&event, id).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrEventNotFound
			}

			return err
		}

		if err := tx.
			Where("event_id = ?", id).
			Delete(&EventRegistration{}).Error; err != nil {
			return err
		}

		if err := tx.Delete(&event).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *EventService) RegisterStudent(eventID, studentID uint) error {
	var event Event

	err := s.db.First(&event, eventID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEventNotFound
		}

		return err
	}

	if !event.Date.After(time.Now()) {
		return ErrEventNotFound
	}

	var existing EventRegistration

	err = s.db.
		Where("event_id = ? AND student_id = ?", eventID, studentID).
		First(&existing).Error

	if err == nil {
		return ErrAlreadyRegistered
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	var registrations int64

	if err := s.db.
		Model(&EventRegistration{}).
		Where("event_id = ?", eventID).
		Count(&registrations).Error; err != nil {
		return err
	}

	if registrations >= int64(event.MaxParticipants) {
		return ErrEventFull
	}

	registration := EventRegistration{
		EventID:   eventID,
		StudentID: studentID,
	}

	if err := s.db.Create(&registration).Error; err != nil {
		// Caso duas requisições tentem registrar o mesmo aluno
		// simultaneamente, o índice único protege o banco.
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrAlreadyRegistered
		}

		return err
	}

	return nil
}

func (s *EventService) CancelRegistration(eventID, studentID uint) error {
	result := s.db.
		Where("event_id = ? AND student_id = ?", eventID, studentID).
		Delete(&EventRegistration{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrRegistrationNotFound
	}

	return nil
}

func (s *EventService) buildEventResponse(event Event) (*EventResponse, error) {
	var registrations int64

	if err := s.db.
		Model(&EventRegistration{}).
		Where("event_id = ?", event.ID).
		Count(&registrations).Error; err != nil {
		return nil, err
	}

	availableSlots := event.MaxParticipants - int(registrations)

	if availableSlots < 0 {
		availableSlots = 0
	}

	return &EventResponse{
		ID:              event.ID,
		Title:           event.Title,
		Description:     event.Description,
		Place:           event.Place,
		Category:        event.Category,
		Date:            event.Date,
		MaxParticipants: event.MaxParticipants,
		AvailableSlots:  availableSlots,
	}, nil
}
