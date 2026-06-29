package config

import (
	"fmt"
	"strings"
	"testing"

	"pgregory.net/rapid"
)

// **Validates: Requirements 10.4**
// Property 16: Configuração inválida rejeita inicialização
//
// For any value of configuration that violates its constraints (port outside 1024-65535,
// JWT secret < 32 characters, non-positive expiration times), the Load() function must
// return an error indicating the field name and the violated constraint.

const validDatabaseURL = "postgres://user:pass@localhost:5432/testdb?sslmode=disable"
const validJWTSecret = "this-is-a-secret-key-that-is-at-least-32-chars-long!!"

// validPortGen generates random ports in the valid range [1024, 65535].
func validPortGen() *rapid.Generator[int] {
	return rapid.IntRange(1024, 65535)
}

// validJWTSecretGen generates random strings with length >= 32.
func validJWTSecretGen() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		length := rapid.IntRange(32, 128).Draw(t, "secretLen")
		chars := make([]byte, length)
		for i := range chars {
			chars[i] = byte(rapid.IntRange(33, 126).Draw(t, "char"))
		}
		return string(chars)
	})
}

// validDurationStrGen generates valid positive duration strings.
func validDurationStrGen() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		value := rapid.IntRange(1, 1000).Draw(t, "durationValue")
		unit := rapid.SampledFrom([]string{"s", "m", "h"}).Draw(t, "durationUnit")
		return fmt.Sprintf("%d%s", value, unit)
	})
}

// setValidEnv sets all environment variables to valid values.
func setValidEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", validDatabaseURL)
	t.Setenv("JWT_SECRET", validJWTSecret)
	t.Setenv("SERVER_PORT", "8080")
	t.Setenv("SHUTDOWN_TIMEOUT", "30s")
	t.Setenv("ACCESS_TOKEN_EXPIRY", "15m")
	t.Setenv("REFRESH_TOKEN_EXPIRY", "168h")
	t.Setenv("LOG_LEVEL", "info")
}

func TestProperty_ValidConfigsAlwaysSucceed(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		port := validPortGen().Draw(rt, "port")
		secret := validJWTSecretGen().Draw(rt, "secret")
		accessExpiry := validDurationStrGen().Draw(rt, "accessExpiry")
		refreshExpiry := validDurationStrGen().Draw(rt, "refreshExpiry")
		shutdownTimeout := validDurationStrGen().Draw(rt, "shutdownTimeout")
		logLevel := rapid.SampledFrom([]string{"debug", "info", "warn", "error"}).Draw(rt, "logLevel")

		t.Setenv("DATABASE_URL", validDatabaseURL)
		t.Setenv("JWT_SECRET", secret)
		t.Setenv("SERVER_PORT", fmt.Sprintf("%d", port))
		t.Setenv("SHUTDOWN_TIMEOUT", shutdownTimeout)
		t.Setenv("ACCESS_TOKEN_EXPIRY", accessExpiry)
		t.Setenv("REFRESH_TOKEN_EXPIRY", refreshExpiry)
		t.Setenv("LOG_LEVEL", logLevel)

		cfg, err := Load()
		if err != nil {
			rt.Fatalf("valid config should not fail: port=%d, secret_len=%d, access=%s, refresh=%s, shutdown=%s, log=%s, err=%v",
				port, len(secret), accessExpiry, refreshExpiry, shutdownTimeout, logLevel, err)
		}
		if cfg == nil {
			rt.Fatal("expected non-nil config for valid inputs")
		}
		if cfg.Server.Port != port {
			rt.Fatalf("expected port %d, got %d", port, cfg.Server.Port)
		}
	})
}

func TestProperty_InvalidPortAlwaysRejected(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate ports outside [1024, 65535]
		port := rapid.OneOf(
			rapid.IntRange(0, 1023),
			rapid.IntRange(65536, 100000),
		).Draw(rt, "invalidPort")

		setValidEnv(t)
		t.Setenv("SERVER_PORT", fmt.Sprintf("%d", port))

		_, err := Load()
		if err == nil {
			rt.Fatalf("expected error for invalid port %d, got nil", port)
		}
		errMsg := err.Error()
		if !strings.Contains(errMsg, "SERVER_PORT") {
			rt.Fatalf("error should mention SERVER_PORT for port=%d, got: %s", port, errMsg)
		}
	})
}

func TestProperty_ShortJWTSecretAlwaysRejected(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate strings shorter than 32 characters
		length := rapid.IntRange(0, 31).Draw(rt, "secretLen")
		chars := make([]byte, length)
		for i := range chars {
			chars[i] = byte(rapid.IntRange(33, 126).Draw(rt, "char"))
		}
		shortSecret := string(chars)

		setValidEnv(t)
		t.Setenv("JWT_SECRET", shortSecret)

		_, err := Load()
		if err == nil {
			rt.Fatalf("expected error for JWT_SECRET of length %d, got nil", length)
		}
		errMsg := err.Error()
		if !strings.Contains(errMsg, "JWT_SECRET") {
			rt.Fatalf("error should mention JWT_SECRET for len=%d, got: %s", length, errMsg)
		}
		if !strings.Contains(errMsg, "32") {
			rt.Fatalf("error should mention 32-char minimum for len=%d, got: %s", length, errMsg)
		}
	})
}

func TestProperty_NonPositiveAccessTokenExpiryAlwaysRejected(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate non-positive durations (negative or zero)
		value := rapid.IntRange(-1000, -1).Draw(rt, "negativeValue")
		unit := rapid.SampledFrom([]string{"s", "m", "h"}).Draw(rt, "unit")
		negativeDuration := fmt.Sprintf("%d%s", value, unit)

		setValidEnv(t)
		t.Setenv("ACCESS_TOKEN_EXPIRY", negativeDuration)

		_, err := Load()
		if err == nil {
			rt.Fatalf("expected error for ACCESS_TOKEN_EXPIRY=%s, got nil", negativeDuration)
		}
		errMsg := err.Error()
		if !strings.Contains(errMsg, "ACCESS_TOKEN_EXPIRY") {
			rt.Fatalf("error should mention ACCESS_TOKEN_EXPIRY for value=%s, got: %s", negativeDuration, errMsg)
		}
	})
}

func TestProperty_NonPositiveRefreshTokenExpiryAlwaysRejected(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate non-positive durations (negative or zero)
		value := rapid.IntRange(-1000, -1).Draw(rt, "negativeValue")
		unit := rapid.SampledFrom([]string{"s", "m", "h"}).Draw(rt, "unit")
		negativeDuration := fmt.Sprintf("%d%s", value, unit)

		setValidEnv(t)
		t.Setenv("REFRESH_TOKEN_EXPIRY", negativeDuration)

		_, err := Load()
		if err == nil {
			rt.Fatalf("expected error for REFRESH_TOKEN_EXPIRY=%s, got nil", negativeDuration)
		}
		errMsg := err.Error()
		if !strings.Contains(errMsg, "REFRESH_TOKEN_EXPIRY") {
			rt.Fatalf("error should mention REFRESH_TOKEN_EXPIRY for value=%s, got: %s", negativeDuration, errMsg)
		}
	})
}
