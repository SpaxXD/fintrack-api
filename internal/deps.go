//go:build deps
// +build deps

// Package internal exists to ensure project dependencies remain in go.mod.
// This file is never compiled into the binary.
package internal

import (
	_ "github.com/go-chi/chi/v5"
	_ "github.com/go-playground/validator/v10"
	_ "github.com/golang-jwt/jwt/v5"
	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/google/uuid"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
	_ "github.com/rs/zerolog"
	_ "github.com/spf13/viper"
	_ "github.com/swaggo/http-swagger/v2"
	_ "github.com/swaggo/swag"
	_ "golang.org/x/crypto/bcrypt"
	_ "pgregory.net/rapid"
)
