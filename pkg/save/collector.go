package save

import (
	"log/slog"
	"time"

	"github.com/heathcliff26/minecraft-exporter/pkg/rcon"
	"github.com/heathcliff26/minecraft-exporter/pkg/uuid"
	"github.com/prometheus/client_golang/prometheus"
)

type SaveCollector struct {
	save          *Save
	uuidCache     *uuid.UUIDCache
	ReduceMetrics bool
	Instance      string

	RCON *rcon.RCONClient
}

var (
	commonVariableLabels = []string{"instance", "player"}

	mcStatBlocksMinedReducedDesc    = prometheus.NewDesc("minecraft_stat_blocks_mined", "Blocks a player mined", commonVariableLabels, nil)
	mcStatBlocksPickedUpReducedDesc = prometheus.NewDesc("minecraft_stat_blocks_picked_up", "Blocks a player picked up", commonVariableLabels, nil)
	mcStatBlocksCraftedReducedDesc  = prometheus.NewDesc("minecraft_stat_blocks_crafted", "Items a player crafted", commonVariableLabels, nil)

	mcStatBlocksMinedDesc    = prometheus.NewDesc("minecraft_stat_blocks_mined", "Blocks a player mined", append(commonVariableLabels, "block"), nil)
	mcStatBlocksPickedUpDesc = prometheus.NewDesc("minecraft_stat_blocks_picked_up", "Blocks a player picked up", append(commonVariableLabels, "block"), nil)
	mcStatBlocksCraftedDesc  = prometheus.NewDesc("minecraft_stat_blocks_crafted", "Items a player crafted", append(commonVariableLabels, "block"), nil)

	mcStatDeathsDesc            = prometheus.NewDesc("minecraft_stat_deaths", "How often a player died. Cause \"minecraft:deaths\" is used for total deaths", append(commonVariableLabels, "cause"), nil)
	mcStatJumpsDesc             = prometheus.NewDesc("minecraft_stat_jumps", "How often a player has jumped", commonVariableLabels, nil)
	mcStatCMTraveledDesc        = prometheus.NewDesc("minecraft_stat_cm_traveled", "How many cm a player traveled", append(commonVariableLabels, "method"), nil)
	mcStatXPTotalDesc           = prometheus.NewDesc("minecraft_stat_xp_total", "How much total XP a player earned", commonVariableLabels, nil)
	mcStatCurrentLevelDesc      = prometheus.NewDesc("minecraft_stat_current_level", "How many levels the player currently has", commonVariableLabels, nil)
	mcStatFoodLevelDesc         = prometheus.NewDesc("minecraft_stat_food_level", "How fed the player currently is", commonVariableLabels, nil)
	mcStatHealthDesc            = prometheus.NewDesc("minecraft_stat_health", "How much health the player currently has", commonVariableLabels, nil)
	mcStatScoreDesc             = prometheus.NewDesc("minecraft_stat_score", "The score of the player", commonVariableLabels, nil)
	mcStatEntitiesKilledDesc    = prometheus.NewDesc("minecraft_stat_entities_killed", "Entities killed by player", append(commonVariableLabels, "entity"), nil)
	mcStatDamageTakenDesc       = prometheus.NewDesc("minecraft_stat_damage_taken", "Damage taken by player", commonVariableLabels, nil)
	mcStatDamageDealtDesc       = prometheus.NewDesc("minecraft_stat_damage_dealt", "Damage dealt by player", commonVariableLabels, nil)
	mcStatPlaytimeDesc          = prometheus.NewDesc("minecraft_stat_playtime", "Time in minutes a player was online", commonVariableLabels, nil)
	mcStatAdvancementsDesc      = prometheus.NewDesc("minecraft_stat_advancements", "Number of completed advancements of a player", commonVariableLabels, nil)
	mcStatSleptDesc             = prometheus.NewDesc("minecraft_stat_slept", "Times a player slept in a bed", commonVariableLabels, nil)
	mcStatUsedCraftingTableDesc = prometheus.NewDesc("minecraft_stat_used_crafting_table", "Times a player used a crafting table", commonVariableLabels, nil)
	mcStatCustomDesc            = prometheus.NewDesc("minecraft_stat_custom", "Custom minecraft stat", append(commonVariableLabels, "stat"), nil)
)

// Create new instance of collector, returns error if an world directory is not provided
// Arguments:
//
//	path: The path of the minecraft world directory
//	instance: The instance label to use for the metrics
//	reduceMetrics: Indicate if the amount of metrics should be reduced
func NewSaveCollector(path, instance string, reduceMetrics bool) (*SaveCollector, error) {
	save, err := NewSave(path)
	if err != nil {
		return nil, err
	}

	return &SaveCollector{
		save:          save,
		uuidCache:     uuid.NewUUIDCache(time.Duration(time.Hour * 12)),
		ReduceMetrics: reduceMetrics,
		Instance:      instance,
	}, nil
}

// Implements the Describe function for prometheus.Collector
func (c *SaveCollector) Describe(ch chan<- *prometheus.Desc) {
	if c.ReduceMetrics {
		ch <- mcStatBlocksMinedReducedDesc
		ch <- mcStatBlocksPickedUpReducedDesc
		ch <- mcStatBlocksCraftedReducedDesc
	} else {
		ch <- mcStatBlocksMinedDesc
		ch <- mcStatBlocksPickedUpDesc
		ch <- mcStatBlocksCraftedDesc
	}

	ch <- mcStatDeathsDesc
	ch <- mcStatJumpsDesc
	ch <- mcStatCMTraveledDesc
	ch <- mcStatXPTotalDesc
	ch <- mcStatCurrentLevelDesc
	ch <- mcStatFoodLevelDesc
	ch <- mcStatHealthDesc
	ch <- mcStatScoreDesc
	ch <- mcStatEntitiesKilledDesc
	ch <- mcStatDamageTakenDesc
	ch <- mcStatDamageDealtDesc
	ch <- mcStatPlaytimeDesc
	ch <- mcStatAdvancementsDesc
	ch <- mcStatSleptDesc
	ch <- mcStatUsedCraftingTableDesc
	ch <- mcStatCustomDesc
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

		commonLabels := []string{c.Instance, name}

		if c.ReduceMetrics {
			ch <- prometheus.MustNewConstMetric(mcStatBlocksMinedReducedDesc, prometheus.CounterValue, float64(countTotal(d.Stats.Mined)), commonLabels...)
			ch <- prometheus.MustNewConstMetric(mcStatBlocksPickedUpReducedDesc, prometheus.CounterValue, float64(countTotal(d.Stats.PickedUp)), commonLabels...)
			ch <- prometheus.MustNewConstMetric(mcStatBlocksCraftedReducedDesc, prometheus.CounterValue, float64(countTotal(d.Stats.CraftedItems)), commonLabels...)
		} else {
			mapToMetrics(ch, mcStatBlocksMinedDesc, d.Stats.Mined, commonLabels)
			mapToMetrics(ch, mcStatBlocksPickedUpDesc, d.Stats.PickedUp, commonLabels)
			mapToMetrics(ch, mcStatBlocksCraftedDesc, d.Stats.CraftedItems, commonLabels)
		}

		for key, value := range d.Stats.KilledBy {
			ch <- prometheus.MustNewConstMetric(mcStatDeathsDesc, prometheus.CounterValue, float64(value), append(commonLabels, key)...)
		}
		ch <- prometheus.MustNewConstMetric(mcStatDeathsDesc, prometheus.CounterValue, float64(d.Stats.Custom.Deaths), append(commonLabels, "minecraft:deaths")...)

		ch <- prometheus.MustNewConstMetric(mcStatJumpsDesc, prometheus.CounterValue, float64(d.Stats.Custom.Jump), commonLabels...)
		ch <- prometheus.MustNewConstMetric(mcStatCMTraveledDesc, prometheus.CounterValue, float64(d.Stats.Custom.Walk), append(commonLabels, "walking")...)
		ch <- prometheus.MustNewConstMetric(mcStatCMTraveledDesc, prometheus.CounterValue, float64(d.Stats.Custom.Swim), append(commonLabels, "swimming")...)
		ch <- prometheus.MustNewConstMetric(mcStatCMTraveledDesc, prometheus.CounterValue, float64(d.Stats.Custom.Sprint), append(commonLabels, "sprinting")...)
		ch <- prometheus.MustNewConstMetric(mcStatCMTraveledDesc, prometheus.CounterValue, float64(d.Stats.Custom.Dive), append(commonLabels, "diving")...)
		ch <- prometheus.MustNewConstMetric(mcStatCMTraveledDesc, prometheus.CounterValue, float64(d.Stats.Custom.Fall), append(commonLabels, "falling")...)
		ch <- prometheus.MustNewConstMetric(mcStatCMTraveledDesc, prometheus.CounterValue, float64(d.Stats.Custom.Fly), append(commonLabels, "flying")...)
		ch <- prometheus.MustNewConstMetric(mcStatCMTraveledDesc, prometheus.CounterValue, float64(d.Stats.Custom.Boat), append(commonLabels, "boat")...)
		ch <- prometheus.MustNewConstMetric(mcStatCMTraveledDesc, prometheus.CounterValue, float64(d.Stats.Custom.Horse), append(commonLabels, "Horse")...)
		ch <- prometheus.MustNewConstMetric(mcStatCMTraveledDesc, prometheus.CounterValue, float64(d.Stats.Custom.Climb), append(commonLabels, "climbing")...)

		ch <- prometheus.MustNewConstMetric(mcStatXPTotalDesc, prometheus.CounterValue, float64(d.PlayerData.XPTotal), commonLabels...)
		ch <- prometheus.MustNewConstMetric(mcStatCurrentLevelDesc, prometheus.CounterValue, float64(d.PlayerData.XPLevel), commonLabels...)
		ch <- prometheus.MustNewConstMetric(mcStatFoodLevelDesc, prometheus.CounterValue, float64(d.PlayerData.FoodLevel), commonLabels...)
		ch <- prometheus.MustNewConstMetric(mcStatHealthDesc, prometheus.CounterValue, float64(d.PlayerData.Health), commonLabels...)
		ch <- prometheus.MustNewConstMetric(mcStatScoreDesc, prometheus.CounterValue, float64(d.PlayerData.Score), commonLabels...)

		for key, value := range d.Stats.Killed {
			ch <- prometheus.MustNewConstMetric(mcStatEntitiesKilledDesc, prometheus.CounterValue, float64(value), append(commonLabels, key)...)
		}

		ch <- prometheus.MustNewConstMetric(mcStatDamageTakenDesc, prometheus.CounterValue, float64(d.Stats.Custom.DamageTaken), commonLabels...)
		ch <- prometheus.MustNewConstMetric(mcStatDamageDealtDesc, prometheus.CounterValue, float64(d.Stats.Custom.DamageDealt), commonLabels...)
		ch <- prometheus.MustNewConstMetric(mcStatPlaytimeDesc, prometheus.CounterValue, float64(d.Stats.Custom.Playtime), commonLabels...)
		ch <- prometheus.MustNewConstMetric(mcStatSleptDesc, prometheus.CounterValue, float64(d.Stats.Custom.Sleep), commonLabels...)
		ch <- prometheus.MustNewConstMetric(mcStatUsedCraftingTableDesc, prometheus.CounterValue, float64(d.Stats.Custom.Crafted), commonLabels...)

		advancements := countAdvancements(d.Advancements)
		ch <- prometheus.MustNewConstMetric(mcStatAdvancementsDesc, prometheus.CounterValue, float64(advancements), commonLabels...)

		for key, value := range d.Stats.Custom.Custom {
			ch <- prometheus.MustNewConstMetric(mcStatCustomDesc, prometheus.CounterValue, float64(value), append(commonLabels, key)...)
		}
	}

	// Update the game version in the rcon client
	if c.RCON != nil && c.RCON.Version() != c.save.Version.Name {
		slog.Info("Minecraft Version", "version", c.save.Version.Name)
		c.RCON.UpdateVersion(c.save.Version.Name)
	}

	slog.Debug("Finished collection of minecraft metrics from savedata")
}
