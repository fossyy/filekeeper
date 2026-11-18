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
