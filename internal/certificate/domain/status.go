package domain

func IsValidStatus(value string) bool {
	switch value {
	case StatusIssued, StatusRevoked, StatusExpired:
		return true
	default:
		return false
	}
}
