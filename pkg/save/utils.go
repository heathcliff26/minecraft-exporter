package save

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"

	"github.com/Tnze/go-mc/nbt"
	"github.com/prometheus/client_golang/prometheus"
)

// Check if the given path is a directory
func isDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

// Count the total number of earned advancements
func countAdvancements(advancements map[string]Advancement) uint {
	var i uint
	for _, value := range advancements {
		if value.Done {
			i++
		}
	}
	return i
}

// Read a nbt file and parse it to the given struct.
// Assumes the file is gzip compressed.
func readNBT(path string, target interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	fReader, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer fReader.Close()

	_, err = nbt.NewDecoder(fReader).Decode(target)
	if err != nil {
		return err
	}

	return nil
}

// Read a json file and parse it to the given struct
func readJSON(path string, target interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, target)
	if err != nil {
		return err
	}

	return nil
}

// Count up all numbers of the map
func countTotal(values map[string]int) int {
	var total = 0

	for _, v := range values {
		total = total + v
	}
	return total
}

// Convert a given map to metrics
func mapToMetrics(ch chan<- prometheus.Metric, desc *prometheus.Desc, values map[string]int, player string) {
	for k, v := range values {
		ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, float64(v), player, k)
	}
}
