package auth

import (
	"testing"
	"time"
)

func TestValidateEmailDomain(t *testing.T) {
	allowed := "iiitp.ac.in"
	tests := []struct {
		email   string
		wantErr bool
	}{
		// From ARCHITECTURE.md section 7
		{"a@iiitp.ac.in", false},
		{"a@cse.iiitp.ac.in", false},
		{"a@gmail.com", true},
		{"a@eviliiitp.ac.in", true},
		{"a@iiitp.ac.in.evil.com", true},
		// Additional cases
		{"invalidemail", true},
		{"", true},
		{"  User@CSE.IIITP.AC.IN  ", false}, // trimmed + lowered
		{"112400000@cse.iiitp.ac.in", false},
		{"user@ece.iiitp.ac.in", false},
	}
	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			err := ValidateEmailDomain(tt.email, allowed)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmailDomain(%q) error = %v, wantErr %v", tt.email, err, tt.wantErr)
			}
		})
	}
}

func TestNormalizeEmail(t *testing.T) {
	if got := NormalizeEmail("  FOO@Bar.COM  "); got != "foo@bar.com" {
		t.Errorf("got %q", got)
	}
}

func TestOTPHashRoundtrip(t *testing.T) {
	secret := "test-secret"
	code := "123456"
	hash := HashOTP(code, secret)
	if !VerifyOTPHash(code, hash, secret) {
		t.Error("hash roundtrip failed")
	}
	if VerifyOTPHash("654321", hash, secret) {
		t.Error("wrong code should not verify")
	}
}

func TestGenerateOTPCode(t *testing.T) {
	code, err := GenerateOTPCode()
	if err != nil {
		t.Fatal(err)
	}
	if len(code) != 6 {
		t.Errorf("expected 6 digits, got %q", code)
	}
}

func TestJWTRoundtrip(t *testing.T) {
	secret := "test-jwt-secret"
	token, err := CreateJWT("user-123", "student", "a@iiitp.ac.in", secret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := ParseJWT(token, secret)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("UserID = %q", claims.UserID)
	}
	if claims.Role != "student" {
		t.Errorf("Role = %q", claims.Role)
	}
	if claims.Email != "a@iiitp.ac.in" {
		t.Errorf("Email = %q", claims.Email)
	}
}

func TestJWTWrongSecret(t *testing.T) {
	token, _ := CreateJWT("u", "student", "a@b.com", "secret1", time.Hour)
	_, err := ParseJWT(token, "secret2")
	if err == nil {
		t.Error("expected error for wrong secret")
	}
}
