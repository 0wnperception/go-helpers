package config

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/0wnperception/go-helpers/pkg/slice"
	"github.com/stretchr/testify/require"
)

const (
	testConfig = `
general:
  debug: true
  grpcPort: 8080

db:
  accounting:
    host: "localhost:8080"
    timeout: "2m"
`
	mapConfig = `
map:
  value:
    "1": "EUR"
    "2": "USD"
    "3": "RUB"
`
)

func TestConfigOptions(T *testing.T) {
	c := New(WithConfigFile("file.yaml"),
		WithEncPrefix("prefix"),
		WithConfigPath(""),
		WithConfigPath("./clock"),
		WithConfigPath("./cache"),
		WithConfigPath("./clock"))

	cfg, ok := c.(*cfg)

	require.True(T, ok)
	require.Equal(T, "file.yaml", cfg.configFile)
	require.Equal(T, "prefix", cfg.envPrefix)
	require.Equal(T, 2, len(cfg.configPaths))
	f, err := filepath.Abs("./clock")
	require.NoError(T, err)
	require.True(T, slice.Contains(f, cfg.configPaths))
	f, err = filepath.Abs("./cache")
	require.NoError(T, err)
	require.True(T, slice.Contains(f, cfg.configPaths))
}

func TestConfig(t *testing.T) {
	configReader := New(WithReader(strings.NewReader(testConfig)))

	c := CT{}

	err := configReader.Read(&c)

	require.NoError(t, err)
	require.Equal(t, "localhost:8080", c.DB.Accounting.Host)
	require.Equal(t, true, c.General.Debug)
	require.Equal(t, 8080, c.General.GRPCPort)
	require.Equal(t, 2*time.Minute, c.DB.Accounting.Timeout)
	require.Equal(t, 1*time.Minute, c.DB.Accounting.Timeout1)
}

type CT struct {
	General struct {
		Debug    bool
		GRPCPort int `yaml:"grpcPort"`
	}

	DB struct {
		Accounting struct {
			Host     string
			Timeout  time.Duration
			Timeout1 time.Duration `yaml:"tt" default:"1m"`
		}
	}
}

func TestEmptyConfig(t *testing.T) {
	c := CT{}

	cfg := New(WithConfigFile(""))

	err := cfg.Read(&c)
	require.NoError(t, err)
	require.Equal(t, 1*time.Minute, c.DB.Accounting.Timeout1)
}

type MapConfig struct {
	Map struct {
		Value map[string]string
	}
}

func TestReadMap(t *testing.T) {
	configReader := New(WithReader(strings.NewReader(mapConfig)))

	c := MapConfig{}

	err := configReader.Read(&c)

	require.NoError(t, err)

	require.Equal(t, "EUR", c.Map.Value["1"])
	require.Equal(t, "USD", c.Map.Value["2"])
	require.Equal(t, "RUB", c.Map.Value["3"])
}
