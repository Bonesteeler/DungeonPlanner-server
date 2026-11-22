package service

import (
	"errors"
	"strings"
	"testing"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{name: "minimum length", password: "Aa1!aaaa", valid: true},
		{name: "maximum length", password: strings.Repeat("a", 124) + "Aa1!", valid: true},
		{name: "unicode character classes", password: "Ää1!password", valid: true},
		{name: "empty", password: "", valid: false},
		{name: "too short", password: "Aa1!aaa", valid: false},
		{name: "too long", password: strings.Repeat("a", 125) + "Aa1!", valid: false},
		{name: "missing uppercase", password: "password1!", valid: false},
		{name: "missing lowercase", password: "PASSWORD1!", valid: false},
		{name: "missing number", password: "Password!", valid: false},
		{name: "missing special", password: "Password1", valid: false},
		{name: "space is not special", password: "Password1 ", valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.password)
			if tt.valid && err != nil {
				t.Fatalf("validatePassword() returned error: %v", err)
			}
			if !tt.valid && !errors.Is(err, ErrInvalidPassword) {
				t.Errorf("validatePassword() error = %v, want ErrInvalidPassword", err)
			}
		})
	}
}
 