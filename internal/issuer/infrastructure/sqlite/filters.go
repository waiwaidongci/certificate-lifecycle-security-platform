package sqlite

import "strings"

func buildWhere(filters map[string]string) (string, []any) {
	clauses := make([]string, 0, len(filters))
	args := make([]any, 0, len(filters))
	for key, value := range filters {
		column := sanitizeFilter(key)
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
