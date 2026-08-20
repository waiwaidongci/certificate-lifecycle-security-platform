package sqlite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/acme/certpilot/internal/rotation/domain"
	"github.com/acme/certpilot/internal/shared/apperror"
	"github.com/acme/certpilot/internal/shared/database"
)

type Repository struct{}

func NewRepository() *Repository { return &Repository{} }

func (r *Repository) CreatePlan(ctx context.Context, exec database.Executor, plan domain.Plan) error {
	_, err := exec.ExecContext(ctx, `INSERT INTO rotation_plans (id, certificate_id, service_id, due_at, advance_days, priority, status, reason, created_at, updated_at, version) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		plan.ID, plan.CertificateID, plan.ServiceID, plan.DueAt.Format(time.RFC3339Nano), plan.AdvanceDays, plan.Priority, plan.Status, plan.Reason, plan.CreatedAt.Format(time.RFC3339Nano), plan.UpdatedAt.Format(time.RFC3339Nano), plan.Version)
	if err != nil {
		return fmt.Errorf("insert rotation plan: %w", err)
	}
	return nil
}

func (r *Repository) GetPlan(ctx context.Context, exec database.Executor, id string) (domain.Plan, error) {
	return scanPlan(exec.QueryRowContext(ctx, `SELECT id, certificate_id, service_id, due_at, advance_days, priority, status, reason, created_at, updated_at, version FROM rotation_plans WHERE id = ?`, id))
}

func (r *Repository) ListPlans(ctx context.Context, exec database.Executor, options domain.ListOptions) ([]domain.Plan, int, error) {
	where, args := buildWhere(options.Filters)
	var total int
	if err := exec.QueryRowContext(ctx, "SELECT COUNT(*) FROM rotation_plans"+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count rotation plans: %w", err)
	}
	sort := sanitizeSort(options.Sort)
	order := strings.ToUpper(options.Order)
	if order != "DESC" {
		order = "ASC"
	}
	rows, err := exec.QueryContext(ctx, "SELECT id, certificate_id, service_id, due_at, advance_days, priority, status, reason, created_at, updated_at, version FROM rotation_plans"+where+" ORDER BY "+sort+" "+order+" LIMIT ? OFFSET ?", append(args, options.PageSize, (options.Page-1)*options.PageSize)...)
	if err != nil {
		return nil, 0, fmt.Errorf("list rotation plans: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Plan, 0, options.PageSize)
	for rows.Next() {
		item, err := scanPlan(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *Repository) UpdatePlanStatus(ctx context.Context, exec database.Executor, plan domain.Plan) error {
	result, err := exec.ExecContext(ctx, `UPDATE rotation_plans SET status = ?, reason = ?, updated_at = ?, version = ? WHERE id = ? AND version = ?`,
		plan.Status, plan.Reason, plan.UpdatedAt.Format(time.RFC3339Nano), plan.Version, plan.ID, plan.Version-1)
	if err != nil {
		return fmt.Errorf("update rotation plan: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return apperror.Conflict("rotation plan was modified concurrently")
	}
	return nil
}

func (r *Repository) GetActiveByCertificate(ctx context.Context, exec database.Executor, certificateID string) (domain.Plan, error) {
	return scanPlan(exec.QueryRowContext(ctx, `SELECT id, certificate_id, service_id, due_at, advance_days, priority, status, reason, created_at, updated_at, version FROM rotation_plans WHERE certificate_id = ? AND status IN (?, ?) ORDER BY created_at DESC LIMIT 1`, certificateID, domain.PlanPending, domain.PlanInProgress))
}

func (r *Repository) CreateTask(ctx context.Context, exec database.Executor, task domain.Task) error {
	_, err := exec.ExecContext(ctx, `INSERT INTO rotation_tasks (id, plan_id, status, assigned_to, attempts, last_error, started_at, completed_at, created_at, updated_at, version) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		task.ID, task.PlanID, task.Status, task.AssignedTo, task.Attempts, task.LastError, nullableTime(task.StartedAt), nullableTime(task.CompletedAt), task.CreatedAt.Format(time.RFC3339Nano), task.UpdatedAt.Format(time.RFC3339Nano), task.Version)
	if err != nil {
		return fmt.Errorf("insert rotation task: %w", err)
	}
	return nil
}

func (r *Repository) GetTask(ctx context.Context, exec database.Executor, id string) (domain.Task, error) {
	return scanTask(exec.QueryRowContext(ctx, `SELECT id, plan_id, status, assigned_to, attempts, last_error, started_at, completed_at, created_at, updated_at, version FROM rotation_tasks WHERE id = ?`, id))
}

func (r *Repository) ListTasks(ctx context.Context, exec database.Executor, options domain.ListOptions) ([]domain.Task, int, error) {
	where, args := buildWhere(options.Filters)
	var total int
	if err := exec.QueryRowContext(ctx, "SELECT COUNT(*) FROM rotation_tasks"+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count rotation tasks: %w", err)
	}
	sort := sanitizeSort(options.Sort)
	order := strings.ToUpper(options.Order)
	if order != "DESC" {
		order = "ASC"
	}
	rows, err := exec.QueryContext(ctx, "SELECT id, plan_id, status, assigned_to, attempts, last_error, started_at, completed_at, created_at, updated_at, version FROM rotation_tasks"+where+" ORDER BY "+sort+" "+order+" LIMIT ? OFFSET ?", append(args, options.PageSize, (options.Page-1)*options.PageSize)...)
	if err != nil {
		return nil, 0, fmt.Errorf("list rotation tasks: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Task, 0, options.PageSize)
	for rows.Next() {
		item, err := scanTask(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *Repository) ListDueTasks(ctx context.Context, exec database.Executor, before time.Time, limit int) ([]domain.Task, error) {
	rows, err := exec.QueryContext(ctx, `SELECT t.id, t.plan_id, t.status, t.assigned_to, t.attempts, t.last_error, t.started_at, t.completed_at, t.created_at, t.updated_at, t.version FROM rotation_tasks t JOIN rotation_plans p ON p.id = t.plan_id WHERE t.status = ? AND p.due_at <= ? ORDER BY p.priority DESC, p.due_at ASC LIMIT ?`, domain.PlanPending, before.Format(time.RFC3339Nano), limit)
	if err != nil {
		return nil, fmt.Errorf("list due tasks: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Task, 0, limit)
	for rows.Next() {
		item, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) UpdateTask(ctx context.Context, exec database.Executor, task domain.Task) error {
	result, err := exec.ExecContext(ctx, `UPDATE rotation_tasks SET status = ?, assigned_to = ?, attempts = ?, last_error = ?, started_at = ?, completed_at = ?, updated_at = ?, version = ? WHERE id = ? AND version = ?`,
		task.Status, task.AssignedTo, task.Attempts, task.LastError, nullableTime(task.StartedAt), nullableTime(task.CompletedAt), task.UpdatedAt.Format(time.RFC3339Nano), task.Version, task.ID, task.Version-1)
	if err != nil {
		return fmt.Errorf("update rotation task: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return apperror.Conflict("rotation task was modified concurrently")
	}
	return nil
}
