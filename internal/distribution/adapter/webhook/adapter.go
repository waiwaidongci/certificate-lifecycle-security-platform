package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/acme/certpilot/internal/distribution/domain"
)

type Adapter struct {
	client  *http.Client
	timeout time.Duration
}

func NewAdapter(timeout time.Duration) *Adapter {
	return &Adapter{client: &http.Client{Timeout: timeout}, timeout: timeout}
}

func (a *Adapter) Send(ctx context.Context, record domain.DistributionRecord) (string, error) {
	if record.Target == "" {
		return "", fmt.Errorf("webhook target is empty")
	}
	body, err := json.Marshal(map[string]any{
		"id":             record.ID,
		"service_id":     record.ServiceID,
		"template_id":    record.TemplateID,
		"certificate_id": record.CertificateID,
		"payload":        json.RawMessage(record.PayloadJSON),
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, record.Target, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	return readWebhookResponse(resp)
}
