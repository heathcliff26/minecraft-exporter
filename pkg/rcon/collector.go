package rcon

import (
	"log/slog"

	"github.com/heathcliff26/containers/apps/minecraft-exporter/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
)

type RCONCollector struct {
	rcon          *RCONClient
	ServerType    string
	DynmapEnabled bool
}

var (
	mcPlayerOnlineDesc = prometheus.NewDesc("minecraft_player_online", "Show currently online players. Value is always 1", []string{"player"}, nil)

	forgeTPSDimDesc          = prometheus.NewDesc("forge_tps_dim", "TPS of a dimension", []string{"dimension_id", "dimension_name"}, nil)
	forgeTicktimeDimDesc     = prometheus.NewDesc("forge_ticktime_dim", "Time a Tick took in a Dimension", []string{"dimension_id", "dimension_name"}, nil)
	forgeTPSOverallDesc      = prometheus.NewDesc("forge_tps_overall", "Overall TPS", nil, nil)
	forgeTicktimeOverallDesc = prometheus.NewDesc("forge_ticktime_overall", "Overall Ticktime", nil, nil)
	forgeEntitiesCountDesc   = prometheus.NewDesc("forge_entity_count", "Type and count of active entites", []string{"entity"}, nil)

	paperTPS1mDesc  = prometheus.NewDesc("paper_tps_1m", "1 Minute TPS", nil, prometheus.Labels{"tps": "1m"})
	paperTPS5mDesc  = prometheus.NewDesc("paper_tps_5m", "5 Minute TPS", nil, prometheus.Labels{"tps": "5m"})
	paperTPS15mDesc = prometheus.NewDesc("paper_tps_15m", "15 Minute TPS", nil, prometheus.Labels{"tps": "15m"})

	dynmapTileRenderStatDesc       = prometheus.NewDesc("dynmap_tile_render_stat", "Tile Render Statistics reported by Dynmap", []string{"type", "file"}, nil)
	dynmapChunkLoadingCountDesc    = prometheus.NewDesc("dynmap_chunk_loading_count", "Chunk Loading Statistics reported by Dynmap", []string{"type"}, nil)
	dynmapChunkLoadingDurationDesc = prometheus.NewDesc("dynmap_chunk_loading_duration", "Chunk Loading Statistics reported by Dynmap", []string{"type"}, nil)
)

// Create new instance of collector, returns error if RCON is not correctly configured not provided
// Arguments:
//
//	cfg: Configuration for minecraft-exporter. Needs RCON to be filled out in full
func NewRCONCollector(cfg config.Config) (*RCONCollector, error) {
	rc, err := NewRCONClient(cfg.RCON.Host, cfg.RCON.Port, cfg.RCON.Password)
	if err != nil {
		return nil, err
	}
	return &RCONCollector{
		rcon:          rc,
		ServerType:    cfg.ServerType,
		DynmapEnabled: cfg.DynmapEnabled,
	}, nil
}

// Implements the Describe function for prometheus.Collector
func (c *RCONCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

// Implements the Collect function for prometheus.Collector
func (c *RCONCollector) Collect(ch chan<- prometheus.Metric) {
	slog.Debug("Starting collection of minecraft metrics via RCON")
	players := c.rcon.GetPlayersOnline()
	for _, player := range players {
		ch <- prometheus.MustNewConstMetric(mcPlayerOnlineDesc, prometheus.GaugeValue, 1, player)
	}
	switch c.ServerType {
	case config.SERVER_TYPE_FORGE:
		slog.Debug("Gathering forge metrics")
		dimStats, overallStat, err := c.rcon.GetForgeTPS()
		if err != nil {
			slog.Error("Failed to collect forge tps stats", "err", err)
		} else {
			for _, stat := range dimStats {
				ch <- prometheus.MustNewConstMetric(forgeTPSDimDesc, prometheus.CounterValue, stat.TPS, stat.ID, stat.Name)
				ch <- prometheus.MustNewConstMetric(forgeTicktimeDimDesc, prometheus.CounterValue, stat.Ticktime, stat.ID, stat.Name)
			}
			ch <- prometheus.MustNewConstMetric(forgeTPSOverallDesc, prometheus.CounterValue, overallStat.TPS)
			ch <- prometheus.MustNewConstMetric(forgeTicktimeOverallDesc, prometheus.CounterValue, overallStat.Ticktime)
		}
		entities, err := c.rcon.GetForgeEntities()
		if err != nil {
			slog.Error("Failed to retrieve forge entity list", "err", err)
		} else {
			for _, entity := range entities {
				ch <- prometheus.MustNewConstMetric(forgeEntitiesCountDesc, prometheus.CounterValue, float64(entity.Count), entity.Name)
			}
		}
	case config.SERVER_TYPE_PAPER:
		slog.Debug("Gathering paper metrics")
		paperTPS, err := c.rcon.GetPaperTPS()
		if err != nil {
			slog.Error("Failed to collect paper tps stats", "err", err)
		} else {
			if len(paperTPS) == 3 {
				ch <- prometheus.MustNewConstMetric(paperTPS1mDesc, prometheus.CounterValue, paperTPS[0])
				ch <- prometheus.MustNewConstMetric(paperTPS5mDesc, prometheus.CounterValue, paperTPS[1])
				ch <- prometheus.MustNewConstMetric(paperTPS15mDesc, prometheus.CounterValue, paperTPS[2])
			}
		}
	}

	if c.DynmapEnabled {
		slog.Debug("Gathering dynmap metrics")
		render, chunks, err := c.rcon.GetDynmapStats()
		if err != nil {
			slog.Error("Failed to collect dynmap stats", "err", err)
		} else {
			for _, stat := range render {
				ch <- prometheus.MustNewConstMetric(dynmapTileRenderStatDesc, prometheus.CounterValue, float64(stat.Processed), "processed", stat.Dim)
				ch <- prometheus.MustNewConstMetric(dynmapTileRenderStatDesc, prometheus.CounterValue, float64(stat.Rendered), "rendered", stat.Dim)
				ch <- prometheus.MustNewConstMetric(dynmapTileRenderStatDesc, prometheus.CounterValue, float64(stat.Updated), "updated", stat.Dim)
			}
			for _, stat := range chunks {
				ch <- prometheus.MustNewConstMetric(dynmapChunkLoadingCountDesc, prometheus.CounterValue, float64(stat.Count), stat.State)
				ch <- prometheus.MustNewConstMetric(dynmapChunkLoadingDurationDesc, prometheus.CounterValue, stat.Duration, stat.State)
			}
		}
	}
	slog.Debug("Finished collection of minecraft metrics via RCON")
}

// Close the RCON connection
func (c *RCONCollector) Close() error {
	return c.rcon.Close()
}
