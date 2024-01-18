package save

import (
	"os"
	"path/filepath"
	"strings"

	version "github.com/hashicorp/go-version"
)

const (
	STATS_DIR        = "/stats"
	PLAYER_DIR       = "/playerdata"
	ADVANCEMENTS_DIR = "/advancements"
)

type Save struct {
	worldDir, statsDir, playerDir, advancementsDir string

	Version MinecraftVersion
}

// Create a new save from the given path
func NewSave(path string) (*Save, error) {
	if !isDirectory(path) {
		return nil, NewErrNoWorldDirectory("\"" + path + "\"" + " is not a directory")
	}

	s := &Save{
		worldDir:        path,
		statsDir:        path + STATS_DIR,
		playerDir:       path + PLAYER_DIR,
		advancementsDir: path + ADVANCEMENTS_DIR,
	}
	if !isDirectory(s.statsDir) {
		return nil, NewErrNoWorldDirectory("Failed to find \"stats\" subdirectory")
	}
	if !isDirectory(s.statsDir) {
		return nil, NewErrNoWorldDirectory("Failed to find \"playerdata\" subdirectory")
	}
	if !isDirectory(s.advancementsDir) {
		return nil, NewErrNoWorldDirectory("Failed to find \"advancements\" subdirectory")
	}

	return s, nil
}

// Return the Minecraft Version of the save
func (s *Save) GetVersion() error {
	var data MinecraftLevelDat
	err := readNBT(filepath.Join(s.worldDir, "level.dat"), &data)
	if err != nil {
		return err
	}

	s.Version = data.Data.Version
	return nil
}

// Return all players from the save
func (s *Save) GetPlayers() ([]string, error) {
	files, err := os.ReadDir(s.statsDir)
	if err != nil {
		return nil, err
	}
	players := make([]string, 0, len(files))
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if player, ok := strings.CutSuffix(f.Name(), ".json"); ok {
			players = append(players, player)
		}
	}
	return players, nil
}

// Load all relevant data for the given player
func (s *Save) LoadPlayerData(player string) (PlayerData, error) {
	advancements, err := s.loadAdvancements(player)
	if err != nil {
		return PlayerData{}, err
	}

	err = s.GetVersion()
	if err != nil {
		return PlayerData{}, err
	}

	stats, err := s.loadStats(player)
	if err != nil {
		return PlayerData{}, err
	}

	var data MinecraftPlayerData
	err = readNBT(filepath.Join(s.playerDir, player+".dat"), &data)
	if err != nil {
		return PlayerData{}, err
	}

	return PlayerData{
		Advancements: advancements,
		Stats:        stats,
		PlayerData:   data,
	}, nil
}

// Load the advancements for the given player
func (s *Save) loadAdvancements(player string) (map[string]Advancement, error) {
	var result MinecraftAdvancements
	err := readJSON(filepath.Join(s.advancementsDir, player+".json"), &result)
	if err != nil {
		return nil, err
	}

	return result.Advancements, nil
}

// Load the stats for the given player
func (s *Save) loadStats(player string) (Stats, error) {
	v115 := version.Must(version.NewSemver("1.15.0"))
	v, err := version.NewSemver(s.Version.Name)
	if err != nil {
		return Stats{}, err
	}

	if v115.LessThanOrEqual(v) {
		var stats MinecraftStats
		err := readJSON(filepath.Join(s.statsDir, player+".json"), &stats)
		if err != nil {
			return Stats{}, err
		}
		return stats.Stats, err
	} else {
		return s.loadStatsPre115(player)
	}
}

// Loads stats from saves prior to 1.15 and parses them into the new format
func (s *Save) loadStatsPre115(player string) (Stats, error) {
	var tmp map[string]int

	err := readJSON(filepath.Join(s.statsDir, player+".json"), &tmp)
	if err != nil {
		return Stats{}, err
	}

	stats := NewStats()

	for k, value := range tmp {
		key := strings.Split(k, ".")
		if key[0] != "stat" {
			continue
		}
		if len(key) < 2 {
			return Stats{}, NewErrFailedToParseStat(k, value)
		}
		switch key[1] {
		case "craftItem":
			name := strings.Join(key[2:], ":")
			stats.CraftedItems[name] = value
		case "mineBlock":
			name := strings.Join(key[2:], ":")
			stats.Mined[name] = value
		case "pickup":
			name := strings.Join(key[2:], ":")
			stats.PickedUp[name] = value
		case "killEntity":
			name := strings.Join(key[2:], ":")
			stats.Killed[name] = value
		case "entityKilledBy":
			name := strings.Join(key[2:], ":")
			stats.KilledBy[name] = value
		case "jump":
			stats.Custom.Jump = value
		case "deaths":
			stats.Custom.Deaths = value
		case "damageTaken":
			stats.Custom.DamageTaken = value
		case "damageDealt":
			stats.Custom.DamageDealt = value
		case "playOneMinute":
			stats.Custom.Playtime = value
		case "walkOneCm":
			stats.Custom.Walk = value
		case "swimOneCm":
			stats.Custom.Swim = value
		case "sprintOneCm":
			stats.Custom.Sprint = value
		case "diveOneCm":
			stats.Custom.Dive = value
		case "fallOneCm":
			stats.Custom.Fall = value
		case "flyOneCm":
			stats.Custom.Fly = value
		case "boatOneCm":
			stats.Custom.Boat = value
		case "horseOneCm":
			stats.Custom.Horse = value
		case "climbOneCm":
			stats.Custom.Climb = value
		case "sleepInBed":
			stats.Custom.Sleep = value
		case "craftingTableInteraction":
			stats.Custom.Crafted = value
		default:
			name := strings.Join(key[1:], ".")
			stats.Custom.Custom[name] = value
		}
	}
	return stats, nil
}
