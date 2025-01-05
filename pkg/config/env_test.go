package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type C struct {
	General struct {
		HealthPort int `default:"13102" env:"HPORT"`
		GRPCPort   int `yaml:"grpcPort" default:"13001"`
		Jaeger     string
	}

	DB struct {
		Webhook struct {
			IdleLifetime time.Duration
		}
	}
}

func TestCollect(t *testing.T) {
	dbCfg := C{}

	configReader := New(WithConfigFile(""))

	err := configReader.Read(&dbCfg)
	require.NoError(t, err)

	entries, err := collect("test", &dbCfg)

	assert.NoError(t, err)

	for _, e := range entries {
		t.Logf("Name: %s, Key: %s, Alt: %s", e.name, e.key, e.alt)
	}

	require.Equal(t, 13102, dbCfg.General.HealthPort)

	os.Setenv("TEST_GENERAL_HPORT", "14101")
	os.Setenv("TEST_DB_WEBHOOK_IDLE_LIFETIME", "1m")

	err = InjectFromEnv("test", &dbCfg)

	require.NoError(t, err)
	require.Equal(t, 14101, dbCfg.General.HealthPort)
	require.Equal(t, time.Minute, dbCfg.DB.Webhook.IdleLifetime)
}
