package domain

func AllowedTransition(from, to string) bool {
	switch from {
	case PlanPending:
		return to == PlanInProgress || to == PlanSkipped
	case PlanInProgress:
		return to == PlanCompleted || to == PlanFailed
	case PlanCompleted, PlanFailed, PlanSkipped:
		return false
	default:
		return false
	}
}

func IsTerminal(status string) bool {
	switch status {
	case PlanCompleted, PlanFailed, PlanSkipped:
		return true
	default:
		return false
	}
}
