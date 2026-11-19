package utils

import (
	"strings"
	"testing"
)

func TestGenerateRandomString(t *testing.T) {
	for _, length := range []int{1, 16, 64, 128} {
		got := GenerateRandomString(length)
		if len(got) != length {
			t.Errorf("GenerateRandomString(%d) returned length %d", length, len(got))
		}
		for _, r := range got {
			if !strings.ContainsRune(alphanumericCharset, r) {
				t.Errorf("GenerateRandomString(%d) produced invalid character %q", length, r)
			}
		}
	}
	if GenerateRandomString(8) == GenerateRandomString(8) {
		t.Error("two calls returned identical strings")
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name string
		value string
		valid bool
	}{
		{name: "valid", value: "Password1", valid: true},
		{name: "missing digit", value: "Password", valid: false},
		{name: "too short", value: "Pass1", valid: false},
		{name: "bcrypt byte limit", value: strings.Repeat("a", 72) + "A1", valid: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidatePassword(tt.value); got != tt.valid {
				t.Fatalf("ValidatePassword() = %v, want %v", got, tt.valid)
			}
		})
	}
}
