package domain

import "strings"

const (
	TargetLog     = "log"
	TargetWebhook = "webhook"
)

func NormalizeTargetType(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func ValidTargetType(value string) bool {
	return value == TargetLog || value == TargetWebhook
}
