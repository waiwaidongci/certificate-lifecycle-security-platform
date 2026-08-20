package application

import (
	"context"

	"github.com/acme/certpilot/internal/notification/domain"
)

func (s *Service) claimPending(id string) bool {
	s.pendingMu.Lock()
	defer s.pendingMu.Unlock()
	if _, exists := s.pendingClaim[id]; exists {
		return false
	}
	s.pendingClaim[id] = struct{}{}
	return true
}

func (s *Service) releasePending(id string) {
	s.pendingMu.Lock()
	delete(s.pendingClaim, id)
	s.pendingMu.Unlock()
}

func (s *Service) processPending(ctx context.Context, reminder domain.Reminder) (domain.Reminder, error) {
	defer s.releasePending(reminder.ID)
	adapter, ok := s.adapters[reminder.Channel]
	if !ok {
		reminder.Status = domain.ReminderFailed
		reminder.Message = "unsupported notification channel"
	} else if response, sendErr := adapter.Send(ctx, reminder); sendErr != nil {
		reminder.Status = domain.ReminderFailed
		reminder.Message = sendErr.Error()
	} else {
		reminder.Status = domain.ReminderSent
		reminder.Message = response
		now := s.clock.Now()
		reminder.SentAt = &now
	}
	if err := s.repository.UpdateStatus(ctx, s.db, reminder); err != nil {
		return domain.Reminder{}, err
	}
	return reminder, nil
}
