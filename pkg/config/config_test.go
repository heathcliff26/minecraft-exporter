package config

import (
	"log/slog"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidConfigs(t *testing.T) {
	c1 := Config{
		LogLevel:   "warn",
		Port:       80,
		Interval:   Duration(5 * time.Minute),
		ServerType: SERVER_TYPE_VANILLA,
		WorldDir:   "/path/to/world",
		RCON: RCONConfig{
			Enable:   true,
			Host:     "localhost",
			Port:     25575,
			Password: "password",
		},
		Remote: RemoteConfig{
			JobName: DEFAULT_REMOTE_JOB_NAME,
		},
	}
	c2 := Config{
		LogLevel:   "debug",
		Port:       2080,
		Interval:   Duration(30 * time.Minute),
		ServerType: SERVER_TYPE_VANILLA,
		WorldDir:   DEFAULT_WORLD_DIR,
		Remote: RemoteConfig{
			Enable:   true,
			URL:      "https://example.org/",
			Instance: "test",
			JobName:  "testjob",
			Username: "somebody",
			Password: "somebody's password",
		},
	}
	c3 := Config{
		LogLevel:   "error",
		Port:       DEFAULT_PORT,
		Interval:   DEFAULT_INTERVAL,
		ServerType: SERVER_TYPE_VANILLA,
		WorldDir:   DEFAULT_WORLD_DIR,
		Remote: RemoteConfig{
			Enable:   true,
			URL:      "https://example.org/",
			Instance: "test",
			JobName:  "testjob",
		},
	}
	tMatrix := []struct {
		Name, Path string
		Result     Config
	}{
		{
			Name:   "EmptyConfig",
			Path:   "",
			Result: DefaultConfig(),
		},
		{
			Name:   "Config1",
			Path:   "testdata/valid-config-1.yaml",
			Result: c1,
		},
		{
			Name:   "Config2",
			Path:   "testdata/valid-config-2.yaml",
			Result: c2,
		},
		{
			Name:   "Config3",
			Path:   "testdata/valid-config-3.yaml",
			Result: c3,
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			c, err := LoadConfig(tCase.Path, false)

			assert := assert.New(t)

			if !assert.NoError(err) {
				t.Fatalf("Failed to load config: %v", err)
			}
			assert.Equal(tCase.Result, c)
		})
	}
}

func TestInvalidConfig(t *testing.T) {
	tMatrix := []struct {
		Name, Path, Mode, Error string
	}{
		{
			Name:  "InvalidPath",
			Path:  "file-does-not-exist.yaml",
			Error: "*fs.PathError",
		},
		{
			Name:  "NotYaml",
			Path:  "testdata/not-a-config.txt",
			Error: "*fmt.wrapError",
		},
		{
			Name:  "InvalidInterval",
			Path:  "testdata/invalid-config-1.yaml",
			Error: "*fmt.wrapError",
		},
		{
			Name:  "MissingRemoteEndpoint",
			Path:  "testdata/invalid-config-2.yaml",
			Error: "promremote.ErrMissingEndpoint",
		},
		{
			Name:  "IncompleteRemoteCredentials",
			Path:  "testdata/invalid-config-3.yaml",
			Error: "promremote.ErrMissingAuthCredentials",
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			_, err := LoadConfig(tCase.Path, false)

			if !assert.Error(t, err) {
				t.Fatal("Did not receive an error")
			}
			if !assert.Equal(t, tCase.Error, reflect.TypeOf(err).String()) {
				t.Fatalf("Received invalid error: %v", err)
			}
		})
	}
}

func TestEnvSubstitution(t *testing.T) {
	c := Config{
		LogLevel:   "debug",
		Port:       2080,
		Interval:   Duration(time.Minute),
		ServerType: SERVER_TYPE_VANILLA,
		WorldDir:   "/some/server/world",
		Remote: RemoteConfig{
			JobName: DEFAULT_REMOTE_JOB_NAME,
		},
	}
	t.Setenv("MINECRAFT_EXPORTER_LOG_LEVEL", c.LogLevel)
	t.Setenv("MINECRAFT_EXPORTER_PORT", strconv.Itoa(c.Port))
	t.Setenv("MINECRAFT_EXPORTER_CACHE", c.Interval.String())
	t.Setenv("MINECRAFT_EXPORTER_WORLD_DIR", c.WorldDir)

	res, err := LoadConfig("testdata/env-config.yaml", true)

	assert := assert.New(t)

	assert.NoError(err)
	assert.Equal(c, res)
}

func TestSetLogLevel(t *testing.T) {
	tMatrix := []struct {
		Name  string
		Level slog.Level
		Error error
	}{
		{"debug", slog.LevelDebug, nil},
		{"info", slog.LevelInfo, nil},
		{"warn", slog.LevelWarn, nil},
		{"error", slog.LevelError, nil},
		{"DEBUG", slog.LevelDebug, nil},
		{"INFO", slog.LevelInfo, nil},
		{"WARN", slog.LevelWarn, nil},
		{"ERROR", slog.LevelError, nil},
		{"Unknown", 0, &ErrUnknownLogLevel{"Unknown"}},
	}
	t.Cleanup(func() {
		err := setLogLevel(DEFAULT_LOG_LEVEL)
		if err != nil {
			t.Fatalf("Failed to cleanup after test: %v", err)
		}
	})

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			err := setLogLevel(tCase.Name)

			assert := assert.New(t)

			if !assert.Equal(tCase.Error, err) {
				t.Fatalf("Received invalid error: %v", err)
			}
			if err == nil {
				assert.Equal(tCase.Level, logLevel.Level())
			}
		})
	}
}
