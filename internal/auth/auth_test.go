package auth

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		expectError bool
	}{
		{"valid email", "test@example.com", false},
		{"invalid email", "invalid-email", true},
		{"empty email", "", true},
		{"invalid format 1", "test@.com", true},
		{"invalid format 2", "@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IsValidEmail(tt.email)
			assert.Equal(t, tt.expectError, err != nil, "IsValidEmail(%q) error presence should be %v, got err=%v", tt.email, tt.expectError, err)

		})
	}
}

func TestIsStrongPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{"valid password", "StrongP@ssw0rd", false},
		{"too short", "S@1a", true},
		{"no uppercase", "weakp@ssw0rd", true},
		{"no lowercase", "WEAKP@SSW0RD", true},
		{"no number", "WeakP@ssword", true},
		{"no special char", "WeakPassw0rd", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IsStrongPassword(tt.password)
			assert.Equal(t, tt.expectError, err != nil, "IsStrongPassword(%q) error presence should be %v, got err=%v", tt.password, tt.expectError, err)
		})
	}
}

func TestCheckPriority(t *testing.T) {
	tests := []struct {
		name        string
		priority    string
		expectError error
	}{
		{"valid priority", "high", nil},
		{"valid urgent priority", "urgent", nil},
		{"invalid priority", "critical", fmt.Errorf("Invalid priority value: %s", "critical")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CheckPriority(tt.priority)
			if tt.expectError != nil {
				assert.EqualError(t, err, tt.expectError.Error(), "CheckPriority(%q) error should be %v, got %v", tt.priority, tt.expectError, err)
			} else {
				assert.NoError(t, err, "CheckPriority(%q) error should be nil, got %v", tt.priority, err)
			}
		})
	}
}
