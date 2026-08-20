package http_test

import (
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	certificatehttp "github.com/acme/certpilot/internal/certificate/adapter/http"
	distributionhttp "github.com/acme/certpilot/internal/distribution/adapter/http"
	issuerhttp "github.com/acme/certpilot/internal/issuer/adapter/http"
	notificationhttp "github.com/acme/certpilot/internal/notification/adapter/http"
	policyhttp "github.com/acme/certpilot/internal/policy/adapter/http"
)

func TestMalformedCreateRequestsKeepInvalidArgumentResponse(t *testing.T) {
	tests := []struct {
		name   string
		handle func(stdhttp.ResponseWriter, *stdhttp.Request)
	}{
		{name: "certificate issue", handle: new(certificatehttp.Handler).Issue},
		{name: "issuer create", handle: new(issuerhttp.Handler).Create},
		{name: "policy create", handle: new(policyhttp.Handler).Create},
		{name: "template create", handle: new(distributionhttp.Handler).CreateTemplate},
		{name: "notification scan", handle: new(notificationhttp.Handler).Scan},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(stdhttp.MethodPost, "/", strings.NewReader("{"))
			test.handle(recorder, request)

			if recorder.Code != stdhttp.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body=%s", recorder.Code, stdhttp.StatusBadRequest, recorder.Body.String())
			}
			var body struct {
				Code string `json:"code"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body.Code != "INVALID_ARGUMENT" {
				t.Fatalf("error code = %q, want INVALID_ARGUMENT", body.Code)
			}
		})
	}
}
