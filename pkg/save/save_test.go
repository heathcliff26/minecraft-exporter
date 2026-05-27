package save

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testUUID = "6f003e33-7076-4e45-a270-87841b218ec7"
)

func TestNewSave(t *testing.T) {
	t.Run("Legacy", func(t *testing.T) {
		require := require.New(t)

		path := "testdata/1.20"

		s, err := NewSave(path)
		require.NoError(err, "Should succeed")
		require.NotNil(s, "Should return save")

		expectedSave := &Save{
			worldDir:        path,
			statsDir:        path + STATS_DIR_LEGACY,
			playerDir:       path + PLAYER_DIR_LEGACY,
			advancementsDir: path + ADVANCEMENTS_DIR_LEGACY,
			Version: MinecraftVersion{
				Id:       3465,
				Name:     "1.20.1",
				Snapshot: false,
			},
		}
		require.Equal(expectedSave, s, "Should return expected save")
	})
	t.Run("v26", func(t *testing.T) {
		require := require.New(t)

		path := "testdata/26"

		s, err := NewSave(path)
		require.NoError(err, "Should succeed")
		require.NotNil(s, "Should return save")

		expectedSave := &Save{
			worldDir:        path,
			statsDir:        path + STATS_DIR,
			playerDir:       path + PLAYER_DIR,
			advancementsDir: path + ADVANCEMENTS_DIR,
			Version: MinecraftVersion{
				Id:       4790,
				Name:     "26.1.2",
				Snapshot: false,
			},
		}
		require.Equal(expectedSave, s, "Should return expected save")
	})
	t.Run("InvalidPath", func(t *testing.T) {
		assert := assert.New(t)

		s, err := NewSave("invalid-path")
		assert.ErrorContains(err, "is not a directory", "Should fail")
		assert.Nil(s, "Should return nil save")
	})
	t.Run("FailedToReadVersion", func(t *testing.T) {
		assert := assert.New(t)

		tmpDir := t.TempDir()

		s, err := NewSave(tmpDir)
		assert.ErrorContains(err, "Failed to read minecraft version:", "Should fail")
		assert.Nil(s, "Should return nil save")
	})
	t.Run("MissingStatsDir", func(t *testing.T) {
		require := require.New(t)

		tmpDir := t.TempDir()
		err := copyFile("testdata/1.20/level.dat", tmpDir+"/level.dat")
		require.NoError(err, "Should copy level.dat")

		s, err := NewSave(tmpDir)

		require.ErrorContains(err, "Failed to find player stats subdirectory", "Should fail")
		require.Nil(s, "Should return nil save")
	})
	t.Run("MissingPlayerDataDir", func(t *testing.T) {
		require := require.New(t)

		tmpDir := t.TempDir()
		err := copyFile("testdata/1.20/level.dat", tmpDir+"/level.dat")
		require.NoError(err, "Should copy level.dat")
		err = os.Mkdir(tmpDir+STATS_DIR_LEGACY, 0755)
		require.NoError(err, "Should create stats dir")

		s, err := NewSave(tmpDir)

		require.ErrorContains(err, "Failed to find player data subdirectory", "Should fail")
		require.Nil(s, "Should return nil save")
	})
	t.Run("MissingAdvancemetsDir", func(t *testing.T) {
		require := require.New(t)

		tmpDir := t.TempDir()
		err := copyFile("testdata/1.20/level.dat", tmpDir+"/level.dat")
		require.NoError(err, "Should copy level.dat")
		err = os.Mkdir(tmpDir+STATS_DIR_LEGACY, 0755)
		require.NoError(err, "Should create stats dir")
		err = os.Mkdir(tmpDir+PLAYER_DIR_LEGACY, 0755)
		require.NoError(err, "Should create player data dir")

		s, err := NewSave(tmpDir)

		require.ErrorContains(err, "Failed to find player advancements subdirectory", "Should fail")
		require.Nil(s, "Should return nil save")
	})
}

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
		{
			Name: "26",
			Version: MinecraftVersion{
				Id:       4790,
				Name:     "26.1.2",
				Snapshot: false,
			},
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			require := require.New(t)

			s, err := NewSave("./testdata/" + tCase.Name)
			require.NoError(err)
			require.Equal(tCase.Version, s.Version)
		})
	}
}

func TestGetPlayers(t *testing.T) {
	tMatrix := []string{"1.12", "1.20", "26"}

	for _, tCase := range tMatrix {
		t.Run(tCase, func(t *testing.T) {
			assert := assert.New(t)

			s, err := NewSave("./testdata/" + tCase)
			if !assert.NoError(err) {
				t.FailNow()
			}

			players, err := s.GetPlayers()

			if !assert.NoError(err) {
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
		{
			Name:          "26",
			LAdvancements: 304,
			LCustom:       13,
			LCrafted:      43,
			LMined:        45,
			LPickedUp:     76,
			LKilled:       7,
			LKilledBy:     2,
			PlayerData: MinecraftPlayerData{
				XPTotal: 3, XPLevel: 0, Score: 383, Health: 20, FoodLevel: 20,
			},
			Stats: CustomStats{
				Jump:        2568,
				Deaths:      5,
				DamageTaken: 1752,
				DamageDealt: 5801,
				Playtime:    169215,
				Walk:        334136,
				Swim:        20013,
				Sprint:      585346,
				Dive:        3081,
				Fall:        40889,
				Fly:         415053,
				Boat:        6762,
				Horse:       601,
				Climb:       683,
				Sleep:       4,
				Crafted:     74,
			},
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)

			s, err := NewSave("./testdata/" + tCase.Name)
			if !assert.NoError(err) {
				t.FailNow()
			}

			d, err := s.LoadPlayerData(testUUID)
			if !assert.NoError(err) {
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

func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}
