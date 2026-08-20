package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/acme/certpilot/internal/notification/domain"
	"github.com/acme/certpilot/internal/shared/apperror"
)

type scanner interface {
	Scan(dest ...any) error
}

func scanReminder(row scanner) (domain.Reminder, error) {
	var reminder domain.Reminder
	var sentAt sql.NullString
	var createdAt string
	if err := row.Scan(&reminder.ID, &reminder.CertificateID, &reminder.ServiceID, &reminder.DaysLeft, &reminder.Channel, &reminder.Recipient, &reminder.Status, &reminder.Message, &sentAt, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Reminder{}, apperror.NotFound("notification reminder not found")
		}
		return domain.Reminder{}, fmt.Errorf("scan notification reminder: %w", err)
	}
	reminder.SentAt = parseNullableTime(sentAt)
	reminder.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	return reminder, nil
}

func parseNullableTime(value sql.NullString) *time.Time {
	if !value.Valid {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value.String)
	if err != nil {
		return nil
	}
	return &parsed
}
