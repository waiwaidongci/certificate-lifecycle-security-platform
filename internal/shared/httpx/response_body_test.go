package httpx

import (
	"errors"
	"io"
	"testing"
)

type trackingReadCloser struct {
	readErr error
	closed  bool
}

func (b *trackingReadCloser) Read(p []byte) (int, error) {
	if b.readErr != nil {
		return 0, b.readErr
	}
	copy(p, "ok")
	return 2, io.EOF
}

func (b *trackingReadCloser) Close() error {
	b.closed = true
	return nil
}

func TestReadAndCloseClosesUnderlyingBodyOnReadError(t *testing.T) {
	wantErr := errors.New("broken response stream")
	body := &trackingReadCloser{readErr: wantErr}
	_, err := ReadAndClose(body, 64)
	if !errors.Is(err, wantErr) {
		t.Fatalf("ReadAndClose() error = %v, want wrapped read error", err)
	}
	if !body.closed {
		t.Fatal("ReadAndClose() left the response body open")
	}
}
