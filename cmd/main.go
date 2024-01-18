package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/heathcliff26/containers/apps/minecraft-exporter/pkg/config"
	"github.com/heathcliff26/containers/apps/minecraft-exporter/pkg/rcon"
	"github.com/heathcliff26/containers/apps/minecraft-exporter/pkg/save"
	"github.com/heathcliff26/promremote/promremote"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	configPath string
	env        bool
)

// Initialize the logger
func init() {
	flag.StringVar(&configPath, "config", "", "Optional: Path to config file")
	flag.BoolVar(&env, "env", false, "Used together with -config, when set will expand enviroment variables in config")
}

// Handle requests to the webroot.
// Serves static, human-readable HTML that provides a link to /metrics
func ServerRootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<html><body><h1>Welcome to minecraft-exporter</h1>Click <a href='/metrics'>here</a> to see metrics.</body></html>")
}

func main() {
	flag.Parse()

	cfg, err := config.LoadConfig(configPath, env)
	if err != nil {
		slog.Error("Could not load configuration", slog.String("path", configPath), slog.String("err", err.Error()))
		os.Exit(1)
	}

	reg := prometheus.NewRegistry()

	if cfg.RCON.Enable {
		rc, err := rcon.NewRCONCollector(cfg)
		if err != nil {
			slog.Error("Failed to create rcon collector", "err", err)
			os.Exit(1)
		}
		defer rc.Close()
		reg.MustRegister(rc)
	}

	sc, err := save.NewSaveCollector(cfg.WorldDir, cfg.ReduceMetrics)
	if err != nil {
		slog.Error("Failed to create save collector", "err", err)
		os.Exit(1)
	}
	reg.MustRegister(sc)

	if cfg.Remote.Enable {
		rwClient, err := promremote.NewWriteClient(cfg.Remote.URL, cfg.Remote.Instance, "integrations/minecraft-exporter", reg)
		if err != nil {
			slog.Error("Failed to create remote write client", "err", err)
			os.Exit(1)
		}
		if cfg.Remote.Username != "" {
			err := rwClient.SetBasicAuth(cfg.Remote.Username, cfg.Remote.Password)
			if err != nil {
				slog.Error("Failed to create remote_write client", "err", err)
				os.Exit(1)
			}
		}

		slog.Info("Starting remote_write client")
		rwQuit := make(chan bool)
		rwClient.Run(cfg.Interval, rwQuit)
		defer func() {
			rwQuit <- true
			close(rwQuit)
		}()
	}

	http.HandleFunc("/", ServerRootHandler)
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))

	addr := ":" + strconv.Itoa(cfg.Port)
	slog.Info("Starting http server", slog.String("addr", addr))
	err = http.ListenAndServe(addr, nil)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Failed to start http server", "err", err)
		os.Exit(1)
	}
}
