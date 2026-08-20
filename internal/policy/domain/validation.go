package domain

import "strings"

func DomainAllowed(patterns []string, domain string) bool {
	for _, pattern := range patterns {
		if strings.HasPrefix(pattern, "*.") && strings.HasSuffix(domain, strings.TrimPrefix(pattern, "*")) {
			return true
		}
		if pattern == domain {
			return true
		}
	}
	return false
}
