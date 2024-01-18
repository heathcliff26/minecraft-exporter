package rcon

import (
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/text"
)

// Parse the output of the list command
func parsePlayersOnline(input string) []string {
	s := strings.Split(input, "players online:")
	if len(s) < 2 {
		return []string{}
	}

	players := strings.TrimSpace(s[1])
	if players == "" {
		return []string{}
	}

	return strings.Split(players, ", ")
}

// Parse the TPS statistics returned from forge
func parseForgeTPS(input string) ([]TPSStat, TPSStat, error) {
	reg := regexp.MustCompile(`Dim\s*(-*\d*)\s\((.*?)\)\s:\sMean tick time:\s(.*?) ms\. Mean TPS: (\d*\.\d*)`)
	stats := reg.FindAllStringSubmatch(input, -1)
	dimStats := make([]TPSStat, len(stats))
	for i, stat := range stats {
		ticktime, err := strconv.ParseFloat(stat[3], 64)
		if err != nil {
			slog.Error("Failed to parse ticktime", "err", err, "value", stat[3])
			return nil, TPSStat{}, err
		}
		tps, err := strconv.ParseFloat(stat[4], 64)
		if err != nil {
			slog.Error("Failed to parse tps", "err", err, "value", stat[4])
			return nil, TPSStat{}, err
		}
		dimStats[i] = TPSStat{
			ID:       stat[1],
			Name:     stat[2],
			Ticktime: ticktime,
			TPS:      tps,
		}
	}

	reg = regexp.MustCompile(`Overall\s?: Mean tick time: (.*) ms. Mean TPS: (.*)`)
	overallStat := reg.FindStringSubmatch(input)
	if overallStat == nil {
		return nil, TPSStat{}, ErrForgeTPS{}
	}
	ticktime, err := strconv.ParseFloat(overallStat[1], 64)
	if err != nil {
		slog.Error("Failed to parse ticktime", "err", err, "value", overallStat[1])
		return nil, TPSStat{}, err
	}
	tps, err := strconv.ParseFloat(overallStat[2], 64)
	if err != nil {
		slog.Error("Failed to parse tps", "err", err, "value", overallStat[2])
		return nil, TPSStat{}, err
	}

	return dimStats, TPSStat{Ticktime: ticktime, TPS: tps}, nil
}

// Parse the count and name of all loaded forge entities
func parseForgeEntities(input string) ([]EntityCount, error) {
	reg := regexp.MustCompile(`(\d+): (.*?:.*?)\s`)
	matches := reg.FindAllStringSubmatch(input+" ", -1)
	res := make([]EntityCount, len(matches))
	for i, s := range matches {
		count, err := strconv.Atoi(s[1])
		if err != nil {
			return nil, err
		}
		res[i] = EntityCount{
			Name:  s[2],
			Count: count,
		}
	}

	return res, nil
}

// Parse the TPS statistics returned from paper
func parsePaperTPS(input string) ([]float64, error) {
	// Starting with 1.20, the output is colored
	input = text.StripEscape(input)
	for _, chars := range []string{"§6", "§a", "§r", "\n"} {
		input = strings.ReplaceAll(input, chars, "")
	}

	input = strings.TrimPrefix(input, "TPS from last 1m, 5m, 15m: ")
	s := strings.Split(input, ", ")
	if len(s) != 3 {
		return []float64{}, NewErrPaperTPS(input, len(s))
	}

	tps := make([]float64, 3)
	for i := 0; i < len(s); i++ {
		var err error
		tps[i], err = strconv.ParseFloat(s[i], 64)
		if err != nil {
			return []float64{}, err
		}
	}

	return tps, nil
}

// Parse the render statistics returned from Dynmap
func parseDynmapStats(input string) ([]DynmapRenderStat, []DynmapChunkloadingStat, error) {
	reg := regexp.MustCompile(`  (.*?): processed=(\d*), rendered=(\d*), updated=(\d*)`)
	matches := reg.FindAllStringSubmatch(input, -1)
	renderStats := make([]DynmapRenderStat, len(matches))
	for i, stat := range matches {
		processed, err := strconv.Atoi(stat[2])
		if err != nil {
			slog.Error("Failed to parse dynmap render stats", "err", err, "out", input)
			return nil, nil, err
		}
		rendered, err := strconv.Atoi(stat[3])
		if err != nil {
			slog.Error("Failed to parse dynmap render stats", "err", err, "out", input)
			return nil, nil, err
		}
		updated, err := strconv.Atoi(stat[4])
		if err != nil {
			slog.Error("Failed to parse dynmap render stats", "err", err, "out", input)
			return nil, nil, err
		}
		renderStats[i] = DynmapRenderStat{
			Dim:       stat[1],
			Processed: processed,
			Rendered:  rendered,
			Updated:   updated,
		}
	}

	reg = regexp.MustCompile(`Chunks processed: (.*?): count=(\d*), (\d*.\d*)`)
	matches = reg.FindAllStringSubmatch(input, -1)
	chunkloadingStats := make([]DynmapChunkloadingStat, len(matches))
	for i, stat := range matches {
		count, err := strconv.Atoi(stat[2])
		if err != nil {
			slog.Error("Failed to parse dynmap chunkloading stats", "err", err, "out", input)
			return nil, nil, err
		}
		duration, err := strconv.ParseFloat(stat[3], 64)
		if err != nil {
			slog.Error("Failed to parse dynmap chunkloading stats", "err", err, "out", input)
			return nil, nil, err
		}
		chunkloadingStats[i] = DynmapChunkloadingStat{
			State:    stat[1],
			Count:    count,
			Duration: duration,
		}
	}

	return renderStats, chunkloadingStats, nil
}
