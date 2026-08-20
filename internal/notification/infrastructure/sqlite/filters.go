package sqlite

import (
	"strings"
	"time"
)

func buildWhere(filters map[string]string) (string, []any) {
	clauses := make([]string, 0, len(filters))
	args := make([]any, 0, len(filters))
	for key, value := range filters {
		column := sanitizeFilter(key)
		if column == "" {
			continue
		}
		clauses = append(clauses, column+" = ?")
		args = append(args, value)
	}
	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func sanitizeFilter(key string) string {
	switch key {
	case "status", "certificate_id", "service_id", "channel":
		return key
	default:
		return ""
	}
}

func sanitizeSort(value string) string {
	switch value {
	case "status", "days_left", "created_at":
		return value
	default:
		return "created_at"
	}
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.Format(time.RFC3339Nano)
}
