package sqlite

import (
	"fmt"
	"testing"

	"github.com/acme/certpilot/internal/shared/apperror"
)

func TestScanPolicyRejectsNilRow(t *testing.T) {
	tests := []struct {
		name string
		row  scanner
	}{
		{name: "nil interface"},
		{name: "typed nil", row: (*nilPolicyRow)(nil)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := scanWithoutPanic(test.row)
			appErr, ok := apperror.As(err)
			if !ok || appErr.Code != apperror.CodeInternal {
				t.Fatalf("nil row returned an uncontrolled error: %v", err)
			}
		})
	}
}

type nilPolicyRow struct{}

func (*nilPolicyRow) Scan(...any) error {
	panic("nil policy row was scanned")
}

func scanWithoutPanic(row scanner) (policyValue any, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("scan panicked: %v", recovered)
		}
	}()
	return scanPolicy(row)
}
