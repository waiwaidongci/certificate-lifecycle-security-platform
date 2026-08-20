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
	case "name", "status", "service_id", "template_id", "target_type":
		return key
	default:
		return ""
	}
}

func sanitizeSort(value string) string {
	switch value {
	case "name", "status", "created_at", "updated_at":
		return value
	default:
		return "created_at"
	}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func nullableStringPtr(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.Format(time.RFC3339Nano)
}
