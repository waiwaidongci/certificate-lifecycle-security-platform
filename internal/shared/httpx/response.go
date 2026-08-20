package httpx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/acme/certpilot/internal/shared/apperror"
)

// ReadAndClose consumes a bounded response body and reports read failures.
func ReadAndClose(body io.ReadCloser, limit int) ([]byte, error) {
	return readAndCloseResponse(body, limit)
}

func readAndCloseResponse(body io.ReadCloser, limit int) ([]byte, error) {
	if body == nil {
		return nil, fmt.Errorf("response body is nil")
	}
	if limit <= 0 {
		limit = 64 * 1024
	}
	data, err := io.ReadAll(io.LimitReader(body, int64(limit)+1))
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	if len(data) > limit {
		return nil, fmt.Errorf("response body exceeds %d bytes", limit)
	}
	return []byte(strings.TrimSpace(string(data))), nil
}

type ErrorResponse struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	RequestID string         `json:"request_id"`
	Details   map[string]any `json:"details,omitempty"`
}

type Meta struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
}

type ListResponse struct {
	Data any  `json:"data"`
	Meta Meta `json:"meta"`
}

func WriteJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func WriteCreated(w http.ResponseWriter, value any) {
	WriteJSON(w, http.StatusCreated, value)
}

func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func WriteError(ctx context.Context, w http.ResponseWriter, err error) {
	apiErr, ok := apperror.As(err)
	if !ok {
		apiErr = apperror.Internal(err)
	}
	status := apperror.HTTPStatus(apiErr.Code)
	requestID := RequestIDFromContext(ctx)
	WriteJSON(w, status, ErrorResponse{
		Code:      string(apiErr.Code),
		Message:   apiErr.Message,
		RequestID: requestID,
		Details:   apiErr.Details,
	})
}

type contextKey string

const requestIDKey contextKey = "request_id"

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func RequestIDFromContext(ctx context.Context) string {
	if value, ok := ctx.Value(requestIDKey).(string); ok {
		return value
	}
	return ""
}

func RequestStartFromContext(ctx context.Context) time.Time {
	if value, ok := ctx.Value("request_start").(time.Time); ok {
		return value
	}
	return time.Now()
}
