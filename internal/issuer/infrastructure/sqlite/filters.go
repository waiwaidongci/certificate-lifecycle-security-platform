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
			normalized := strings.ToLower(strings.TrimSpace(value))
			switch normalized {
			case "true", "1":
				args = append(args, 1)
			case "false", "0":
				args = append(args, 0)
			default:
				return "", nil, apperror.Invalid("invalid enabled filter value: " + value)
			}
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
