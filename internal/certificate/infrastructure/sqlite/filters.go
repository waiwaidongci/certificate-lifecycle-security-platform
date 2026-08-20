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
	case "status", "service_id", "common_name", "serial_number", "issuer_id":
		return key
	default:
		return ""
	}
}

func sanitizeSort(value string) string {
	switch value {
	case "common_name", "serial_number", "status", "created_at", "updated_at", "not_after":
		return value
	default:
		return "created_at"
	}
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullableStringPtr(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}
