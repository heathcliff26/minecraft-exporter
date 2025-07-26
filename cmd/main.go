package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/heathcliff26/minecraft-exporter/pkg/config"
	"github.com/heathcliff26/minecraft-exporter/pkg/rcon"
	"github.com/heathcliff26/minecraft-exporter/pkg/save"
	"github.com/heathcliff26/minecraft-exporter/pkg/version"
	"github.com/heathcliff26/promremote/promremote"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	configPath  string
	env         bool
	showVersion bool
)

// Initialize the logger
func init() {
	flag.StringVar(&configPath, "config", "", "Optional: Path to config file")
	flag.BoolVar(&env, "env", false, "Used together with -config, when set will expand enviroment variables in config")
	flag.BoolVar(&showVersion, "version", false, "Show the version information and exit")
}

// Handle requests to the webroot.
// Serves static, human-readable HTML that provides a link to /metrics
func ServerRootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<html><body><h1>Welcome to minecraft-exporter</h1>Click <a href='/metrics'>here</a> to see metrics.</body></html>")
}

// handleVersionFlag checks if version flag is set and prints version if needed.
// Returns true if version was printed and program should exit.
func handleVersionFlag() bool {
	if showVersion {
		fmt.Print(version.Version())
		return true
	}
	return false
}

// setupPrometheusRegistry creates and configures a Prometheus registry with collectors.
// Returns the configured registry and any error encountered.
func setupPrometheusRegistry(cfg config.Config) (*prometheus.Registry, error) {
	reg := prometheus.NewRegistry()

	sc, err := save.NewSaveCollector(cfg.WorldDir, cfg.ReduceMetrics)
	if err != nil {
		slog.Error("Failed to create save collector", "err", err)
		return nil, err
	}
	reg.MustRegister(sc)

	if cfg.RCON.Enable {
		rc, err := rcon.NewRCONCollector(cfg)
		if err != nil {
			slog.Error("Failed to create rcon collector", "err", err)
			return nil, err
		}
		defer rc.Close()
		reg.MustRegister(rc)
		sc.RCON = rc.Client()
	}

	return reg, nil
}

// loadConfigWithLogging loads configuration and logs any errors.
// Returns the config and any error encountered.
func loadConfigWithLogging(configPath string, env bool) (config.Config, error) {
	cfg, err := config.LoadConfig(configPath, env)
	if err != nil {
		slog.Error("Could not load configuration", slog.String("path", configPath), slog.String("err", err.Error()))
		return cfg, err
	}
	return cfg, nil
}

// setupRemoteWrite configures remote write client if enabled.
// Returns a cleanup function and any error encountered.
func setupRemoteWrite(cfg config.Config, reg *prometheus.Registry) (func(), error) {
	if !cfg.Remote.Enable {
		return func() {}, nil
	}

	rwClient, err := promremote.NewWriteClient(cfg.Remote.URL, cfg.Remote.Instance, "integrations/minecraft-exporter", reg)
	if err != nil {
		slog.Error("Failed to create remote write client", "err", err)
		return nil, err
	}
	if cfg.Remote.Username != "" {
		err := rwClient.SetBasicAuth(cfg.Remote.Username, cfg.Remote.Password)
		if err != nil {
			slog.Error("Failed to create remote_write client", "err", err)
			return nil, err
		}
	}

	slog.Info("Starting remote_write client")
	rwQuit := make(chan bool)
	rwClient.Run(time.Duration(cfg.Interval), rwQuit)

	return func() {
		rwQuit <- true
		close(rwQuit)
	}, nil
}

func main() {
	flag.Parse()

	if handleVersionFlag() {
		os.Exit(0)
	}

	cfg, err := loadConfigWithLogging(configPath, env)
	if err != nil {
		os.Exit(1)
	}

	reg, err := setupPrometheusRegistry(cfg)
	if err != nil {
		os.Exit(1)
	}

	cleanup, err := setupRemoteWrite(cfg, reg)
	if err != nil {
		os.Exit(1)
	}
	defer cleanup()

	router := http.NewServeMux()
	router.HandleFunc("/", ServerRootHandler)
	router.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))

	server := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	slog.Info("Starting http server", slog.String("addr", server.Addr))
	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Failed to start http server", "err", err)
		os.Exit(1)
	}
}
