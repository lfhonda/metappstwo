package event

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrEventNotFound        = errors.New("event not found")
	ErrEventFull            = errors.New("event is full")
	ErrAlreadyRegistered    = errors.New("student is already registered")
	ErrRegistrationNotFound = errors.New("registration not found")

	ErrAlreadyCheckedIn    = errors.New("student already checked in")
	ErrCheckInClosed       = errors.New("check-in is not available")
	ErrInvalidCheckInToken = errors.New("invalid check-in token")
	ErrNotRegistered       = errors.New("student is not registered for this event")

	ErrInvalidEventDate = errors.New("event end date must be after start date")
	ErrInvalidMaxPlaces = errors.New("max participants must be greater than zero")
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

// ============================================================
// EVENT
// ============================================================

func (s *Service) CreateEvent(event *Event) error {
	if event.EndDate.Before(event.Date) || event.EndDate.Equal(event.Date) {
		return ErrInvalidEventDate
	}

	if event.MaxParticipants <= 0 {
		return ErrInvalidMaxPlaces
	}

	event.CheckInToken = uuid.NewString()

	return s.db.Create(event).Error
}

func (s *Service) GetEvent(id uint) (*Event, error) {
	var event Event

	if err := s.db.First(&event, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEventNotFound
		}

		return nil, err
	}

	return &event, nil
}

func (s *Service) ListEvents() ([]Event, error) {
	var events []Event

	err := s.db.
		Where("date > ?", time.Now()).
		Order("date ASC").
		Find(&events).
		Error

	if err != nil {
		return nil, err
	}

	return events, nil
}

func (s *Service) UpdateEvent(id uint, data *Event) (*Event, error) {
	var event Event

	if err := s.db.First(&event, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEventNotFound
		}

		return nil, err
	}

	if data.EndDate.Before(data.Date) || data.EndDate.Equal(data.Date) {
		return nil, ErrInvalidEventDate
	}

	if data.MaxParticipants <= 0 {
		return nil, ErrInvalidMaxPlaces
	}

	var registeredCount int64

	if err := s.db.
		Model(&EventRegistration{}).
		Where("event_id = ?", id).
		Count(&registeredCount).
		Error; err != nil {
		return nil, err
	}

	if int64(data.MaxParticipants) < registeredCount {
		return nil, fmt.Errorf(
			"max participants cannot be lower than current registrations (%d)",
			registeredCount,
		)
	}

	event.Title = data.Title
	event.Description = data.Description
	event.Place = data.Place
	event.Category = data.Category
	event.Date = data.Date
	event.EndDate = data.EndDate
	event.MaxParticipants = data.MaxParticipants

	if err := s.db.Save(&event).Error; err != nil {
		return nil, err
	}

	return &event, nil
}

func (s *Service) DeleteEvent(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var event Event

		if err := tx.First(&event, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrEventNotFound
			}

			return err
		}

		// Remove inscrições.
		if err := tx.
			Where("event_id = ?", id).
			Delete(&EventRegistration{}).
			Error; err != nil {
			return err
		}

		// Remove presenças.
		if err := tx.
			Where("event_id = ?", id).
			Delete(&EventAttendance{}).
			Error; err != nil {
			return err
		}

		// Remove certificados.
		if err := tx.
			Where("event_id = ?", id).
			Delete(&Certificate{}).
			Error; err != nil {
			return err
		}

		// Remove o evento.
		if err := tx.Delete(&event).Error; err != nil {
			return err
		}

		return nil
	})
}

// ============================================================
// EVENT DETAILS
// ============================================================

type EventDetails struct {
	Event

	RegisteredCount int64 `json:"registered_count"`
	AvailableSlots  int64 `json:"available_slots"`
}

func (s *Service) GetEventDetails(id uint) (*EventDetails, error) {
	event, err := s.GetEvent(id)
	if err != nil {
		return nil, err
	}

	var registeredCount int64

	if err := s.db.
		Model(&EventRegistration{}).
		Where("event_id = ?", id).
		Count(&registeredCount).
		Error; err != nil {
		return nil, err
	}

	availableSlots := int64(event.MaxParticipants) - registeredCount

	if availableSlots < 0 {
		availableSlots = 0
	}

	return &EventDetails{
		Event:           *event,
		RegisteredCount: registeredCount,
		AvailableSlots:  availableSlots,
	}, nil
}

// ============================================================
// REGISTRATION
// ============================================================

func (s *Service) RegisterStudent(
	eventID uint,
	studentID uint,
) (*EventRegistration, error) {

	var registration EventRegistration

	err := s.db.
		Where(
			"event_id = ? AND student_id = ?",
			eventID,
			studentID,
		).
		First(&registration).
		Error

	if err == nil {
		return nil, ErrAlreadyRegistered
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var event Event

	if err := s.db.First(&event, eventID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEventNotFound
		}

		return nil, err
	}

	// Não permite inscrição em eventos que já começaram.
	if !time.Now().Before(event.Date) {
		return nil, ErrCheckInClosed
	}

	var registeredCount int64

	if err := s.db.
		Model(&EventRegistration{}).
		Where("event_id = ?", eventID).
		Count(&registeredCount).
		Error; err != nil {
		return nil, err
	}

	if registeredCount >= int64(event.MaxParticipants) {
		return nil, ErrEventFull
	}

	registration = EventRegistration{
		EventID:   eventID,
		StudentID: studentID,
	}

	if err := s.db.Create(&registration).Error; err != nil {
		return nil, err
	}

	return &registration, nil
}

func (s *Service) CancelRegistration(
	eventID uint,
	studentID uint,
) error {

	var registration EventRegistration

	err := s.db.
		Where(
			"event_id = ? AND student_id = ?",
			eventID,
			studentID,
		).
		First(&registration).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrRegistrationNotFound
	}

	if err != nil {
		return err
	}

	return s.db.Delete(&registration).Error
}

// ============================================================
// CHECK-IN
// ============================================================

type CheckInResult struct {
	Attendance  *EventAttendance `json:"attendance"`
	Certificate *Certificate     `json:"certificate"`
}

func (s *Service) CheckIn(
	eventID uint,
	studentID uint,
	token string,
	now time.Time,
) (*CheckInResult, error) {

	var result CheckInResult

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var event Event

		if err := tx.First(&event, eventID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrEventNotFound
			}

			return err
		}

		// Valida o token do QR Code.
		if event.CheckInToken != token {
			return ErrInvalidCheckInToken
		}

		// O check-in começa 30 minutos antes do evento.
		checkInStart := event.Date.Add(-30 * time.Minute)

		// O check-in termina quando o evento termina.
		if now.Before(checkInStart) || now.After(event.EndDate) {
			return ErrCheckInClosed
		}

		// O aluno precisa estar inscrito.
		var registration EventRegistration

		err := tx.
			Where(
				"event_id = ? AND student_id = ?",
				eventID,
				studentID,
			).
			First(&registration).
			Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotRegistered
		}

		if err != nil {
			return err
		}

		// Não permite check-in duplicado.
		var existingAttendance EventAttendance

		err = tx.
			Where(
				"event_id = ? AND student_id = ?",
				eventID,
				studentID,
			).
			First(&existingAttendance).
			Error

		if err == nil {
			return ErrAlreadyCheckedIn
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		attendance := &EventAttendance{
			EventID:     eventID,
			StudentID:   studentID,
			CheckedInAt: now,
		}

		if err := tx.Create(attendance).Error; err != nil {
			return err
		}

		// Cria o certificado dentro da mesma transaction.
		certificate := &Certificate{
			EventID:   eventID,
			StudentID: studentID,
			Code:      generateCertificateCode(),
			IssuedAt:  now,
		}

		if err := tx.Create(certificate).Error; err != nil {
			return err
		}

		result.Attendance = attendance
		result.Certificate = certificate

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func generateCertificateCode() string {
	return fmt.Sprintf(
		"CERT-%s",
		uuid.NewString(),
	)
}

// ============================================================
// ATTENDANCE / CHECK-IN SCREEN
// ============================================================

type CheckInInfo struct {
	EventID         uint   `json:"event_id"`
	Title           string `json:"title"`
	Date            string `json:"date"`
	EndDate         string `json:"end_date"`
	CheckInStartsAt string `json:"check_in_starts_at"`
	CheckInEndsAt   string `json:"check_in_ends_at"`

	QRCode string `json:"qr_code"`

	TotalRegistered int64 `json:"total_registered"`
	TotalPresent    int64 `json:"total_present"`
}

func (s *Service) GetCheckInInfo(
	eventID uint,
	frontendURL string,
) (*CheckInInfo, error) {

	var event Event

	if err := s.db.First(&event, eventID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEventNotFound
		}

		return nil, err
	}

	qrCode, err := GenerateCheckInQRCode(
		&event,
		frontendURL,
	)

	if err != nil {
		return nil, err
	}

	var registered int64
	var present int64

	if err := s.db.
		Model(&EventRegistration{}).
		Where("event_id = ?", eventID).
		Count(&registered).
		Error; err != nil {
		return nil, err
	}

	if err := s.db.
		Model(&EventAttendance{}).
		Where("event_id = ?", eventID).
		Count(&present).
		Error; err != nil {
		return nil, err
	}

	return &CheckInInfo{
		EventID:         event.ID,
		Title:           event.Title,
		Date:            event.Date.Format(time.RFC3339),
		EndDate:         event.EndDate.Format(time.RFC3339),
		CheckInStartsAt: event.Date.Add(-30 * time.Minute).Format(time.RFC3339),
		CheckInEndsAt:   event.EndDate.Format(time.RFC3339),

		QRCode: qrCode,

		TotalRegistered: registered,
		TotalPresent:    present,
	}, nil
}

func (s *Service) GetAttendanceCount(
	eventID uint,
) (int64, error) {

	var count int64

	err := s.db.
		Model(&EventAttendance{}).
		Where("event_id = ?", eventID).
		Count(&count).
		Error

	return count, err
}

// ============================================================
// CERTIFICATE
// ============================================================

func (s *Service) GetCertificate(
	id uint,
) (*Certificate, error) {

	var certificate Certificate

	if err := s.db.First(&certificate, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}

		return nil, err
	}

	return &certificate, nil
}

func (s *Service) GetCertificateByCode(
	code string,
) (*Certificate, error) {

	var certificate Certificate

	if err := s.db.
		Where("code = ?", code).
		First(&certificate).
		Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}

		return nil, err
	}

	return &certificate, nil
}
