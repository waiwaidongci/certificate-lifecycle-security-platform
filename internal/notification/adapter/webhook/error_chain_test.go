package webhook

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/acme/certpilot/internal/notification/domain"
)

func TestSendPreservesWebhookRejectionChain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	_, err := NewAdapter(0).Send(context.Background(), domain.Reminder{ID: "r-012", Recipient: server.URL})
	var deliveryErr *DeliveryError
	if !errors.As(err, &deliveryErr) {
		t.Fatalf("Send error = %v, want DeliveryError", err)
	}
	if !errors.Is(err, ErrWebhookRejected) {
		t.Fatalf("Send error = %v, want ErrWebhookRejected in chain", err)
	}
}
