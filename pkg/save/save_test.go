package save

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testUUID = "6f003e33-7076-4e45-a270-87841b218ec7"
)

func TestGetVersion(t *testing.T) {
	tMatrix := []struct {
		Name    string
		Version MinecraftVersion
	}{
		{
			Name: "1.12",
			Version: MinecraftVersion{
				Id:       1343,
				Name:     "1.12.2",
				Snapshot: false,
			},
		},
		{
			Name: "1.20",
			Version: MinecraftVersion{
				Id:       3465,
				Name:     "1.20.1",
				Snapshot: false,
			},
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)

			s, err := NewSave("./testdata/" + tCase.Name)
			if !assert.Nil(err) {
				t.FailNow()
			}

			err = s.GetVersion()

			if !assert.Nil(err) {
				t.FailNow()
			}
			assert.Equal(tCase.Version, s.Version)
		})
	}
}

func TestGetPlayers(t *testing.T) {
	tMatrix := []string{"1.12", "1.20"}

	for _, tCase := range tMatrix {
		t.Run(tCase, func(t *testing.T) {
			assert := assert.New(t)

			s, err := NewSave("./testdata/" + tCase)
			if !assert.Nil(err) {
				t.FailNow()
			}

			players, err := s.GetPlayers()

			if !assert.Nil(err) {
				t.FailNow()
			}
			assert.Equal([]string{testUUID}, players)
		})
	}
}

func TestLoadPlayerData(t *testing.T) {
	tMatrix := []struct {
		Name                                                     string
		LAdvancements                                            uint
		LCustom, LCrafted, LMined, LPickedUp, LKilled, LKilledBy int
		PlayerData                                               MinecraftPlayerData
		Stats                                                    CustomStats
	}{
		{
			Name:          "1.12",
			LAdvancements: 172,
			LCustom:       185,
			LCrafted:      378,
			LMined:        169,
			LPickedUp:     367,
			LKilled:       32,
			LKilledBy:     0,
			PlayerData: MinecraftPlayerData{
				XPTotal: 1604582, XPLevel: 554, Score: 1604867, Health: 156.79771423339844, FoodLevel: 20,
			},
			Stats: CustomStats{
				Jump:        4723,
				Deaths:      1,
				DamageTaken: -2147482291,
				DamageDealt: 289113,
				Playtime:    3767192,
				Walk:        9788652,
				Swim:        6497,
				Sprint:      1143985,
				Dive:        14008,
				Fall:        380871,
				Fly:         7036476,
				Boat:        0,
				Horse:       0,
				Climb:       6511,
				Sleep:       16,
				Crafted:     87,
			},
		},
		{
			Name:          "1.20",
			LAdvancements: 284,
			LCustom:       19,
			LCrafted:      37,
			LMined:        45,
			LPickedUp:     88,
			LKilled:       8,
			LKilledBy:     4,
			PlayerData: MinecraftPlayerData{
				XPTotal: 38, XPLevel: 3, Score: 850, Health: 20, FoodLevel: 20,
			},
			Stats: CustomStats{
				Jump:        3905,
				Deaths:      9,
				DamageTaken: 4867,
				DamageDealt: 3854,
				Playtime:    558274,
				Walk:        2413118,
				Swim:        56357,
				Sprint:      1132998,
				Dive:        45833,
				Fall:        63968,
				Fly:         1279370,
				Boat:        12555,
				Horse:       0,
				Climb:       87210,
				Sleep:       13,
				Crafted:     57,
			},
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)

			s, err := NewSave("./testdata/" + tCase.Name)
			if !assert.Nil(err) {
				t.FailNow()
			}

			d, err := s.LoadPlayerData(testUUID)
			if !assert.Nil(err) {
				t.Fatalf("Failed to load data: %v", err)
			}

			// Details are not necessary here
			assert.Equal(tCase.LAdvancements, countAdvancements(d.Advancements), "Advancements")
			assert.Equal(tCase.LCustom, len(d.Stats.Custom.Custom), "Custom")
			assert.Equal(tCase.LCrafted, len(d.Stats.CraftedItems), "CraftedItems")
			assert.Equal(tCase.LMined, len(d.Stats.Mined), "Mined")
			assert.Equal(tCase.LPickedUp, len(d.Stats.PickedUp), "PickedUp")
			assert.Equal(tCase.LKilled, len(d.Stats.Killed), "Killed")
			assert.Equal(tCase.LKilledBy, len(d.Stats.KilledBy), "KilledBy")

			d.Stats.Custom.Custom = nil

			assert.Equal(tCase.PlayerData, d.PlayerData)
			assert.Equal(tCase.Stats, d.Stats.Custom)

			// Print content when failing, to help adapt new test cases
			if t.Failed() {
				b, _ := json.Marshal(d.Stats.Custom)
				fmt.Println("Stats: " + string(b))
				b, _ = json.Marshal(d.PlayerData)
				fmt.Println("PlayerData: " + string(b))
			}
		})
	}

}
