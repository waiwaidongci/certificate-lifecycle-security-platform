package domain

func IsTerminal(value string) bool {
	switch value {
	case ReminderSent:
		return true
	case ReminderFailed:
		return false
	default:
		return false
	}
}
