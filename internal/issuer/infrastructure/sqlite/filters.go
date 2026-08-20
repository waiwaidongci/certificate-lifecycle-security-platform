package sqlite

import (
	"strings"

	"github.com/acme/certpilot/internal/shared/apperror"
)

func buildWhere(filters map[string]string) (string, []any, error) {
	clauses := make([]string, 0, len(filters))
	args := make([]any, 0, len(filters))
	for key, value := range filters {
		column := sanitizeFilter(key)
		if column == "" {
			continue
		}
		if key == "enabled" {
			enabled, err := parseEnabledFilter(value)
			if err != nil {
				return "", nil, err
			}
			args = append(args, enabled)
			clauses = append(clauses, column+" = ?")
			continue
		}
		clauses = append(clauses, column+" = ?")
		args = append(args, value)
	}
	if len(clauses) == 0 {
		return "", args, nil
	}
	return " WHERE " + strings.Join(clauses, " AND "), args, nil
}

func parseEnabledFilter(value string) (int, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1":
		return 1, nil
	case "false", "0":
		return 0, nil
	default:
		return 0, apperror.Invalid("enabled filter must be true, false, 1, or 0")
	}
}

func sanitizeFilter(key string) string {
	if key == "name" || key == "provider" || key == "enabled" {
		return key
	}
	return ""
}

func sanitizeSort(value string) string {
	if value == "name" || value == "provider" || value == "updated_at" {
		return value
	}
	return "created_at"
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
