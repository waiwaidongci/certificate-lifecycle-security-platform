package domain

func IsTerminal(value string) bool {
	switch value {
	case ReminderSent, ReminderFailed:
		return true
	default:
		return false
	}
}
