package auth

import (
	"bytes"
	"strings"
	"testing"
)

// --- GenerateSalt ---

func TestGenerateSalt_UniqueOnEachCall(t *testing.T) {
	t.Parallel()
	a := GenerateSalt()
	b := GenerateSalt()
	if bytes.Equal(a, b) {
		t.Error("expected different salts on consecutive calls, got identical values")
	}
}

// --- ParseParams ---

func TestParseParams_ValidParams(t *testing.T) {
	t.Parallel()
	m, tt, p, err := ParseParams("m=65536,t=1,p=4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m != 65536 {
		t.Errorf("expected m=65536, got %d", m)
	}
	if tt != 1 {
		t.Errorf("expected t=1, got %d", tt)
	}
	if p != 4 {
		t.Errorf("expected p=4, got %d", p)
	}
}

func TestParseParams_MissingParam(t *testing.T) {
	t.Parallel()
	// p is omitted
	_, _, _, err := ParseParams("m=65536,t=1")
	if err == nil {
		t.Error("expected error for missing parameter, got nil")
	}
}

func TestParseParams_UnknownParam(t *testing.T) {
	t.Parallel()
	_, _, _, err := ParseParams("m=65536,t=1,p=4,x=9")
	if err == nil {
		t.Error("expected error for unknown parameter, got nil")
	}
}

func TestParseParams_InvalidValue(t *testing.T) {
	t.Parallel()
	_, _, _, err := ParseParams("m=notanumber,t=1,p=4")
	if err == nil {
		t.Error("expected error for invalid parameter value, got nil")
	}
}

// --- GenerateString ---

func TestGenerateString_ReturnsNonEmpty(t *testing.T) {
	t.Parallel()
	result := GenerateString("password")
	if result == "" {
		t.Error("expected non-empty hash string")
	}
}

func TestGenerateString_ContainsArgon2idID(t *testing.T) {
	t.Parallel()
	result := GenerateString("password")
	if !strings.Contains(result, "argon2id") {
		t.Errorf("expected hash to contain %q, got %q", "argon2id", result)
	}
}

func TestGenerateString_DifferentSaltsEachCall(t *testing.T) {
	t.Parallel()
	a := GenerateString("password")
	b := GenerateString("password")
	if a == b {
		t.Error("expected different hash strings on consecutive calls with same password")
	}
}

// --- ValidateString ---

func TestValidateString_CorrectPassword(t *testing.T) {
	t.Parallel()
	hash := GenerateString("correct-horse-battery-staple")
	if !ValidateString("correct-horse-battery-staple", hash) {
		t.Error("expected ValidateString to return true for correct password")
	}
}

func TestValidateString_WrongPassword(t *testing.T) {
	t.Parallel()
	hash := GenerateString("correct-horse-battery-staple")
	if ValidateString("wrong-password", hash) {
		t.Error("expected ValidateString to return false for wrong password")
	}
}

func TestValidateString_InvalidHashString(t *testing.T) {
	t.Parallel()
	if ValidateString("password", "this-is-not-a-valid-phc-string") {
		t.Error("expected ValidateString to return false for invalid hash string")
	}
}

func TestValidateString_TamperedOutput(t *testing.T) {
	t.Parallel()
	hash := GenerateString("password")
	// Replace the last character to corrupt the output segment.
	last := hash[len(hash)-1]
	replacement := byte('A')
	if last == 'A' {
		replacement = 'B'
	}
	tampered := hash[:len(hash)-1] + string(replacement)
	if ValidateString("password", tampered) {
		t.Error("expected ValidateString to return false for tampered hash")
	}
}
