package main

import (
	"strings"
	"testing"
)

func TestPropertyRequestPrompt(t *testing.T) {
	tests := []struct {
		name, kind, details, want string
		valid                     bool
	}{
		{"repair", "maintenance", "leaking tap", "maintenance triage note", true},
		{"document", "tenant_document", "lease renewal", "do not invent missing facts", true},
		{"visit", "inspection", "Friday at 9", "inspection reminder", true},
		{"unknown", "billing", "invoice", "", false},
		{"empty details", "maintenance", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := (PropertyRequest{Kind: tt.kind, Property: "Unit 12", Details: tt.details}).prompt()
			if (err == nil) != tt.valid || (tt.valid && !strings.Contains(got, tt.want)) {
				t.Fatalf("prompt=%q error=%v", got, err)
			}
		})
	}
}
