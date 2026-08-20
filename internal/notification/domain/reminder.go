package domain

import (
	"context"
	"time"

	"github.com/acme/certpilot/internal/shared/database"
)

const (
	ReminderPending = "pending"
	ReminderSent    = "sent"
	ReminderFailed  = "failed"
)

type Reminder struct {
	ID            string     `json:"id"`
	CertificateID string     `json:"certificate_id"`
	ServiceID     string     `json:"service_id"`
	DaysLeft      int        `json:"days_left"`
	Channel       string     `json:"channel"`
	Recipient     string     `json:"recipient"`
	Status        string     `json:"status"`
	Message       string     `json:"message"`
	SentAt        *time.Time `json:"sent_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type Repository interface {
	Create(ctx context.Context, exec database.Executor, reminder Reminder) error
	Get(ctx context.Context, exec database.Executor, id string) (Reminder, error)
	List(ctx context.Context, exec database.Executor, options ListOptions) ([]Reminder, int, error)
	ListPending(ctx context.Context, exec database.Executor, limit int) ([]Reminder, error)
	UpdateStatus(ctx context.Context, exec database.Executor, reminder Reminder) error
	GetExisting(ctx context.Context, exec database.Executor, certificateID string, daysLeft int) (Reminder, error)
}

type ListOptions struct {
	Page     int
	PageSize int
	Filters  map[string]string
	Sort     string
	Order    string
}
