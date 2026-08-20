package webhook

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/acme/certpilot/internal/distribution/domain"
)

type trackingBody struct {
	data   string
	closed bool
}

func (b *trackingBody) Read(p []byte) (int, error) {
	if b.data == "" {
		return 0, io.EOF
	}
	n := copy(p, b.data)
	b.data = b.data[n:]
	return n, nil
}

func (b *trackingBody) Close() error {
	b.closed = true
	return nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestSendClosesResponseBodyOnSuccessAndHTTPError(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
	}{
		{name: "success", status: http.StatusOK},
		{name: "http error", status: http.StatusBadGateway},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := &trackingBody{data: "upstream detail"}
			adapter := NewAdapter(0)
			adapter.client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.status, Body: body, Header: make(http.Header)}, nil
			})
			_, err := adapter.Send(context.Background(), domain.DistributionRecord{ID: "r1", Target: "http://upstream.invalid", PayloadJSON: "{}"})
			if tc.status >= 300 && err == nil {
				t.Fatal("Send() returned nil error for HTTP failure")
			}
			if tc.status < 300 && err != nil {
				t.Fatalf("Send() returned unexpected error: %v", err)
			}
			if !body.closed {
				t.Fatal("Send() left the response body open")
			}
		})
	}
}

func TestSendPropagatesResponseReadError(t *testing.T) {
	wantErr := errors.New("response read failed")
	body := &errorBody{err: wantErr}
	adapter := NewAdapter(0)
	adapter.client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: body, Header: make(http.Header)}, nil
	})
	_, err := adapter.Send(context.Background(), domain.DistributionRecord{ID: "r1", Target: "http://upstream.invalid", PayloadJSON: "{}"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Send() error = %v, want wrapped response read error", err)
	}
	if !body.closed {
		t.Fatal("Send() left a failed response body open")
	}
}

type errorBody struct {
	err    error
	closed bool
}

func (b *errorBody) Read([]byte) (int, error) { return 0, b.err }
func (b *errorBody) Close() error             { b.closed = true; return nil }
