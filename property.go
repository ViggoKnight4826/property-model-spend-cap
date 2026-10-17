package main

import (
	"errors"
	"strings"
)

type PropertyRequest struct {
	Kind     string `json:"kind"`
	Property string `json:"property"`
	Details  string `json:"details"`
}

func (r PropertyRequest) prompt() (string, error) {
	if strings.TrimSpace(r.Property) == "" || strings.TrimSpace(r.Details) == "" {
		return "", errors.New("property and details are required")
	}
	switch r.Kind {
	case "maintenance":
		return "Draft a short maintenance triage note for " + r.Property + ": " + r.Details, nil
	case "tenant_document":
		return "Summarize the tenant document for " + r.Property + "; do not invent missing facts: " + r.Details, nil
	case "inspection":
		return "Draft a concise inspection reminder for " + r.Property + ": " + r.Details, nil
	default:
		return "", errors.New("kind must be maintenance, tenant_document, or inspection")
	}
}
