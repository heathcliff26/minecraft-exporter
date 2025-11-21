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

	mcPlayerOnlineDesc *prometheus.Desc

	forgeTPSDimDesc          *prometheus.Desc
	forgeTicktimeDimDesc     *prometheus.Desc
	forgeTPSOverallDesc      *prometheus.Desc
	forgeTicktimeOverallDesc *prometheus.Desc
	forgeEntitiesCountDesc   *prometheus.Desc

	paperTPS1mDesc  *prometheus.Desc
	paperTPS5mDesc  *prometheus.Desc
	paperTPS15mDesc *prometheus.Desc

	dynmapTileRenderStatDesc       *prometheus.Desc
	dynmapChunkLoadingCountDesc    *prometheus.Desc
	dynmapChunkLoadingDurationDesc *prometheus.Desc

	tickTargetDesc  *prometheus.Desc
	tickAverageDesc *prometheus.Desc
	tickP50Desc     *prometheus.Desc
	tickP95Desc     *prometheus.Desc
	tickP99Desc     *prometheus.Desc
}

const (
	instanceLabelName = "instance"
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

		mcPlayerOnlineDesc: prometheus.NewDesc("minecraft_player_online", "Show currently online players. Value is always 1", []string{"player"}, prometheus.Labels{instanceLabelName: cfg.Instance}),

		forgeTPSDimDesc:          prometheus.NewDesc("forge_tps_dim", "TPS of a dimension", []string{"dimension_id", "dimension_name"}, prometheus.Labels{instanceLabelName: cfg.Instance}),
		forgeTicktimeDimDesc:     prometheus.NewDesc("forge_ticktime_dim", "Time a Tick took in a Dimension", []string{"dimension_id", "dimension_name"}, prometheus.Labels{instanceLabelName: cfg.Instance}),
		forgeTPSOverallDesc:      prometheus.NewDesc("forge_tps_overall", "Overall TPS", nil, prometheus.Labels{instanceLabelName: cfg.Instance}),
		forgeTicktimeOverallDesc: prometheus.NewDesc("forge_ticktime_overall", "Overall Ticktime", nil, prometheus.Labels{instanceLabelName: cfg.Instance}),
		forgeEntitiesCountDesc:   prometheus.NewDesc("forge_entity_count", "Type and count of active entities", []string{"entity"}, prometheus.Labels{instanceLabelName: cfg.Instance}),

		paperTPS1mDesc:  prometheus.NewDesc("paper_tps_1m", "1 Minute TPS", nil, prometheus.Labels{instanceLabelName: cfg.Instance, "tps": "1m"}),
		paperTPS5mDesc:  prometheus.NewDesc("paper_tps_5m", "5 Minute TPS", nil, prometheus.Labels{instanceLabelName: cfg.Instance, "tps": "5m"}),
		paperTPS15mDesc: prometheus.NewDesc("paper_tps_15m", "15 Minute TPS", nil, prometheus.Labels{instanceLabelName: cfg.Instance, "tps": "15m"}),

		dynmapTileRenderStatDesc:       prometheus.NewDesc("dynmap_tile_render_stat", "Tile Render Statistics reported by Dynmap", []string{"type", "file"}, prometheus.Labels{instanceLabelName: cfg.Instance}),
		dynmapChunkLoadingCountDesc:    prometheus.NewDesc("dynmap_chunk_loading_count", "Chunk Loading Statistics reported by Dynmap", []string{"type"}, prometheus.Labels{instanceLabelName: cfg.Instance}),
		dynmapChunkLoadingDurationDesc: prometheus.NewDesc("dynmap_chunk_loading_duration", "Chunk Loading Statistics reported by Dynmap", []string{"type"}, prometheus.Labels{instanceLabelName: cfg.Instance}),

		tickTargetDesc:  prometheus.NewDesc("minecraft_tick_target", "Targeted number of ticks per second", nil, prometheus.Labels{instanceLabelName: cfg.Instance}),
		tickAverageDesc: prometheus.NewDesc("minecraft_tick_average", "Average time per tick in milliseconds", nil, prometheus.Labels{instanceLabelName: cfg.Instance}),
		tickP50Desc:     prometheus.NewDesc("minecraft_tick_percentile", "Time per tick in percentiles", nil, prometheus.Labels{instanceLabelName: cfg.Instance, "percentile": "50"}),
		tickP95Desc:     prometheus.NewDesc("minecraft_tick_percentile", "Time per tick in percentiles", nil, prometheus.Labels{instanceLabelName: cfg.Instance, "percentile": "95"}),
		tickP99Desc:     prometheus.NewDesc("minecraft_tick_percentile", "Time per tick in percentiles", nil, prometheus.Labels{instanceLabelName: cfg.Instance, "percentile": "99"}),
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
		ch <- prometheus.MustNewConstMetric(c.mcPlayerOnlineDesc, prometheus.GaugeValue, 1, player)
	}
	switch c.ServerType {
	case config.SERVER_TYPE_FORGE, config.SERVER_TYPE_NEOFORGE:
		slog.Debug("Gathering forge metrics")
		dimStats, overallStat, err := c.rcon.GetForgeTPS(c.ServerType)
		if err != nil {
			slog.Error("Failed to collect forge tps stats", "err", err)
		} else {
			for _, stat := range dimStats {
				ch <- prometheus.MustNewConstMetric(c.forgeTPSDimDesc, prometheus.CounterValue, stat.TPS, stat.ID, stat.Name)
				ch <- prometheus.MustNewConstMetric(c.forgeTicktimeDimDesc, prometheus.CounterValue, stat.Ticktime, stat.ID, stat.Name)
			}
			ch <- prometheus.MustNewConstMetric(c.forgeTPSOverallDesc, prometheus.CounterValue, overallStat.TPS)
			ch <- prometheus.MustNewConstMetric(c.forgeTicktimeOverallDesc, prometheus.CounterValue, overallStat.Ticktime)
		}
		entities, err := c.rcon.GetForgeEntities(c.ServerType)
		if err != nil {
			slog.Error("Failed to retrieve forge entity list", "err", err)
		} else {
			for _, entity := range entities {
				ch <- prometheus.MustNewConstMetric(c.forgeEntitiesCountDesc, prometheus.CounterValue, float64(entity.Count), entity.Name)
			}
		}
	case config.SERVER_TYPE_PAPER:
		slog.Debug("Gathering paper metrics")
		paperTPS, err := c.rcon.GetPaperTPS()
		if err != nil {
			slog.Error("Failed to collect paper tps stats", "err", err)
		} else {
			if len(paperTPS) == 3 {
				ch <- prometheus.MustNewConstMetric(c.paperTPS1mDesc, prometheus.CounterValue, paperTPS[0])
				ch <- prometheus.MustNewConstMetric(c.paperTPS5mDesc, prometheus.CounterValue, paperTPS[1])
				ch <- prometheus.MustNewConstMetric(c.paperTPS15mDesc, prometheus.CounterValue, paperTPS[2])
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
				ch <- prometheus.MustNewConstMetric(c.dynmapTileRenderStatDesc, prometheus.CounterValue, float64(stat.Processed), "processed", stat.Dim)
				ch <- prometheus.MustNewConstMetric(c.dynmapTileRenderStatDesc, prometheus.CounterValue, float64(stat.Rendered), "rendered", stat.Dim)
				ch <- prometheus.MustNewConstMetric(c.dynmapTileRenderStatDesc, prometheus.CounterValue, float64(stat.Updated), "updated", stat.Dim)
			}
			for _, stat := range chunks {
				ch <- prometheus.MustNewConstMetric(c.dynmapChunkLoadingCountDesc, prometheus.CounterValue, float64(stat.Count), stat.State)
				ch <- prometheus.MustNewConstMetric(c.dynmapChunkLoadingDurationDesc, prometheus.CounterValue, stat.Duration, stat.State)
			}
		}
	}

	if c.rcon.V120() {
		tickStats, err := c.rcon.GetTickQuery()
		if err != nil {
			slog.Error("Failed to collect tick stats", "err", err)
		}

		ch <- prometheus.MustNewConstMetric(c.tickTargetDesc, prometheus.CounterValue, tickStats.Target)
		ch <- prometheus.MustNewConstMetric(c.tickAverageDesc, prometheus.CounterValue, tickStats.Average)
		ch <- prometheus.MustNewConstMetric(c.tickP50Desc, prometheus.CounterValue, tickStats.P50)
		ch <- prometheus.MustNewConstMetric(c.tickP95Desc, prometheus.CounterValue, tickStats.P95)
		ch <- prometheus.MustNewConstMetric(c.tickP99Desc, prometheus.CounterValue, tickStats.P99)
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
