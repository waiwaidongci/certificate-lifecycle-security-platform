package domain

func IsValidStatus(value string) bool {
	switch value {
	case StatusIssued, StatusRevoked:
		return true
	case StatusExpired:
		return false
	default:
		return false
	}
}
