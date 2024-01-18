package save

import (
	"log/slog"
	"time"

	"github.com/heathcliff26/containers/apps/minecraft-exporter/pkg/uuid"
	"github.com/prometheus/client_golang/prometheus"
)

type SaveCollector struct {
	save          *Save
	uuidCache     *uuid.UUIDCache
	ReduceMetrics bool
}

var (
	mcStatBlocksMinedReducedDesc    = prometheus.NewDesc("minecraft_stat_blocks_mined", "Blocks a player mined", []string{"player"}, nil)
	mcStatBlocksPickedUpReducedDesc = prometheus.NewDesc("minecraft_stat_blocks_picked_up", "Blocks a player picked up", []string{"player"}, nil)
	mcStatBlocksCraftedReducedDesc  = prometheus.NewDesc("minecraft_stat_blocks_crafted", "Items a player crafted", []string{"player"}, nil)

	mcStatBlocksMinedDesc       = prometheus.NewDesc("minecraft_stat_blocks_mined", "Blocks a player mined", []string{"player", "block"}, nil)
	mcStatBlocksPickedUpDesc    = prometheus.NewDesc("minecraft_stat_blocks_picked_up", "Blocks a player picked up", []string{"player", "block"}, nil)
	mcStatBlocksCraftedDesc     = prometheus.NewDesc("minecraft_stat_blocks_crafted", "Items a player crafted", []string{"player", "block"}, nil)
	mcStatDeathsDesc            = prometheus.NewDesc("minecraft_stat_deaths", "How often a player died. Cause \"minecraft:deaths\" is used for total deaths", []string{"player", "cause"}, nil)
	mcStatJumpsDesc             = prometheus.NewDesc("minecraft_stat_jumps", "How often a player has jumped", []string{"player"}, nil)
	mcStatCMTraveledDesc        = prometheus.NewDesc("minecraft_stat_cm_traveled", "How many cm a player traveled", []string{"player", "method"}, nil)
	mcStatXPTotalDesc           = prometheus.NewDesc("minecraft_stat_xp_total", "How much total XP a player earned", []string{"player"}, nil)
	mcStatCurrentLevelDesc      = prometheus.NewDesc("minecraft_stat_current_level", "How many levels the player currently has", []string{"player"}, nil)
	mcStatFoodLevelDesc         = prometheus.NewDesc("minecraft_stat_food_level", "How fed the player currently is", []string{"player"}, nil)
	mcStatHealthDesc            = prometheus.NewDesc("minecraft_stat_health", "How much health the player currently has", []string{"player"}, nil)
	mcStatScoreDesc             = prometheus.NewDesc("minecraft_stat_score", "The score of the player", []string{"player"}, nil)
	mcStatEntitiesKilledDesc    = prometheus.NewDesc("minecraft_stat_entities_killed", "Entities killed by player", []string{"player", "entity"}, nil)
	mcStatDamageTakenDesc       = prometheus.NewDesc("minecraft_stat_damage_taken", "Damage taken by player", []string{"player"}, nil)
	mcStatDamageDealtDesc       = prometheus.NewDesc("minecraft_stat_damage_dealt", "Damage dealt by player", []string{"player"}, nil)
	mcStatPlaytimeDesc          = prometheus.NewDesc("minecraft_stat_playtime", "Time in minutes a player was online", []string{"player"}, nil)
	mcStatAdvancementsDesc      = prometheus.NewDesc("minecraft_stat_advancements", "Number of completed advancements of a player", []string{"player"}, nil)
	mcStatSleptDesc             = prometheus.NewDesc("minecraft_stat_slept", "Times a player slept in a bed", []string{"player"}, nil)
	mcStatUsedCraftingTableDesc = prometheus.NewDesc("minecraft_stat_used_crafting_table", "Times a player used a crafting table", []string{"player"}, nil)
	mcStatCustomDesc            = prometheus.NewDesc("minecraft_stat_custom", "Custom minecraft stat", []string{"player", "stat"}, nil)
)

// Create new instance of collector, returns error if an world directory is not provided
// Arguments:
//
//		path: The path of the minecraft world directory
//	 reduceMetrics: Indicate if the amount of metrics should be reduced
func NewSaveCollector(path string, reduceMetrics bool) (*SaveCollector, error) {
	save, err := NewSave(path)
	if err != nil {
		return nil, err
	}

	return &SaveCollector{
		save:          save,
		uuidCache:     uuid.NewUUIDCache(time.Duration(time.Hour * 12)),
		ReduceMetrics: reduceMetrics,
	}, nil
}

// Implements the Describe function for prometheus.Collector
func (c *SaveCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

// Implements the Collect function for prometheus.Collector
func (c *SaveCollector) Collect(ch chan<- prometheus.Metric) {
	slog.Debug("Starting collection of minecraft metrics from savedata")

	players, err := c.save.GetPlayers()
	if err != nil {
		slog.Error("Failed to get list of players", "err", err)
		return
	}

	for _, player := range players {
		name, err := c.uuidCache.GetNameFromUUID(player)
		if err != nil {
			slog.Error("Failed to fetch name from uuid", "err", err, "player", player)
			return
		}

		d, err := c.save.LoadPlayerData(player)
		if err != nil {
			slog.Error("Failed to load data for player", "err", err, "player", player)
			return
		}

		if c.ReduceMetrics {
			ch <- prometheus.MustNewConstMetric(mcStatBlocksMinedReducedDesc, prometheus.CounterValue, float64(countTotal(d.Stats.Mined)), name)
			ch <- prometheus.MustNewConstMetric(mcStatBlocksPickedUpReducedDesc, prometheus.CounterValue, float64(countTotal(d.Stats.PickedUp)), name)
			ch <- prometheus.MustNewConstMetric(mcStatBlocksCraftedReducedDesc, prometheus.CounterValue, float64(countTotal(d.Stats.CraftedItems)), name)
		} else {
			mapToMetrics(ch, mcStatBlocksMinedDesc, d.Stats.Mined, name)
			mapToMetrics(ch, mcStatBlocksPickedUpDesc, d.Stats.PickedUp, name)
			mapToMetrics(ch, mcStatBlocksCraftedDesc, d.Stats.CraftedItems, name)
		}

		for key, value := range d.Stats.KilledBy {
			ch <- prometheus.MustNewConstMetric(mcStatDeathsDesc, prometheus.CounterValue, float64(value), name, key)
		}
		ch <- prometheus.MustNewConstMetric(mcStatDeathsDesc, prometheus.CounterValue, float64(d.Stats.Custom.Deaths), name, "minecraft:deaths")

		ch <- prometheus.MustNewConstMetric(mcStatJumpsDesc, prometheus.CounterValue, float64(d.Stats.Custom.Jump), name)
		ch <- prometheus.MustNewConstMetric(mcStatCMTraveledDesc, prometheus.CounterValue, float64(d.Stats.Custom.Walk), name, "walking")
		ch <- prometheus.MustNewConstMetric(mcStatCMTraveledDesc, prometheus.CounterValue, float64(d.Stats.Custom.Swim), name, "swimming")
		ch <- prometheus.MustNewConstMetric(mcStatCMTraveledDesc, prometheus.CounterValue, float64(d.Stats.Custom.Sprint), name, "sprinting")
		ch <- prometheus.MustNewConstMetric(mcStatCMTraveledDesc, prometheus.CounterValue, float64(d.Stats.Custom.Dive), name, "diving")
		ch <- prometheus.MustNewConstMetric(mcStatCMTraveledDesc, prometheus.CounterValue, float64(d.Stats.Custom.Fall), name, "falling")
		ch <- prometheus.MustNewConstMetric(mcStatCMTraveledDesc, prometheus.CounterValue, float64(d.Stats.Custom.Fly), name, "flying")
		ch <- prometheus.MustNewConstMetric(mcStatCMTraveledDesc, prometheus.CounterValue, float64(d.Stats.Custom.Boat), name, "boat")
		ch <- prometheus.MustNewConstMetric(mcStatCMTraveledDesc, prometheus.CounterValue, float64(d.Stats.Custom.Horse), name, "Horse")
		ch <- prometheus.MustNewConstMetric(mcStatCMTraveledDesc, prometheus.CounterValue, float64(d.Stats.Custom.Climb), name, "climbing")

		ch <- prometheus.MustNewConstMetric(mcStatXPTotalDesc, prometheus.CounterValue, float64(d.PlayerData.XPTotal), name)
		ch <- prometheus.MustNewConstMetric(mcStatCurrentLevelDesc, prometheus.CounterValue, float64(d.PlayerData.XPLevel), name)
		ch <- prometheus.MustNewConstMetric(mcStatFoodLevelDesc, prometheus.CounterValue, float64(d.PlayerData.FoodLevel), name)
		ch <- prometheus.MustNewConstMetric(mcStatHealthDesc, prometheus.CounterValue, float64(d.PlayerData.Health), name)
		ch <- prometheus.MustNewConstMetric(mcStatScoreDesc, prometheus.CounterValue, float64(d.PlayerData.Score), name)

		for key, value := range d.Stats.Killed {
			ch <- prometheus.MustNewConstMetric(mcStatEntitiesKilledDesc, prometheus.CounterValue, float64(value), name, key)
		}

		ch <- prometheus.MustNewConstMetric(mcStatDamageTakenDesc, prometheus.CounterValue, float64(d.Stats.Custom.DamageTaken), name)
		ch <- prometheus.MustNewConstMetric(mcStatDamageDealtDesc, prometheus.CounterValue, float64(d.Stats.Custom.DamageDealt), name)
		ch <- prometheus.MustNewConstMetric(mcStatPlaytimeDesc, prometheus.CounterValue, float64(d.Stats.Custom.Playtime), name)
		ch <- prometheus.MustNewConstMetric(mcStatSleptDesc, prometheus.CounterValue, float64(d.Stats.Custom.Sleep), name)
		ch <- prometheus.MustNewConstMetric(mcStatUsedCraftingTableDesc, prometheus.CounterValue, float64(d.Stats.Custom.Crafted), name)

		advancements := countAdvancements(d.Advancements)
		ch <- prometheus.MustNewConstMetric(mcStatAdvancementsDesc, prometheus.CounterValue, float64(advancements), name)

		for key, value := range d.Stats.Custom.Custom {
			ch <- prometheus.MustNewConstMetric(mcStatCustomDesc, prometheus.CounterValue, float64(value), name, key)
		}
	}

	slog.Debug("Finished collection of minecraft metrics from savedata")
}
