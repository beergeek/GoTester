package main

import (
	"testing"
)

// TestAnalyzePayload tests the analyzePayload function
func TestAnalyzePayload(t *testing.T) {
	tests := []struct {
		name     string
		payload  []byte
		expected PayloadType
	}{
		// Positive Test Cases
		{"No Payload", []byte(""), NoPayload},
		{"Generic Payload", []byte("some generic content"), GenericPayload},
		{"Potential Malicious Payload - 1", []byte("this is a malware"), PotentialMalicious},
		{"Potential Malicious Payload - 2", []byte("this is a virus"), PotentialMalicious},
		{"Potential Malicious Payload - 3", []byte("some exploit"), PotentialMalicious},
		{"Potential Malicious Payload - 4", []byte("sophisticated attack"), PotentialMalicious},
		{"Potential Malicious Payload - 5", []byte("java code in payload"), PotentialMalicious},
		{"Potential Malicious Payload - 6", []byte("kubernetes config"), PotentialMalicious},

		// Negative Test Cases
		{"Non-malicious payload 1", []byte("some safe text"), GenericPayload},
		{"Non-malicious payload 2", []byte("another example of safe text"), GenericPayload},
		{"Case Insensitivity Check", []byte("MALWARE PayLoad"), PotentialMalicious},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzePayload(tt.payload)
			if result != tt.expected {
				t.Errorf("analyzePayload() = %s; expected %s", result, tt.expected)
			}
		})
	}
}
