package webhook

import (
	"fmt"
	"net/http"

	"github.com/acme/certpilot/internal/shared/httpx"
)

func readWebhookResponse(resp *http.Response) (string, error) {
	body, readErr := httpx.ReadAndClose(resp.Body, 64*1024)
	if readErr != nil {
		return "", fmt.Errorf("read webhook response: %w", readErr)
	}
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(body))
	}
	return "webhook accepted", nil
}
