package domain

import (
	"context"
	"time"

	"github.com/acme/certpilot/internal/shared/database"
)

const (
	PlanPending    = "pending"
	PlanInProgress = "in_progress"
	PlanCompleted  = "completed"
	PlanFailed     = "failed"
	PlanSkipped    = "skipped"
)

type Plan struct {
	ID            string    `json:"id"`
	CertificateID string    `json:"certificate_id"`
	ServiceID     string    `json:"service_id"`
	DueAt         time.Time `json:"due_at"`
	AdvanceDays   int       `json:"advance_days"`
	Priority      int       `json:"priority"`
	Status        string    `json:"status"`
	Reason        string    `json:"reason"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Version       int       `json:"version"`
}

type Task struct {
	ID          string     `json:"id"`
	PlanID      string     `json:"plan_id"`
	Status      string     `json:"status"`
	AssignedTo  string     `json:"assigned_to"`
	Attempts    int        `json:"attempts"`
	LastError   string     `json:"last_error"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Version     int        `json:"version"`
}

type Repository interface {
	CreatePlan(ctx context.Context, exec database.Executor, plan Plan) error
	GetPlan(ctx context.Context, exec database.Executor, id string) (Plan, error)
	ListPlans(ctx context.Context, exec database.Executor, options ListOptions) ([]Plan, int, error)
	UpdatePlanStatus(ctx context.Context, exec database.Executor, plan Plan) error
	GetActiveByCertificate(ctx context.Context, exec database.Executor, certificateID string) (Plan, error)
	CreateTask(ctx context.Context, exec database.Executor, task Task) error
	GetTask(ctx context.Context, exec database.Executor, id string) (Task, error)
	ListTasks(ctx context.Context, exec database.Executor, options ListOptions) ([]Task, int, error)
	ListDueTasks(ctx context.Context, exec database.Executor, before time.Time, limit int) ([]Task, error)
	UpdateTask(ctx context.Context, exec database.Executor, task Task) error
}

type ListOptions struct {
	Page     int
	PageSize int
	Filters  map[string]string
	Sort     string
	Order    string
}
