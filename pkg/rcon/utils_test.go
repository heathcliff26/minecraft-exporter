package rcon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePlayersOnline(t *testing.T) {
	tMatrix := []struct {
		Name, Input string
		Players     []string
	}{
		{"1.12", "There are 2/10 players online:Foo1234, Bar5678", []string{"Foo1234", "Bar5678"}},
		{"1.12-empty", "There are 0/10 players online:", []string{}},
		{"1.20", "There are 0 of a max of 20 players online: Foo1234, Bar5678", []string{"Foo1234", "Bar5678"}},
		{"1.20-empty", "There are 0 of a max of 20 players online: ", []string{}},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			p := parsePlayersOnline(tCase.Input)

			assert.Equal(t, tCase.Players, p)
		})
	}
}

func TestParseForgeTPS(t *testing.T) {
	tMatrix := []struct {
		Name, Input string
		Stats       []TPSStat
		Total       TPSStat
	}{
		{
			Name:  "1.12",
			Input: "Dim  0 (DIM_0) : Mean tick time: 7.672 ms. Mean TPS: 20.000Dim 144 (CompactMachines) : Mean tick time: 0.294 ms. Mean TPS: 20.000Dim  2 (Storage Cell) : Mean tick time: 0.022 ms. Mean TPS: 20.000Overall : Mean tick time: 8.037 ms. Mean TPS: 20.000",
			Stats: []TPSStat{
				{
					ID:       "0",
					Name:     "DIM_0",
					Ticktime: 7.672,
					TPS:      20,
				},
				{
					ID:       "144",
					Name:     "CompactMachines",
					Ticktime: 0.294,
					TPS:      20,
				},
				{
					ID:       "2",
					Name:     "Storage Cell",
					Ticktime: 0.022,
					TPS:      20,
				},
			},
			Total: TPSStat{
				Ticktime: 8.037,
				TPS:      20,
			},
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			stats, total, err := parseForgeTPS(tCase.Input)

			assert := assert.New(t)

			assert.Nil(err)
			assert.Equal(tCase.Stats, stats)
			assert.Equal(tCase.Total, total)
		})
	}
}

func TestParseForgeEntities(t *testing.T) {
	tMatrix := []struct {
		Name, Input string
		Count       []EntityCount
	}{
		{Name: "1.12",
			Input: "Total: 24  12: minecraft:chicken  5: minecraft:cow  2: minecraft:item  2: minecraft:item_frame  2: minecraft:squid  1: farmingforblockheads:merchant",
			Count: []EntityCount{
				{
					Name:  "minecraft:chicken",
					Count: 12,
				},
				{
					Name:  "minecraft:cow",
					Count: 5,
				},
				{
					Name:  "minecraft:item",
					Count: 2,
				},
				{
					Name:  "minecraft:item_frame",
					Count: 2,
				},
				{
					Name:  "minecraft:squid",
					Count: 2,
				},
				{
					Name:  "farmingforblockheads:merchant",
					Count: 1,
				},
			},
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			c, err := parseForgeEntities(tCase.Input)

			assert := assert.New(t)

			assert.Nil(err)
			assert.Equal(tCase.Count, c)
		})
	}
}

func TestParsePaperTPS(t *testing.T) {
	tMatrix := []struct {
		Name, Input string
		TPS         []float64
	}{
		{Name: "1.20_raw",
			//lint:ignore ST1018 I need this string
			Input: "[0;33mTPS from last 1m, 5m, 15m: [0;1;32m20.0[0m, [0;1;32m20.0[0m, [0;1;32m20.0[0m[0m",
			TPS:   []float64{20, 20, 20},
		},
		{Name: "1.20",
			Input: "§6TPS from last 1m, 5m, 15m: §a20.0§r, §a20.0§r, §a20.0\n",
			TPS:   []float64{20, 20, 20},
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			tps, err := parsePaperTPS(tCase.Input)

			assert := assert.New(t)

			assert.Nil(err)
			assert.Equal(tCase.TPS, tps)
		})
	}
}

func TestParseDynmapStats(t *testing.T) {
	tMatrix := []struct {
		Name, Input string
		Render      []DynmapRenderStat
		Chunk       []DynmapChunkloadingStat
	}{
		{Name: "3.6",
			Input: "Tile Render Statistics:\n  world.cave: processed=50672, rendered=50672, updated=4757, transparent=0\n  world.flat: processed=30535, rendered=30535, updated=0, transparent=0\n  TOTALS: processed=81207, rendered=81207, updated=4757, transparent=0\n  Triggered update queue size: 0 + 0\n  Active render jobs:\nChunk Loading Statistics:\n  Cache hit rate: 91.25%\n  Chunks processed: Cached: count=3289892, 0.00 msec/chunk\n  Chunks processed: Already Loaded: count=315091, 21.15 msec/chunk\n  Chunks processed: Load Required: count=92, 5.91 msec/chunk\n  Chunks processed: Not Generated: count=1, 0.00 msec/chunk\n",
			Render: []DynmapRenderStat{
				{"world.cave", 50672, 50672, 4757},
				{"world.flat", 30535, 30535, 0},
				{"TOTALS", 81207, 81207, 4757},
			},
			Chunk: []DynmapChunkloadingStat{
				{"Cached", 3289892, 0},
				{"Already Loaded", 315091, 21.15},
				{"Load Required", 92, 5.91},
				{"Not Generated", 1, 0},
			},
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			r, c, err := parseDynmapStats(tCase.Input)

			assert := assert.New(t)

			assert.Nil(err)
			assert.Equal(tCase.Render, r)
			assert.Equal(tCase.Chunk, c)
		})
	}
}
