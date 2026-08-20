package domain

func IsTerminal(value string) bool {
	return value == ReminderSent || value == ReminderFailed
}
