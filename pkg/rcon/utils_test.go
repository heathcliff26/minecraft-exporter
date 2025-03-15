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
			Name:  "Forge-1.12",
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
		{
			Name: "NeoForge-1.21",
			//lint:ignore ST1018 I need this string
			Input: `Overworld: 20.000 TPS (16.387 ms/tick)[0m
Nether Mining: 20.000 TPS (0.078 ms/tick)[0m
The End: 20.000 TPS (0.033 ms/tick)[0m
compactmachines:compact_world: 20.000 TPS (7.352 ms/tick)[0m
The Bumblezone: 20.000 TPS (0.054 ms/tick)[0m
The Void: 20.000 TPS (0.028 ms/tick)[0m
End Mining: 20.000 TPS (0.025 ms/tick)[0m
The Nether: 20.000 TPS (0.025 ms/tick)[0m
Mining: 20.000 TPS (0.024 ms/tick)[0m
the_afterdark:afterdark: 20.000 TPS (0.024 ms/tick)[0m
Eternal Starlight: 20.000 TPS (0.026 ms/tick)[0m
AE2 Spatial Storage: 20.000 TPS (0.025 ms/tick)[0m
Overall: 20.000 TPS (24.329 ms/tick)[0m
[0m`,
			Stats: []TPSStat{
				{
					Name:     "Overworld",
					Ticktime: 16.387,
					TPS:      20,
				},
				{
					Name:     "Nether Mining",
					Ticktime: 0.078,
					TPS:      20,
				},
				{
					Name:     "The End",
					Ticktime: 0.033,
					TPS:      20,
				},
				{
					Name:     "compactmachines:compact_world",
					Ticktime: 7.352,
					TPS:      20,
				},
				{
					Name:     "The Bumblezone",
					Ticktime: 0.054,
					TPS:      20,
				},
				{
					Name:     "The Void",
					Ticktime: 0.028,
					TPS:      20,
				},
				{
					Name:     "End Mining",
					Ticktime: 0.025,
					TPS:      20,
				},
				{
					Name:     "The Nether",
					Ticktime: 0.025,
					TPS:      20,
				},
				{
					Name:     "Mining",
					Ticktime: 0.024,
					TPS:      20,
				},
				{
					Name:     "the_afterdark:afterdark",
					Ticktime: 0.024,
					TPS:      20,
				},
				{
					Name:     "Eternal Starlight",
					Ticktime: 0.026,
					TPS:      20,
				},
				{
					Name:     "AE2 Spatial Storage",
					Ticktime: 0.025,
					TPS:      20,
				},
			},
			Total: TPSStat{
				Ticktime: 24.329,
				TPS:      20,
			},
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			stats, total, err := parseForgeTPS(tCase.Input)

			assert := assert.New(t)

			assert.NoError(err)
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
		{
			Name:  "Forge-1.12",
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
		{
			Name: "NeoForge-1.21",
			//lint:ignore ST1018 I need this string
			Input: `Total: 105[0m
  31: minecraft:chicken[0m
  15: minecraft:bat[0m
  12: minecraft:sheep[0m
  1: minecraft:horse[0m
[0m`,
			Count: []EntityCount{
				{
					Name:  "minecraft:chicken",
					Count: 31,
				},
				{
					Name:  "minecraft:bat",
					Count: 15,
				},
				{
					Name:  "minecraft:sheep",
					Count: 12,
				},
				{
					Name:  "minecraft:horse",
					Count: 1,
				},
			},
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			c, err := parseForgeEntities(tCase.Input)

			assert := assert.New(t)

			assert.NoError(err)
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

			assert.NoError(err)
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

			assert.NoError(err)
			assert.Equal(tCase.Render, r)
			assert.Equal(tCase.Chunk, c)
		})
	}
}

func TestParseTickQuery(t *testing.T) {
	tMatrix := []struct {
		Name, Input string
		Result      TickStats
	}{
		{

			Name:  "1.20.6",
			Input: "The game is running normallyTarget tick rate: 20,0 per second.\nAverage time per tick: 0,4ms (Target: 50,0ms)Percentiles: P50: 0,3ms P95: 0,6ms P99: 2,5ms, sample: 100",
			Result: TickStats{
				Target:  20.0,
				Average: 0.4,
				P50:     0.3,
				P95:     0.6,
				P99:     2.5,
			},
		},
		{
			Name:  "1.21.1",
			Input: "The game is running normallyTarget tick rate: 20,0 per second.\nAverage time per tick: 0,1ms (Target: 50,0ms)Percentiles: P50: 0,1ms P95: 0,1ms P99: 0,2ms, sample: 100",
			Result: TickStats{
				Target:  20.0,
				Average: 0.1,
				P50:     0.1,
				P95:     0.1,
				P99:     0.2,
			},
		},
		{

			Name:  "1.21.1-neoforge",
			Input: "The game is running normally\nTarget tick rate: 20.0 per second.\nAverage time per tick: 34.2ms (Target: 50.0ms)\nPercentiles: P50: 30.9ms P95: 49.6ms P99: 63.2ms, sample: 100\n",
			Result: TickStats{
				Target:  20.0,
				Average: 34.2,
				P50:     30.9,
				P95:     49.6,
				P99:     63.2,
			},
		},
		{
			Name:  "1.21.4",
			Input: "The game is running normallyTarget tick rate: 20.0 per second.\nAverage time per tick: 7.7ms (Target: 50.0ms)Percentiles: P50: 7.4ms P95: 9.9ms P99: 11.1ms, sample: 100",
			Result: TickStats{
				Target:  20.0,
				Average: 7.7,
				P50:     7.4,
				P95:     9.9,
				P99:     11.1,
			},
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)

			res, err := parseTickQuery(tCase.Input)

			assert.NoError(err)
			assert.Equal(tCase.Result, res)
		})
	}
}
