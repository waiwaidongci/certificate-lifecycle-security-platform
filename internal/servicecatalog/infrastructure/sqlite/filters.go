package sqlite

import "strings"

func buildWhere(filters map[string]string) (string, []any) {
	clauses := make([]string, 0, len(filters))
	args := make([]any, 0, len(filters))
	for key, value := range filters {
		column := sanitizeColumn(key)
		if column == "" {
			continue
		}
		if key == "enabled" {
			if value == "true" || value == "1" {
				args = append(args, 1)
			} else {
				args = append(args, 0)
			}
			clauses = append(clauses, column+" = ?")
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

func sanitizeColumn(key string) string {
	switch key {
	case "name", "environment", "domain", "owner", "region", "enabled":
		return key
	default:
		return ""
	}
}

func allowedSort(sort string) string {
	switch sort {
	case "name", "environment", "domain", "owner", "region", "updated_at":
		return sort
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

func nullableString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}
