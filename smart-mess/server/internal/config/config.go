package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	// StaleAfterMinutes is how old a crowd update can be before it's "stale".
	StaleAfterMinutes = 30

	// SessionMaxAge is how long a session cookie lasts.
	SessionMaxAge = 14 * 24 * time.Hour

	// OTP expiry.
	OTPExpiryMinutes = 10

	// OTP rate limits.
	OTPMaxAttempts      = 5
	OTPMaxRequestsPerHr = 5

	// Timezone.
	Timezone = "Asia/Kolkata"
)

// Config holds all environment-driven configuration.
type Config struct {
	DatabaseURL        string
	DatabaseURLTest    string
	JWTSecret          string
	OTPSecret          string
	ClientOrigin       string
	AllowedEmailDomain string
	SMTPHost           string
	SMTPPort           string
	SMTPUser           string
	SMTPPass           string
	MailFrom           string
	AppEnv             string
	Port               string
}

// IsDev returns true when running in development mode.
func (c *Config) IsDev() bool {
	return c.AppEnv != "production"
}

// SMTPConfigured returns true when SMTP credentials are set.
func (c *Config) SMTPConfigured() bool {
	return c.SMTPHost != "" && c.SMTPPort != "" && c.SMTPUser != "" && c.SMTPPass != "" && c.MailFrom != ""
}

// Load reads required env vars and fails fast if any are missing.
func Load() (*Config, error) {
	missing := []string{}
	get := func(key string) string {
		v := os.Getenv(key)
		if v == "" {
			missing = append(missing, key)
		}
		return v
	}
	opt := func(key string) string {
		return os.Getenv(key)
	}

	cfg := &Config{
		DatabaseURL:        get("DATABASE_URL"),
		DatabaseURLTest:    opt("DATABASE_URL_TEST"),
		JWTSecret:          get("JWT_SECRET"),
		OTPSecret:          get("OTP_SECRET"),
		ClientOrigin:       opt("CLIENT_ORIGIN"),
		AllowedEmailDomain: get("ALLOWED_EMAIL_DOMAIN"),
		SMTPHost:           opt("SMTP_HOST"),
		SMTPPort:           opt("SMTP_PORT"),
		SMTPUser:           opt("SMTP_USER"),
		SMTPPass:           opt("SMTP_PASS"),
		MailFrom:           opt("MAIL_FROM"),
		AppEnv:             opt("APP_ENV"),
		Port:               opt("PORT"),
	}

	if cfg.AppEnv == "" {
		cfg.AppEnv = "development"
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.ClientOrigin == "" {
		cfg.ClientOrigin = "http://localhost:5173"
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}
