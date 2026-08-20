package log

import (
	"context"
	"encoding/json"

	"github.com/acme/certpilot/internal/distribution/domain"
	"github.com/acme/certpilot/internal/shared/logger"
)

type Adapter struct {
	log *logger.Logger
}

func NewAdapter(log *logger.Logger) *Adapter {
	return &Adapter{log: log}
}

func (a *Adapter) Send(ctx context.Context, _ domain.DistributionRecord) (string, error) {
	a.log.Info(ctx, "distribution delivered through log adapter", "record_id", "")
	return "logged", nil
}

func (a *Adapter) Describe(payload any) string {
	raw, _ := json.Marshal(payload)
	return string(raw)
}
