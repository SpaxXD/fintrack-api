package config

import (
	"os"
	"testing"
	"time"
)

// setRequiredEnv sets the minimum required environment variables for a valid config.
func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb?sslmode=disable")
	t.Setenv("JWT_SECRET", "this-is-a-secret-key-that-is-at-least-32-chars-long")
}

func TestLoad_ValidConfig_Defaults(t *testing.T) {
	setRequiredEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Check defaults
	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Server.ShutdownTimeout != 30*time.Second {
		t.Errorf("expected shutdown timeout 30s, got %v", cfg.Server.ShutdownTimeout)
	}
	if cfg.JWT.AccessTokenExpiry != 15*time.Minute {
		t.Errorf("expected access token expiry 15m, got %v", cfg.JWT.AccessTokenExpiry)
	}
	if cfg.JWT.RefreshTokenExpiry != 168*time.Hour {
		t.Errorf("expected refresh token expiry 168h, got %v", cfg.JWT.RefreshTokenExpiry)
	}
	if cfg.Log.Level != "info" {
		t.Errorf("expected log level info, got %q", cfg.Log.Level)
	}
	if cfg.Migration.AutoMigrate != false {
		t.Errorf("expected auto migrate false, got %v", cfg.Migration.AutoMigrate)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("SERVER_PORT", "9090")
	t.Setenv("SHUTDOWN_TIMEOUT", "10s")
	t.Setenv("ACCESS_TOKEN_EXPIRY", "30m")
	t.Setenv("REFRESH_TOKEN_EXPIRY", "72h")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("AUTO_MIGRATE", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}
	if cfg.Server.ShutdownTimeout != 10*time.Second {
		t.Errorf("expected shutdown timeout 10s, got %v", cfg.Server.ShutdownTimeout)
	}
	if cfg.JWT.AccessTokenExpiry != 30*time.Minute {
		t.Errorf("expected access token expiry 30m, got %v", cfg.JWT.AccessTokenExpiry)
	}
	if cfg.JWT.RefreshTokenExpiry != 72*time.Hour {
		t.Errorf("expected refresh token expiry 72h, got %v", cfg.JWT.RefreshTokenExpiry)
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("expected log level debug, got %q", cfg.Log.Level)
	}
	if cfg.Migration.AutoMigrate != true {
		t.Errorf("expected auto migrate true, got %v", cfg.Migration.AutoMigrate)
	}
}

func TestLoad_InvalidPort_TooLow(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("SERVER_PORT", "80")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for port < 1024")
	}
	if !containsAll(err.Error(), "SERVER_PORT", "1024", "65535") {
		t.Errorf("error should mention SERVER_PORT and constraints, got: %v", err)
	}
}

func TestLoad_InvalidPort_TooHigh(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("SERVER_PORT", "70000")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for port > 65535")
	}
	if !containsAll(err.Error(), "SERVER_PORT", "1024", "65535") {
		t.Errorf("error should mention SERVER_PORT and constraints, got: %v", err)
	}
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	t.Setenv("JWT_SECRET", "this-is-a-secret-key-that-is-at-least-32-chars-long")
	// DATABASE_URL not set - clear it explicitly
	os.Unsetenv("DATABASE_URL")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing DATABASE_URL")
	}
	if !containsAll(err.Error(), "DATABASE_URL") {
		t.Errorf("error should mention DATABASE_URL, got: %v", err)
	}
}

func TestLoad_JWTSecretTooShort(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb?sslmode=disable")
	t.Setenv("JWT_SECRET", "short")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for JWT_SECRET < 32 chars")
	}
	if !containsAll(err.Error(), "JWT_SECRET", "32") {
		t.Errorf("error should mention JWT_SECRET and 32 char minimum, got: %v", err)
	}
}

func TestLoad_InvalidAccessTokenExpiry(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("ACCESS_TOKEN_EXPIRY", "-5m")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for negative ACCESS_TOKEN_EXPIRY")
	}
	if !containsAll(err.Error(), "ACCESS_TOKEN_EXPIRY", "positive") {
		t.Errorf("error should mention ACCESS_TOKEN_EXPIRY and positive, got: %v", err)
	}
}

func TestLoad_InvalidRefreshTokenExpiry(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("REFRESH_TOKEN_EXPIRY", "-1h")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for negative REFRESH_TOKEN_EXPIRY")
	}
	if !containsAll(err.Error(), "REFRESH_TOKEN_EXPIRY", "positive") {
		t.Errorf("error should mention REFRESH_TOKEN_EXPIRY and positive, got: %v", err)
	}
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("LOG_LEVEL", "verbose")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid LOG_LEVEL")
	}
	if !containsAll(err.Error(), "LOG_LEVEL") {
		t.Errorf("error should mention LOG_LEVEL, got: %v", err)
	}
}

func TestLoad_PortBoundary_1024(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("SERVER_PORT", "1024")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("port 1024 should be valid, got error: %v", err)
	}
	if cfg.Server.Port != 1024 {
		t.Errorf("expected port 1024, got %d", cfg.Server.Port)
	}
}

func TestLoad_PortBoundary_65535(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("SERVER_PORT", "65535")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("port 65535 should be valid, got error: %v", err)
	}
	if cfg.Server.Port != 65535 {
		t.Errorf("expected port 65535, got %d", cfg.Server.Port)
	}
}

func TestLoad_JWTSecretExactly32Chars(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb?sslmode=disable")
	t.Setenv("JWT_SECRET", "12345678901234567890123456789012") // exactly 32 chars

	cfg, err := Load()
	if err != nil {
		t.Fatalf("JWT secret with exactly 32 chars should be valid, got error: %v", err)
	}
	if cfg.JWT.Secret != "12345678901234567890123456789012" {
		t.Errorf("unexpected JWT secret value")
	}
}

// containsAll checks that s contains all the given substrings.
func containsAll(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if !contains(s, sub) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsSubstring(s, sub))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
