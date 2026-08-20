package domain

import "strings"

type DispatchKey struct {
	ServiceID     string
	TemplateID    string
	CertificateID string
	TargetType    string
	Target        string
}

func NewDispatchKey(serviceID, templateID, certificateID, targetType, target string) DispatchKey {
	return DispatchKey{
		ServiceID:     strings.TrimSpace(serviceID),
		TemplateID:    strings.TrimSpace(templateID),
		CertificateID: strings.TrimSpace(certificateID),
		TargetType:    strings.ToLower(strings.TrimSpace(targetType)),
		Target:        strings.TrimSpace(target),
	}
}
