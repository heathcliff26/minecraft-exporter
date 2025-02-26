package config

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/heathcliff26/promremote/promremote"
	"sigs.k8s.io/yaml"
)

const (
	DEFAULT_LOG_LEVEL   = "info"
	DEFAULT_PORT        = 8080
	DEFAULT_INTERVAL    = Duration(1 * time.Minute)
	DEFAULT_WORLD_DIR   = "/world"
	SERVER_TYPE_VANILLA = "vanilla"
	SERVER_TYPE_FORGE   = "forge"
	SERVER_TYPE_PAPER   = "paper"
)

var logLevel *slog.LevelVar

// Initialize the logger
func init() {
	logLevel = &slog.LevelVar{}
	opts := slog.HandlerOptions{
		Level: logLevel,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &opts))
	slog.SetDefault(logger)
}

type Config struct {
	LogLevel      string       `json:"logLevel,omitempty"`
	Port          int          `json:"port,omitempty"`
	Interval      Duration     `json:"interval,omitempty"`
	ReduceMetrics bool         `json:"reduceMetrics,omitempty"`
	ServerType    string       `json:"server,omitempty"`
	DynmapEnabled bool         `json:"dynmap,omitempty"`
	WorldDir      string       `json:"world,omitempty"`
	RCON          RCONConfig   `json:"rcon,omitempty"`
	Remote        RemoteConfig `json:"remote,omitempty"`
}

type RCONConfig struct {
	Enable   bool   `json:"enable"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
}

type RemoteConfig struct {
	Enable   bool   `json:"enable"`
	URL      string `json:"url"`
	Instance string `json:"instance"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// Returns a Config with default values set
func DefaultConfig() Config {
	return Config{
		LogLevel:   DEFAULT_LOG_LEVEL,
		Port:       DEFAULT_PORT,
		Interval:   DEFAULT_INTERVAL,
		ServerType: SERVER_TYPE_VANILLA,
		WorldDir:   DEFAULT_WORLD_DIR,
	}
}

// Loads config from file, returns error if config is invalid
// Arguments:
//
//	path: Path to config file
//	env: Determines if enviroment variables in the file will be expanded before decoding
//
// RCON Parameters are validated inside the RCON package itself, so it is not checked here.
func LoadConfig(path string, env bool) (Config, error) {
	c := DefaultConfig()

	if path == "" {
		_ = setLogLevel(DEFAULT_LOG_LEVEL)
		return c, nil
	}

	f, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	if env {
		f = []byte(os.ExpandEnv(string(f)))
	}

	err = yaml.Unmarshal(f, &c)
	if err != nil {
		return Config{}, err
	}

	err = setLogLevel(c.LogLevel)
	if err != nil {
		return Config{}, err
	}
	if c.ServerType != SERVER_TYPE_VANILLA && c.ServerType != SERVER_TYPE_FORGE && c.ServerType != SERVER_TYPE_PAPER {
		return Config{}, &ErrUnknownServerType{Type: c.ServerType}
	}

	if c.Remote.Enable {
		if c.Remote.URL == "" {
			return Config{}, promremote.ErrMissingEndpoint{}
		}
		if c.Remote.Username != c.Remote.Password && (c.Remote.Username == "" || c.Remote.Password == "") {
			return Config{}, promremote.ErrMissingAuthCredentials{}
		}
		if c.Remote.Instance == "" {
			slog.Info("No instance name provided, defaulting to hostname")
			hostname, err := os.Hostname()
			if err != nil {
				slog.Error("Failed to retrieve hostname, using localhost instead", "err", err)
				hostname = "localhost"
			}
			c.Remote.Instance = hostname
		}
	}

	return c, nil
}

// Parse a given string and set the resulting log level
func setLogLevel(level string) error {
	switch strings.ToLower(level) {
	case "debug":
		logLevel.Set(slog.LevelDebug)
	case "info":
		logLevel.Set(slog.LevelInfo)
	case "warn":
		logLevel.Set(slog.LevelWarn)
	case "error":
		logLevel.Set(slog.LevelError)
	default:
		return &ErrUnknownLogLevel{level}
	}
	return nil
}
