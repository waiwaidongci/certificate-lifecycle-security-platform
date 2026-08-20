package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/acme/certpilot/internal/issuer/domain"
	"github.com/acme/certpilot/internal/shared/apperror"
)

func TestIssuerFilterPreservesEnabledBooleanValues(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  int
	}{
		{name: "trimmed true", value: " true ", want: 1},
		{name: "false", value: "false", want: 0},
		{name: "one", value: "1", want: 1},
		{name: "zero", value: "0", want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			where, args, err := buildWhere(map[string]string{"enabled": tt.value})
			if err != nil {
				t.Fatalf("buildWhere returned error for %q: %v", tt.value, err)
			}
			if where != " WHERE enabled = ?" {
				t.Fatalf("where = %q, want enabled predicate", where)
			}
			if len(args) != 1 || args[0] != tt.want {
				t.Fatalf("args = %#v, want [%d]", args, tt.want)
			}
		})
	}
}

func TestIssuerFilterRejectsMalformedEnabled(t *testing.T) {
	_, _, err := buildWhere(map[string]string{"enabled": "sometimes"})
	if err == nil {
		t.Fatal("buildWhere accepted malformed enabled filter")
	}
	var apiErr *apperror.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %T %v, want apperror.Error", err, err)
	}
	if apiErr.Code != apperror.CodeInvalidArgument {
		t.Fatalf("error code = %s, want %s", apiErr.Code, apperror.CodeInvalidArgument)
	}
}

func TestIssuerRepositoryPreservesFilterValidationError(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, _, err = (&Repository{}).List(context.Background(), db, domain.ListOptions{
		Page:     1,
		PageSize: 20,
		Filters:  map[string]string{"enabled": "sometimes"},
	})
	if err == nil {
		t.Fatal("List accepted malformed enabled filter")
	}
	var apiErr *apperror.Error
	if !errors.As(err, &apiErr) || apiErr.Code != apperror.CodeInvalidArgument {
		t.Fatalf("List error = %T %v, want INVALID_ARGUMENT", err, err)
	}
}
