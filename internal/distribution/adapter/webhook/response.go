package webhook

import (
	"fmt"
	"net/http"

	"github.com/acme/certpilot/internal/shared/httpx"
)

func consumeWebhookResponse(resp *http.Response) (string, error) {
	if resp == nil {
		return "", fmt.Errorf("webhook returned no response")
	}
	body, readErr := httpx.ReadAndClose(resp.Body, 64*1024)
	if readErr != nil {
		return "", fmt.Errorf("read webhook response: %w", readErr)
	}
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(body))
	}
	return "webhook accepted", nil
}
