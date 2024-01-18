package save

import (
	"encoding/json"
)

type MinecraftLevelDat struct {
	Data struct {
		Version        MinecraftVersion `nbt:"Version"`
		StorageVersion int              `nbt:"version"`
	} `nbt:"Data"`
}

type MinecraftVersion struct {
	Id       int    `nbt:"Id"`
	Name     string `nbt:"Name"`
	Snapshot bool   `nbt:"Snapshot"`
}

type PlayerData struct {
	Advancements map[string]Advancement
	Stats        Stats
	PlayerData   MinecraftPlayerData
}

type MinecraftAdvancements struct {
	Advancements map[string]Advancement
}

// Implements Unmarshal, allows to drop "DataVersion"
func (a *MinecraftAdvancements) UnmarshalJSON(data []byte) error {
	var v map[string]json.RawMessage
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	a.Advancements = make(map[string]Advancement, len(v))
	for key, value := range v {
		if key != "DataVersion" {
			var advancement Advancement
			if err := json.Unmarshal(value, &advancement); err != nil {
				return err
			}
			a.Advancements[key] = advancement
		}
	}
	return nil
}

type Advancement struct {
	Done bool `json:"done"`
}

type MinecraftStats struct {
	Stats Stats `json:"stats"`
}

type Stats struct {
	CraftedItems map[string]int `json:"minecraft:crafted"`
	Mined        map[string]int `json:"minecraft:mined"`
	PickedUp     map[string]int `json:"minecraft:picked_up"`
	Killed       map[string]int `json:"minecraft:killed"`
	KilledBy     map[string]int `json:"minecraft:killed_by"`
	Custom       CustomStats    `json:"minecraft:custom"`
}

func NewStats() Stats {
	return Stats{
		CraftedItems: make(map[string]int),
		Mined:        make(map[string]int),
		PickedUp:     make(map[string]int),
		Killed:       make(map[string]int),
		KilledBy:     make(map[string]int),
		Custom: CustomStats{
			Custom: make(map[string]int),
		},
	}
}

type CustomStats struct {
	Jump        int
	Deaths      int
	DamageTaken int
	DamageDealt int
	Playtime    int
	Walk        int
	Swim        int
	Sprint      int
	Dive        int
	Fall        int
	Fly         int
	Boat        int
	Horse       int
	Climb       int
	Sleep       int
	Crafted     int
	Custom      map[string]int
}

// Implements Unmarshal, allows unknown stats to be saved in map
func (s *CustomStats) UnmarshalJSON(data []byte) error {
	var v map[string]int
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	s.Custom = make(map[string]int, len(v))

	for key, value := range v {
		switch key {
		case "minecraft:jump":
			s.Jump = value
		case "minecraft:deaths":
			s.Deaths = value
		case "minecraft:damage_taken":
			s.DamageTaken = value
		case "minecraft:damage_dealt":
			s.DamageDealt = value
		case "minecraft:play_time", "minecraft:play_one_minute":
			s.Playtime = value
		case "minecraft:walk_one_cm":
			s.Walk = value
		case "minecraft:walk_on_water_one_cm":
			s.Swim = value
		case "minecraft:sprint_one_cm":
			s.Sprint = value
		case "minecraft:walk_under_water_one_cm":
			s.Dive = value
		case "minecraft:fall_one_cm":
			s.Fall = value
		case "minecraft:fly_one_cm":
			s.Fly = value
		case "minecraft:boat_one_cm":
			s.Boat = value
		case "minecraft:horse_one_cm":
			s.Horse = value
		case "minecraft:climb_one_cm":
			s.Climb = value
		case "minecraft:sleep_in_bed":
			s.Sleep = value
		case "minecraft:interact_with_crafting_table":
			s.Crafted = value
		default:
			s.Custom[key] = value
		}
	}
	return nil
}

type MinecraftPlayerData struct {
	XPTotal   int     `nbt:"XpTotal"`
	XPLevel   int     `nbt:"XpLevel"`
	Score     int     `nbt:"Score"`
	Health    float64 `nbt:"Health"`
	FoodLevel int     `nbt:"foodLevel"`
}
