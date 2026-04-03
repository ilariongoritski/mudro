// Package config provides environment helpers shared across all services.
package config

import "os"

// EnvOr returns the value of the environment variable named by key,
// or def if the variable is unset or empty.
func EnvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
