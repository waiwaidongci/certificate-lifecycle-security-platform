package webhook

import "errors"

var ErrWebhookRejected = errors.New("webhook rejected notification")

type DeliveryError struct {
	StatusCode int
}

func (e *DeliveryError) Error() string {
	return "webhook returned status " + httpStatusText(e.StatusCode)
}

func httpStatusText(code int) string {
	if code == 0 {
		return "unknown"
	}
	return formatStatusCode(code)
}

func formatStatusCode(code int) string {
	return string(rune('0'+code/100)) + string(rune('0'+(code/10)%10)) + string(rune('0'+code%10))
}
