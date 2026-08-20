package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	certapp "github.com/acme/certpilot/internal/certificate/application"
	"github.com/acme/certpilot/internal/notification/domain"
	serviceapp "github.com/acme/certpilot/internal/servicecatalog/application"
	"github.com/acme/certpilot/internal/shared/apperror"
	"github.com/acme/certpilot/internal/shared/clock"
	"github.com/acme/certpilot/internal/shared/database"
	"github.com/acme/certpilot/internal/shared/eventlog/application"
	"github.com/acme/certpilot/internal/shared/httpx"
	"github.com/acme/certpilot/internal/shared/id"
)

type NotificationAdapter interface {
	Send(ctx context.Context, reminder domain.Reminder) (string, error)
}

type Service struct {
	db           *database.DB
	repository   domain.Repository
	certificates *certapp.Service
	services     *serviceapp.Service
	events       *application.Service
	clock        clock.Clock
	adapters     map[string]NotificationAdapter
}

func NewService(db *database.DB, repository domain.Repository, certificates *certapp.Service, services *serviceapp.Service, events *application.Service, clk clock.Clock, adapters map[string]NotificationAdapter) *Service {
	return &Service{db: db, repository: repository, certificates: certificates, services: services, events: events, clock: clk, adapters: adapters}
}

func (s *Service) Scan(ctx context.Context, advanceDays int, channel string) ([]domain.Reminder, error) {
	if advanceDays <= 0 {
		return nil, apperror.Invalid("advance_days must be positive")
	}
	if channel == "" {
		channel = "log"
	}
	certificates, err := s.certificates.ListExpiring(ctx, s.clock.Now().AddDate(0, 0, advanceDays))
	if err != nil {
		return nil, err
	}
	created := make([]domain.Reminder, 0)
	for _, certificate := range certificates {
		serviceID := ""
		if certificate.ServiceID != nil {
			serviceID = *certificate.ServiceID
		}
		recipient := "system"
		if serviceID != "" {
			if service, getErr := s.services.Get(ctx, serviceID); getErr == nil {
				recipient = service.Owner
			}
		}
		daysLeft := int(time.Until(certificate.NotAfter).Hours() / 24)
		reminder, createErr := s.createReminder(ctx, certificate.ID, serviceID, daysLeft, channel, recipient)
		if createErr != nil {
			if isConflict(createErr) {
				continue
			}
			return created, createErr
		}
		created = append(created, reminder)
	}
	return created, nil
}

func (s *Service) createReminder(ctx context.Context, certificateID, serviceID string, daysLeft int, channel, recipient string) (domain.Reminder, error) {
	var reminder domain.Reminder
	err := s.db.WithTx(ctx, func(tx database.Tx) error {
		if _, getErr := s.repository.GetExisting(ctx, tx, certificateID, daysLeft); getErr == nil {
			return apperror.Conflict("reminder already exists")
		}
		reminder = domain.Reminder{
			ID:            id.New(),
			CertificateID: certificateID,
			ServiceID:     serviceID,
			DaysLeft:      daysLeft,
			Channel:       channel,
			Recipient:     recipient,
			Status:        domain.ReminderPending,
			Message:       fmt.Sprintf("certificate %s expires in %d days", certificateID, daysLeft),
			CreatedAt:     s.clock.Now(),
		}
		if err := s.repository.Create(ctx, tx, reminder); err != nil {
			return err
		}
		return s.events.Record(ctx, tx, httpx.ActorFromContext(ctx), "notification.reminder.created", "notification_reminder", reminder.ID, map[string]any{"certificate_id": certificateID})
	})
	if err != nil {
		return domain.Reminder{}, err
	}
	return reminder, nil
}

func (s *Service) SendPending(ctx context.Context, limit int) ([]domain.Reminder, error) {
	if limit <= 0 {
		limit = 20
	}
	reminders, err := s.repository.ListPending(ctx, s.db, limit)
	if err != nil {
		return nil, err
	}
	sent := make([]domain.Reminder, 0, len(reminders))
	for _, reminder := range reminders {
		adapter, ok := s.adapters[reminder.Channel]
		if !ok {
			reminder.Status = domain.ReminderFailed
			reminder.Message = "unsupported notification channel"
		} else if response, sendErr := adapter.Send(ctx, reminder); sendErr != nil {
			reminder.Status = domain.ReminderFailed
			reminder.Message = sendErr.Error()
		} else {
			reminder.Status = domain.ReminderSent
			reminder.Message = response
			now := s.clock.Now()
			reminder.SentAt = &now
		}
		if err := s.repository.UpdateStatus(ctx, s.db, reminder); err != nil {
			return sent, err
		}
		sent = append(sent, reminder)
	}
	return sent, nil
}

func (s *Service) Get(ctx context.Context, reminderID string) (domain.Reminder, error) {
	return s.repository.Get(ctx, s.db, reminderID)
}

func (s *Service) List(ctx context.Context, options domain.ListOptions) ([]domain.Reminder, int, error) {
	return s.repository.List(ctx, s.db, options)
}

func isConflict(err error) bool {
	var apiErr *apperror.Error
	return errors.As(err, &apiErr) && apiErr.Code == apperror.CodeConflict
}
