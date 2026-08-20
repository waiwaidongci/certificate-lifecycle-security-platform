package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/acme/certpilot/internal/rotation/domain"
	"github.com/acme/certpilot/internal/shared/apperror"
)

type scanner interface {
	Scan(dest ...any) error
}

func scanPlan(row scanner) (domain.Plan, error) {
	var plan domain.Plan
	var dueAt, createdAt, updatedAt string
	if err := row.Scan(&plan.ID, &plan.CertificateID, &plan.ServiceID, &dueAt, &plan.AdvanceDays, &plan.Priority, &plan.Status, &plan.Reason, &createdAt, &updatedAt, &plan.Version); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Plan{}, apperror.NotFound("rotation plan not found")
		}
		return domain.Plan{}, fmt.Errorf("scan rotation plan: %w", err)
	}
	plan.DueAt, _ = time.Parse(time.RFC3339Nano, dueAt)
	plan.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	plan.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return plan, nil
}

func scanTask(row scanner) (domain.Task, error) {
	var task domain.Task
	var startedAt, completedAt, createdAt, updatedAt sql.NullString
	if err := row.Scan(&task.ID, &task.PlanID, &task.Status, &task.AssignedTo, &task.Attempts, &task.LastError, &startedAt, &completedAt, &createdAt, &updatedAt, &task.Version); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Task{}, apperror.NotFound("rotation task not found")
		}
		return domain.Task{}, fmt.Errorf("scan rotation task: %w", err)
	}
	task.StartedAt = parseNullableTime(startedAt)
	task.CompletedAt = parseNullableTime(completedAt)
	task.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt.String)
	task.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt.String)
	return task, nil
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.Format(time.RFC3339Nano)
}

func parseNullableTime(value sql.NullString) *time.Time {
	if !value.Valid {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value.String)
	if err != nil {
		return nil
	}
	return &parsed
}
