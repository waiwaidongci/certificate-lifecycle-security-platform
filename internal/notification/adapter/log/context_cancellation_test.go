package log

import (
	"context"
	"errors"
	"testing"

	"github.com/acme/certpilot/internal/notification/domain"
	"github.com/acme/certpilot/internal/shared/logger"
)

func TestLogSendHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := NewAdapter(logger.New("error")).Send(ctx, domain.Reminder{ID: "cancelled"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Send error = %v, want context cancellation", err)
	}
}
