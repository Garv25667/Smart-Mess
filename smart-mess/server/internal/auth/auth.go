// Package auth provides email validation, OTP hashing, and JWT session management.
package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// --- Email domain validation ---

// ValidateEmailDomain checks that the email belongs to the allowed domain or a subdomain.
// Rules: lowercase and trim. Accept if the part after "@" equals allowedDomain or ends with "."+allowedDomain.
func ValidateEmailDomain(email, allowedDomain string) error {
	e := strings.TrimSpace(strings.ToLower(email))
	parts := strings.SplitN(e, "@", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("invalid email format")
	}
	domain := parts[1]
	if domain == allowedDomain {
		return nil
	}
	if strings.HasSuffix(domain, "."+allowedDomain) {
		return nil
	}
	return fmt.Errorf("email domain not allowed")
}

// NormalizeEmail lowercases and trims an email address.
func NormalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

// --- OTP ---

// GenerateOTPCode returns a random 6-digit code string.
func GenerateOTPCode() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// HashOTP returns an HMAC-SHA256 hex digest of the code using the given secret.
func HashOTP(code, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(code))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyOTPHash performs constant-time comparison of a code against a stored hash.
func VerifyOTPHash(code, storedHash, secret string) bool {
	computedHash := HashOTP(code, secret)
	return subtle.ConstantTimeCompare([]byte(computedHash), []byte(storedHash)) == 1
}

// --- JWT ---

// Claims for the session JWT.
type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"uid"`
	Role   string `json:"role"`
	Email  string `json:"email"`
}

// CreateJWT signs a new JWT with the given claims.
func CreateJWT(userID, role, email, secret string, maxAge time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(maxAge)),
		},
		UserID: userID,
		Role:   role,
		Email:  email,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseJWT parses and validates a JWT string.
func ParseJWT(tokenStr, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// SetSessionCookie writes an httpOnly JWT cookie.
func SetSessionCookie(w http.ResponseWriter, token string, maxAge time.Duration, secure bool) {
	cookie := &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		MaxAge:   int(maxAge.Seconds()),
	}
	http.SetCookie(w, cookie)
}

// ClearSessionCookie clears the session cookie.
func ClearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
}

// GetSessionCookie reads the session cookie value from the request.
func GetSessionCookie(r *http.Request) (string, error) {
	c, err := r.Cookie("session")
	if err != nil {
		return "", err
	}
	return c.Value, nil
}
