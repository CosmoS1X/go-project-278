package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	envDatabaseURL  = "DATABASE_URL"
	envBaseShortURL = "BASE_SHORT_URL"
	envServerPort   = "SERVER_PORT"
	envCORSOrigin   = "CORS_ORIGIN"

	testDatabaseURL  = "postgres://localhost:5432/app"
	testBaseShortURL = "https://sho.rt"
)

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	prev, existed := os.LookupEnv(key)
	require.NoError(t, os.Unsetenv(key))
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, prev)
		}
	})
}

func setRequiredEnv(t *testing.T) {
	t.Helper()

	t.Setenv(envDatabaseURL, testDatabaseURL)
	t.Setenv(envBaseShortURL, testBaseShortURL)
}

func TestLoadDefaults(t *testing.T) {
	setRequiredEnv(t)
	unsetEnv(t, envServerPort)
	unsetEnv(t, envCORSOrigin)

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, testDatabaseURL, cfg.DatabaseURL)
	assert.Equal(t, testBaseShortURL, cfg.BaseShortURL)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "http://localhost:5173", cfg.CORSOrigin)
}

func TestLoadOverrides(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv(envServerPort, "9090")
	t.Setenv(envCORSOrigin, "https://app.io")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "9090", cfg.Port)
	assert.Equal(t, "https://app.io", cfg.CORSOrigin)
}

func TestLoadMissingRequiredVar(t *testing.T) {
	setRequiredEnv(t)
	unsetEnv(t, envDatabaseURL)

	cfg, err := Load()
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), envDatabaseURL)
}
