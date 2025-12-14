package rcon

import (
	"log/slog"

	"github.com/heathcliff26/minecraft-exporter/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
)

type RCONCollector struct {
	rcon          *RCONClient
	ServerType    string
	DynmapEnabled bool

	Instance string
}

var (
	commonVariableLabels = []string{"instance"}

	mcPlayerOnlineDesc = prometheus.NewDesc("minecraft_player_online", "Show currently online players. Value is always 1", append(commonVariableLabels, "player"), nil)

	forgeTPSDimDesc          = prometheus.NewDesc("forge_tps_dim", "TPS of a dimension", append(commonVariableLabels, "dimension_id", "dimension_name"), nil)
	forgeTicktimeDimDesc     = prometheus.NewDesc("forge_ticktime_dim", "Time a Tick took in a Dimension", append(commonVariableLabels, "dimension_id", "dimension_name"), nil)
	forgeTPSOverallDesc      = prometheus.NewDesc("forge_tps_overall", "Overall TPS", commonVariableLabels, nil)
	forgeTicktimeOverallDesc = prometheus.NewDesc("forge_ticktime_overall", "Overall Ticktime", commonVariableLabels, nil)
	forgeEntitiesCountDesc   = prometheus.NewDesc("forge_entity_count", "Type and count of active entities", append(commonVariableLabels, "entity"), nil)

	paperTPS1mDesc  = prometheus.NewDesc("paper_tps_1m", "1 Minute TPS", commonVariableLabels, prometheus.Labels{"tps": "1m"})
	paperTPS5mDesc  = prometheus.NewDesc("paper_tps_5m", "5 Minute TPS", commonVariableLabels, prometheus.Labels{"tps": "5m"})
	paperTPS15mDesc = prometheus.NewDesc("paper_tps_15m", "15 Minute TPS", commonVariableLabels, prometheus.Labels{"tps": "15m"})

	dynmapTileRenderStatDesc       = prometheus.NewDesc("dynmap_tile_render_stat", "Tile Render Statistics reported by Dynmap", append(commonVariableLabels, "type", "file"), nil)
	dynmapChunkLoadingCountDesc    = prometheus.NewDesc("dynmap_chunk_loading_count", "Chunk Loading Statistics reported by Dynmap", append(commonVariableLabels, "type"), nil)
	dynmapChunkLoadingDurationDesc = prometheus.NewDesc("dynmap_chunk_loading_duration", "Chunk Loading Statistics reported by Dynmap", append(commonVariableLabels, "type"), nil)

	tickTargetDesc  = prometheus.NewDesc("minecraft_tick_target", "Targeted number of ticks per second", commonVariableLabels, nil)
	tickAverageDesc = prometheus.NewDesc("minecraft_tick_average", "Average time per tick in milliseconds", commonVariableLabels, nil)
	tickP50Desc     = prometheus.NewDesc("minecraft_tick_percentile", "Time per tick in percentiles", commonVariableLabels, prometheus.Labels{"percentile": "50"})
	tickP95Desc     = prometheus.NewDesc("minecraft_tick_percentile", "Time per tick in percentiles", commonVariableLabels, prometheus.Labels{"percentile": "95"})
	tickP99Desc     = prometheus.NewDesc("minecraft_tick_percentile", "Time per tick in percentiles", commonVariableLabels, prometheus.Labels{"percentile": "99"})
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

		Instance: cfg.Instance,
	}, nil
}

// Implements the Describe function for prometheus.Collector
func (c *RCONCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- mcPlayerOnlineDesc

	ch <- forgeTPSDimDesc
	ch <- forgeTicktimeDimDesc
	ch <- forgeTPSOverallDesc
	ch <- forgeTicktimeOverallDesc
	ch <- forgeEntitiesCountDesc

	ch <- paperTPS1mDesc
	ch <- paperTPS5mDesc
	ch <- paperTPS15mDesc

	ch <- dynmapTileRenderStatDesc
	ch <- dynmapChunkLoadingCountDesc
	ch <- dynmapChunkLoadingDurationDesc

	ch <- tickTargetDesc
	ch <- tickAverageDesc
	ch <- tickP50Desc
	ch <- tickP95Desc
	ch <- tickP99Desc
}

// Implements the Collect function for prometheus.Collector
func (c *RCONCollector) Collect(ch chan<- prometheus.Metric) {
	slog.Debug("Starting collection of minecraft metrics via RCON")
	commonLabels := []string{c.Instance}

	players := c.rcon.GetPlayersOnline()
	for _, player := range players {
		ch <- prometheus.MustNewConstMetric(mcPlayerOnlineDesc, prometheus.GaugeValue, 1, append(commonLabels, player)...)
	}
	switch c.ServerType {
	case config.SERVER_TYPE_FORGE, config.SERVER_TYPE_NEOFORGE:
		slog.Debug("Gathering forge metrics")
		dimStats, overallStat, err := c.rcon.GetForgeTPS(c.ServerType)
		if err != nil {
			slog.Error("Failed to collect forge tps stats", "err", err)
		} else {
			for _, stat := range dimStats {
				labels := append(commonLabels, stat.ID, stat.Name)
				ch <- prometheus.MustNewConstMetric(forgeTPSDimDesc, prometheus.CounterValue, stat.TPS, labels...)
				ch <- prometheus.MustNewConstMetric(forgeTicktimeDimDesc, prometheus.CounterValue, stat.Ticktime, labels...)
			}
			ch <- prometheus.MustNewConstMetric(forgeTPSOverallDesc, prometheus.CounterValue, overallStat.TPS, commonLabels...)
			ch <- prometheus.MustNewConstMetric(forgeTicktimeOverallDesc, prometheus.CounterValue, overallStat.Ticktime, commonLabels...)
		}
		entities, err := c.rcon.GetForgeEntities(c.ServerType)
		if err != nil {
			slog.Error("Failed to retrieve forge entity list", "err", err)
		} else {
			for _, entity := range entities {
				ch <- prometheus.MustNewConstMetric(forgeEntitiesCountDesc, prometheus.CounterValue, float64(entity.Count), append(commonLabels, entity.Name)...)
			}
		}
	case config.SERVER_TYPE_PAPER:
		slog.Debug("Gathering paper metrics")
		paperTPS, err := c.rcon.GetPaperTPS()
		if err != nil {
			slog.Error("Failed to collect paper tps stats", "err", err)
		} else {
			if len(paperTPS) == 3 {
				ch <- prometheus.MustNewConstMetric(paperTPS1mDesc, prometheus.CounterValue, paperTPS[0], commonLabels...)
				ch <- prometheus.MustNewConstMetric(paperTPS5mDesc, prometheus.CounterValue, paperTPS[1], commonLabels...)
				ch <- prometheus.MustNewConstMetric(paperTPS15mDesc, prometheus.CounterValue, paperTPS[2], commonLabels...)
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
				ch <- prometheus.MustNewConstMetric(dynmapTileRenderStatDesc, prometheus.CounterValue, float64(stat.Processed), append(commonLabels, "processed", stat.Dim)...)
				ch <- prometheus.MustNewConstMetric(dynmapTileRenderStatDesc, prometheus.CounterValue, float64(stat.Rendered), append(commonLabels, "rendered", stat.Dim)...)
				ch <- prometheus.MustNewConstMetric(dynmapTileRenderStatDesc, prometheus.CounterValue, float64(stat.Updated), append(commonLabels, "updated", stat.Dim)...)
			}
			for _, stat := range chunks {
				ch <- prometheus.MustNewConstMetric(dynmapChunkLoadingCountDesc, prometheus.CounterValue, float64(stat.Count), append(commonLabels, stat.State)...)
				ch <- prometheus.MustNewConstMetric(dynmapChunkLoadingDurationDesc, prometheus.CounterValue, stat.Duration, append(commonLabels, stat.State)...)
			}
		}
	}

	if c.rcon.V120() {
		tickStats, err := c.rcon.GetTickQuery()
		if err != nil {
			slog.Error("Failed to collect tick stats", "err", err)
		}

		ch <- prometheus.MustNewConstMetric(tickTargetDesc, prometheus.CounterValue, tickStats.Target, commonLabels...)
		ch <- prometheus.MustNewConstMetric(tickAverageDesc, prometheus.CounterValue, tickStats.Average, commonLabels...)
		ch <- prometheus.MustNewConstMetric(tickP50Desc, prometheus.CounterValue, tickStats.P50, commonLabels...)
		ch <- prometheus.MustNewConstMetric(tickP95Desc, prometheus.CounterValue, tickStats.P95, commonLabels...)
		ch <- prometheus.MustNewConstMetric(tickP99Desc, prometheus.CounterValue, tickStats.P99, commonLabels...)
	}
	slog.Debug("Finished collection of minecraft metrics via RCON")
}

// Expose the RCON client
func (c *RCONCollector) Client() *RCONClient {
	return c.rcon
}

// Close the RCON connection
func (c *RCONCollector) Close() error {
	return c.rcon.Close()
}
