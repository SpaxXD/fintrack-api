//go:build tools
// +build tools

// Package tools tracks tool dependencies that are not directly imported in source code.
package tools

import (
	_ "github.com/golang-migrate/migrate/v4/cmd/migrate"
	_ "github.com/swaggo/swag/cmd/swag"
)
