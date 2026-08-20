package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	certapp "github.com/acme/certpilot/internal/certificate/application"
	"github.com/acme/certpilot/internal/rotation/domain"
	"github.com/acme/certpilot/internal/shared/apperror"
	"github.com/acme/certpilot/internal/shared/clock"
	"github.com/acme/certpilot/internal/shared/database"
	"github.com/acme/certpilot/internal/shared/eventlog/application"
	"github.com/acme/certpilot/internal/shared/httpx"
	"github.com/acme/certpilot/internal/shared/id"
)

type Service struct {
	db           *database.DB
	repository   domain.Repository
	certificates *certapp.Service
	events       *application.Service
	clock        clock.Clock
}

func NewService(db *database.DB, repository domain.Repository, certificates *certapp.Service, events *application.Service, clk clock.Clock) *Service {
	return &Service{db: db, repository: repository, certificates: certificates, events: events, clock: clk}
}

func (s *Service) Generate(ctx context.Context, advanceDays int) ([]domain.Plan, error) {
	if advanceDays <= 0 {
		return nil, apperror.Invalid("advance_days must be positive")
	}
	certificates, err := s.certificates.ListExpiring(ctx, s.clock.Now().AddDate(0, 0, advanceDays))
	if err != nil {
		return nil, err
	}
	created := make([]domain.Plan, 0)
	for _, certificate := range certificates {
		plan, err := s.createPlanForCertificate(ctx, certificate.ID, certificate.ServiceID, advanceDays, certificate.NotAfter)
		if err != nil {
			if isAlreadyExists(err) {
				continue
			}
			return created, err
		}
		created = append(created, plan)
	}
	return created, nil
}

func (s *Service) createPlanForCertificate(ctx context.Context, certificateID string, serviceID *string, advanceDays int, notAfter time.Time) (domain.Plan, error) {
	var plan domain.Plan
	err := s.db.WithTx(ctx, func(tx database.Tx) error {
		if _, getErr := s.repository.GetActiveByCertificate(ctx, tx, certificateID); getErr == nil {
			return apperror.Conflict("active rotation plan already exists")
		} else if !isNotFound(getErr) {
			return getErr
		}
		now := s.clock.Now()
		serviceValue := ""
		if serviceID != nil {
			serviceValue = *serviceID
		}
		plan = domain.Plan{
			ID:            id.New(),
			CertificateID: certificateID,
			ServiceID:     serviceValue,
			DueAt:         now,
			AdvanceDays:   advanceDays,
			Priority:      priorityFromNotAfter(notAfter),
			Status:        domain.PlanPending,
			CreatedAt:     now,
			UpdatedAt:     now,
			Version:       1,
		}
		if err := s.repository.CreatePlan(ctx, tx, plan); err != nil {
			return err
		}
		task := domain.Task{
			ID:         id.New(),
			PlanID:     plan.ID,
			Status:     domain.PlanPending,
			AssignedTo: "",
			Attempts:   0,
			LastError:  "",
			CreatedAt:  now,
			UpdatedAt:  now,
			Version:    1,
		}
		if err := s.repository.CreateTask(ctx, tx, task); err != nil {
			return err
		}
		return s.events.Record(ctx, tx, httpx.ActorFromContext(ctx), "rotation.plan.generated", "rotation_plan", plan.ID, map[string]any{"certificate_id": certificateID, "service_id": serviceValue})
	})
	if err != nil {
		return domain.Plan{}, err
	}
	return plan, nil
}

func (s *Service) GetPlan(ctx context.Context, planID string) (domain.Plan, error) {
	return s.repository.GetPlan(ctx, s.db, planID)
}

func (s *Service) ListPlans(ctx context.Context, options domain.ListOptions) ([]domain.Plan, int, error) {
	return s.repository.ListPlans(ctx, s.db, options)
}

func (s *Service) GetTask(ctx context.Context, taskID string) (domain.Task, error) {
	return s.repository.GetTask(ctx, s.db, taskID)
}

func (s *Service) ListTasks(ctx context.Context, options domain.ListOptions) ([]domain.Task, int, error) {
	return s.repository.ListTasks(ctx, s.db, options)
}

type TransitionCommand struct {
	TargetStatus string `json:"target_status"`
	Error        string `json:"error,omitempty"`
	AssignedTo   string `json:"assigned_to,omitempty"`
	Version      int    `json:"version"`
}

func (s *Service) TransitionTask(ctx context.Context, taskID string, command TransitionCommand) (domain.Task, error) {
	task, err := s.repository.GetTask(ctx, s.db, taskID)
	if err != nil {
		return domain.Task{}, err
	}
	if command.Version != 0 && command.Version != task.Version {
		return domain.Task{}, apperror.Conflict("optimistic lock conflict")
	}
	if !allowedTransition(task.Status, command.TargetStatus) {
		return domain.Task{}, apperror.Conflict(fmt.Sprintf("cannot transition from %s to %s", task.Status, command.TargetStatus))
	}
	now := s.clock.Now()
	task.Status = command.TargetStatus
	task.LastError = command.Error
	if command.AssignedTo != "" {
		task.AssignedTo = command.AssignedTo
	}
	task.UpdatedAt = now
	task.Version++
	if command.TargetStatus == domain.PlanInProgress {
		task.Attempts++
		task.StartedAt = &now
	}
	if command.TargetStatus == domain.PlanCompleted || command.TargetStatus == domain.PlanFailed || command.TargetStatus == domain.PlanSkipped {
		task.CompletedAt = &now
	}
	err = s.db.WithTx(ctx, func(tx database.Tx) error {
		if err := s.repository.UpdateTask(ctx, tx, task); err != nil {
			return err
		}
		plan, err := s.repository.GetPlan(ctx, tx, task.PlanID)
		if err != nil {
			return err
		}
		plan.Status = command.TargetStatus
		plan.UpdatedAt = now
		plan.Version++
		if command.TargetStatus == domain.PlanFailed && command.Error != "" {
			plan.Reason = command.Error
		}
		if err := s.repository.UpdatePlanStatus(ctx, tx, plan); err != nil {
			return err
		}
		return s.events.Record(ctx, tx, httpx.ActorFromContext(ctx), "rotation.task."+command.TargetStatus, "rotation_task", task.ID, map[string]any{"plan_id": task.PlanID})
	})
	if err != nil {
		return domain.Task{}, err
	}
	return task, nil
}

func (s *Service) ExecuteDueTasks(ctx context.Context, limit int) ([]domain.Task, error) {
	if limit <= 0 {
		limit = 20
	}
	tasks, err := s.repository.ListDueTasks(ctx, s.db, s.clock.Now(), limit)
	if err != nil {
		return nil, err
	}
	executed := make([]domain.Task, 0, len(tasks))
	for _, task := range tasks {
		if task.Status != domain.PlanPending {
			continue
		}
		updated, err := s.TransitionTask(ctx, task.ID, TransitionCommand{TargetStatus: domain.PlanInProgress, AssignedTo: "rotator", Version: task.Version})
		if err != nil {
			if errors.As(err, new(*apperror.Error)) {
				continue
			}
			return executed, err
		}
		updated, err = s.TransitionTask(ctx, task.ID, TransitionCommand{TargetStatus: domain.PlanCompleted, AssignedTo: "rotator", Version: updated.Version})
		if err != nil {
			if errors.As(err, new(*apperror.Error)) {
				continue
			}
			return executed, err
		}
		executed = append(executed, updated)
	}
	return executed, nil
}

func allowedTransition(from, to string) bool {
	switch from {
	case domain.PlanPending:
		return to == domain.PlanInProgress || to == domain.PlanSkipped
	case domain.PlanInProgress:
		return to == domain.PlanCompleted || to == domain.PlanFailed
	default:
		return false
	}
}

func priorityFromNotAfter(notAfter time.Time) int {
	days := int(time.Until(notAfter).Hours() / 24)
	switch {
	case days <= 7:
		return 3
	case days <= 30:
		return 2
	default:
		return 1
	}
}

func isNotFound(err error) bool {
	var apiErr *apperror.Error
	return errors.As(err, &apiErr) && apiErr.Code == apperror.CodeNotFound
}

func isAlreadyExists(err error) bool {
	var apiErr *apperror.Error
	return errors.As(err, &apiErr) && apiErr.Code == apperror.CodeConflict
}
