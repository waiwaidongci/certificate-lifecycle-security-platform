package sqlite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/acme/certpilot/internal/notification/domain"
	"github.com/acme/certpilot/internal/shared/database"
)

type Repository struct{}

func NewRepository() *Repository { return &Repository{} }

func (r *Repository) Create(ctx context.Context, exec database.Executor, reminder domain.Reminder) error {
	_, err := exec.ExecContext(ctx, `INSERT INTO notification_reminders (id, certificate_id, service_id, days_left, channel, recipient, status, message, sent_at, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		reminder.ID, reminder.CertificateID, reminder.ServiceID, reminder.DaysLeft, reminder.Channel, reminder.Recipient, reminder.Status, reminder.Message, nullableTime(reminder.SentAt), reminder.CreatedAt.Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("insert notification reminder: %w", err)
	}
	return nil
}

func (r *Repository) Get(ctx context.Context, exec database.Executor, id string) (domain.Reminder, error) {
	return scanReminder(exec.QueryRowContext(ctx, `SELECT id, certificate_id, service_id, days_left, channel, recipient, status, message, sent_at, created_at FROM notification_reminders WHERE id = ?`, id))
}

func (r *Repository) List(ctx context.Context, exec database.Executor, options domain.ListOptions) ([]domain.Reminder, int, error) {
	where, args := buildWhere(options.Filters)
	var total int
	if err := exec.QueryRowContext(ctx, "SELECT COUNT(*) FROM notification_reminders"+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count notification reminders: %w", err)
	}
	sort := sanitizeSort(options.Sort)
	order := strings.ToUpper(options.Order)
	if order != "DESC" {
		order = "ASC"
	}
	rows, err := exec.QueryContext(ctx, "SELECT id, certificate_id, service_id, days_left, channel, recipient, status, message, sent_at, created_at FROM notification_reminders"+where+" ORDER BY "+sort+" "+order+" LIMIT ? OFFSET ?", append(args, options.PageSize, (options.Page-1)*options.PageSize)...)
	if err != nil {
		return nil, 0, fmt.Errorf("list notification reminders: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Reminder, 0, options.PageSize)
	for rows.Next() {
		item, err := scanReminder(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *Repository) ListPending(ctx context.Context, exec database.Executor, limit int) ([]domain.Reminder, error) {
	rows, err := exec.QueryContext(ctx, `SELECT id, certificate_id, service_id, days_left, channel, recipient, status, message, sent_at, created_at FROM notification_reminders WHERE status = ? ORDER BY created_at ASC LIMIT ?`, domain.ReminderPending, limit)
	if err != nil {
		return nil, fmt.Errorf("list pending reminders: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Reminder, 0, limit)
	for rows.Next() {
		item, err := scanReminder(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) UpdateStatus(ctx context.Context, exec database.Executor, reminder domain.Reminder) error {
	_, err := exec.ExecContext(ctx, `UPDATE notification_reminders SET status = ?, message = ?, sent_at = ? WHERE id = ?`,
		reminder.Status, reminder.Message, nullableTime(reminder.SentAt), reminder.ID)
	if err != nil {
		return fmt.Errorf("update notification reminder: %w", err)
	}
	return nil
}

func (r *Repository) GetExisting(ctx context.Context, exec database.Executor, certificateID string, daysLeft int) (domain.Reminder, error) {
	return scanReminder(exec.QueryRowContext(ctx, `SELECT id, certificate_id, service_id, days_left, channel, recipient, status, message, sent_at, created_at FROM notification_reminders WHERE certificate_id = ? AND days_left = ?`, certificateID, daysLeft))
}
