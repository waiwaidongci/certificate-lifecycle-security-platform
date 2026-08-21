package log

import (
	"context"

	"github.com/acme/certpilot/internal/notification/domain"
	"github.com/acme/certpilot/internal/shared/logger"
)

type Adapter struct {
	log *logger.Logger
}

func NewAdapter(log *logger.Logger) *Adapter {
	return &Adapter{log: log}
}

func (a *Adapter) Send(ctx context.Context, reminder domain.Reminder) (string, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return "", err
		}
	}
	a.log.Info(ctx, "notification sent through log adapter", "reminder_id", reminder.ID, "recipient", reminder.Recipient)
	return "logged", nil
}
