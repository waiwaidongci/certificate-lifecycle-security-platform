package domain

func AllowedTransition(from, to string) bool {
	switch from {
	case PlanPending:
		return to == PlanInProgress || to == PlanSkipped
	case PlanInProgress:
		return to == PlanCompleted || to == PlanFailed
	default:
		return false
	}
}

func IsTerminal(status string) bool {
	return status == PlanCompleted || status == PlanFailed || status == PlanSkipped
}
