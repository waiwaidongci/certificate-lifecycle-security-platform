package webhook

import (
	"errors"
	"fmt"
	"net/http"
)

var ErrWebhookRejected = errors.New("webhook rejected notification")

type DeliveryError struct {
	StatusCode int
	Operation  string
}

func NewDeliveryError(operation string, statusCode int) *DeliveryError {
	return &DeliveryError{Operation: operation, StatusCode: statusCode}
}

func (e *DeliveryError) Error() string {
	operation := e.Operation
	if operation == "" {
		operation = "send notification webhook"
	}
	return fmt.Sprintf("%s: webhook returned status %d (%s)", operation, e.StatusCode, http.StatusText(e.StatusCode))
}

func (e *DeliveryError) Unwrap() error { return ErrWebhookRejected }

func (e *DeliveryError) Is(target error) bool {
	return target == ErrWebhookRejected
}

func (e *DeliveryError) Status() int { return e.StatusCode }
