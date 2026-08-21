package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/acme/certpilot/internal/notification/domain"
)

type Adapter struct {
	client  *http.Client
	timeout time.Duration
}

func NewAdapter(timeout time.Duration) *Adapter {
	return &Adapter{client: &http.Client{Timeout: timeout}, timeout: timeout}
}

func (a *Adapter) Send(ctx context.Context, reminder domain.Reminder) (string, error) {
	if reminder.Recipient == "" {
		return "", fmt.Errorf("recipient is empty")
	}
	body, _ := json.Marshal(reminder)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reminder.Recipient, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", &DeliveryError{StatusCode: resp.StatusCode}
	}
	return "webhook accepted", nil
}
