package query

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/acme/certpilot/internal/shared/apperror"
)

type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type Sort struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

type ListOptions struct {
	Pagination Pagination
	Filters    map[string]string
	Sort       Sort
}

func ParseListOptions(r *http.Request, allowedFilters map[string]bool, allowedSorts map[string]bool, defaultSort string) (ListOptions, error) {
	q := r.URL.Query()
	page := parseIntDefault(q.Get("page"), 1)
	pageSize := parseIntDefault(q.Get("page_size"), 20)
	if page < 1 {
		return ListOptions{}, apperror.Invalid("page must be greater than zero")
	}
	if pageSize < 1 || pageSize > 200 {
		return ListOptions{}, apperror.Invalid("page_size must be between 1 and 200")
	}
	filters := make(map[string]string)
	for key, values := range q {
		if len(values) == 0 || key == "page" || key == "page_size" || key == "sort" || strings.HasPrefix(key, "sort_") {
			continue
		}
		if allowedFilters != nil && !allowedFilters[key] {
			return ListOptions{}, apperror.Invalid("unsupported filter: " + key)
		}
		filters[key] = values[0]
	}
	sortField := q.Get("sort")
	if sortField == "" {
		sortField = defaultSort
	}
	direction := strings.ToLower(q.Get("sort_dir"))
	if direction == "" {
		direction = "asc"
	}
	if direction != "asc" && direction != "desc" {
		return ListOptions{}, apperror.Invalid("sort_dir must be asc or desc")
	}
	if allowedSorts != nil && !allowedSorts[sortField] {
		return ListOptions{}, apperror.Invalid("unsupported sort field: " + sortField)
	}
	return ListOptions{
		Pagination: Pagination{Page: page, PageSize: pageSize},
		Filters:    filters,
		Sort:       Sort{Field: sortField, Direction: direction},
	}, nil
}

func parseIntDefault(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
