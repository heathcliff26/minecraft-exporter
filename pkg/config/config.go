package config

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/heathcliff26/promremote/v2/promremote"
	"go.yaml.in/yaml/v3"
)

const (
	DEFAULT_LOG_LEVEL       = "info"
	DEFAULT_PORT            = 8080
	DEFAULT_INTERVAL        = 1 * time.Minute
	DEFAULT_WORLD_DIR       = "/world"
	DEFAULT_REMOTE_JOB_NAME = "minecraft-exporter"
)

const (
	SERVER_TYPE_VANILLA  = "vanilla"
	SERVER_TYPE_FORGE    = "forge"
	SERVER_TYPE_PAPER    = "paper"
	SERVER_TYPE_NEOFORGE = "neoforge"
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
	LogLevel      string        `yaml:"logLevel,omitempty"`
	Port          int           `yaml:"port,omitempty"`
	Interval      time.Duration `yaml:"interval,omitempty"`
	Instance      string        `yaml:"instance"`
	ReduceMetrics bool          `yaml:"reduceMetrics,omitempty"`
	ServerType    string        `yaml:"server,omitempty"`
	DynmapEnabled bool          `yaml:"dynmap,omitempty"`
	WorldDir      string        `yaml:"world,omitempty"`
	RCON          RCONConfig    `yaml:"rcon,omitempty"`
	Remote        RemoteConfig  `yaml:"remote,omitempty"`
}

type RCONConfig struct {
	Enable   bool   `yaml:"enable"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
}

type RemoteConfig struct {
	Enable   bool   `yaml:"enable"`
	URL      string `yaml:"url"`
	Instance string `yaml:"instance,omitempty"`
	JobName  string `yaml:"jobName,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

// Returns a Config with default values set
func DefaultConfig() Config {
	hostname, err := os.Hostname()
	if err != nil {
		slog.Error("Failed to retrieve hostname, using localhost instead", "err", err)
		hostname = "localhost"
	}
	return Config{
		LogLevel:   DEFAULT_LOG_LEVEL,
		Port:       DEFAULT_PORT,
		Interval:   DEFAULT_INTERVAL,
		Instance:   hostname,
		ServerType: SERVER_TYPE_VANILLA,
		WorldDir:   DEFAULT_WORLD_DIR,
		Remote:     defaultRemoteConfig(),
	}
}

func defaultRemoteConfig() RemoteConfig {
	return RemoteConfig{
		JobName: DEFAULT_REMOTE_JOB_NAME,
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

	// #nosec G304: Local users can decide on their file path themselves.
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
	if c.ServerType != SERVER_TYPE_VANILLA && c.ServerType != SERVER_TYPE_FORGE && c.ServerType != SERVER_TYPE_PAPER && c.ServerType != SERVER_TYPE_NEOFORGE {
		return Config{}, &ErrUnknownServerType{Type: c.ServerType}
	}

	if c.Remote.Instance == "" {
		c.Remote.Instance = c.Instance
	}

	if c.Remote.Enable {
		if c.Remote.URL == "" {
			return Config{}, promremote.ErrMissingEndpoint{}
		}
		if c.Remote.Username != c.Remote.Password && (c.Remote.Username == "" || c.Remote.Password == "") {
			return Config{}, promremote.ErrMissingAuthCredentials{}
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
