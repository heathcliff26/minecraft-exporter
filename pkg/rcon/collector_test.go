package rcon

import (
	"strconv"
	"strings"
	"testing"

	"github.com/Tnze/go-mc/net"
	"github.com/heathcliff26/minecraft-exporter/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testRCONPassword = "testpassword"

func TestRCONCollectorDescribe(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	s, port := newTestServer(t)
	defer s.Close()

	cfg := config.Config{
		ServerType: config.SERVER_TYPE_FORGE,
		RCON: config.RCONConfig{
			Host:     "localhost",
			Port:     port,
			Password: testRCONPassword,
		},
	}

	c, err := NewRCONCollector(cfg)
	require.NoError(err, "Should create Collector")

	expectedDescCount := 17

	ch := make(chan *prometheus.Desc)
	expectedDescs := make([]*prometheus.Desc, 0, expectedDescCount)
	go func() {
		prometheus.DescribeByCollect(c, ch)
		close(ch)
	}()
	for desc := range ch {
		expectedDescs = append(expectedDescs, desc)
	}

	ch = make(chan *prometheus.Desc)
	result := make([]*prometheus.Desc, 0, expectedDescCount)
	go func() {
		c.Describe(ch)
		close(ch)
	}()
	for desc := range ch {
		result = append(result, desc)
	}

	// Can't compare all descriptions, as only a subset of metrics will be returned based on the given configuration.
	for _, desc := range expectedDescs {
		assert.Contains(result, desc, "Descriptor should be present in Describe output")
	}
	assert.Len(result, expectedDescCount, "Should have correct number of descriptors")
}

func TestRCONCollectorCollect(t *testing.T) {
	assert := assert.New(t)

	s, port := newTestServer(t)
	defer s.Close()

	cfg := config.Config{
		ServerType: config.SERVER_TYPE_FORGE,
		RCON: config.RCONConfig{
			Host:     "localhost",
			Port:     port,
			Password: testRCONPassword,
		},
	}

	collector, err := NewRCONCollector(cfg)
	if err != nil {
		t.Fatalf("Failed to create collector: %v", err)
	}

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	collector.Collect(ch)
	close(ch)

	// Should have collected some metrics
	count := 0
	for range ch {
		count++
	}
	assert.True(count > 0, "Should have collected metrics")
}

func TestNewRCONCollector(t *testing.T) {
	assert := assert.New(t)

	// Test successful creation
	cfg := config.Config{
		ServerType: config.SERVER_TYPE_FORGE,
		RCON: config.RCONConfig{
			Enable:   true,
			Host:     "localhost",
			Port:     25575,
			Password: "password",
		},
	}

	collector, err := NewRCONCollector(cfg)
	assert.NoError(err)
	assert.NotNil(collector)
	assert.Equal(config.SERVER_TYPE_FORGE, collector.ServerType)
	assert.False(collector.DynmapEnabled)

	// Test with missing host
	cfg.RCON.Host = ""
	collector, err = NewRCONCollector(cfg)
	assert.Error(err)
	assert.Nil(collector)
}

func TestRCONCollectorClient(t *testing.T) {
	assert := assert.New(t)

	cfg := config.Config{
		RCON: config.RCONConfig{
			Host:     "localhost",
			Port:     25575,
			Password: "password",
		},
	}

	collector, err := NewRCONCollector(cfg)
	if err != nil {
		t.Fatalf("Failed to create collector: %v", err)
	}

	client := collector.Client()
	assert.NotNil(client)
	assert.Equal(collector.rcon, client)
}

func TestRCONCollectorClose(t *testing.T) {
	assert := assert.New(t)

	cfg := config.Config{
		RCON: config.RCONConfig{
			Host:     "localhost",
			Port:     25575,
			Password: "password",
		},
	}

	collector, err := NewRCONCollector(cfg)
	if err != nil {
		t.Fatalf("Failed to create collector: %v", err)
	}

	err = collector.Close()
	assert.NoError(err)
}

func newTestServer(t *testing.T) (*net.RCONListener, int) {
	s, err := net.ListenRCON("localhost:0")
	if err != nil {
		t.Fatalf("Failed to create RCON server: %v", err)
	}

	assert := assert.New(t)

	go func() {
		conn, err := s.Accept()
		if !assert.NoError(err) {
			return
		}
		defer conn.Close()

		err = conn.AcceptLogin(testRCONPassword)
		if !assert.NoError(err) {
			return
		}

		// Handle list command
		cmd, err := conn.AcceptCmd()
		if !assert.NoError(err) {
			return
		}
		if cmd == "list" {
			err = conn.RespCmd("There are 1/10 players online:TestPlayer")
			assert.NoError(err)
		}

		// Handle forge tps command if present
		cmd, err = conn.AcceptCmd()
		if err == nil && cmd == "forge tps" {
			err = conn.RespCmd("Dim  0 (DIM_0) : Mean tick time: 7.672 ms. Mean TPS: 20.000Overall : Mean tick time: 8.037 ms. Mean TPS: 20.000")
			assert.NoError(err)
		}

		// Handle forge entity list command if present
		cmd, err = conn.AcceptCmd()
		if err == nil && cmd == "forge entity list" {
			err = conn.RespCmd("Total: 12  12: minecraft:chicken")
			assert.NoError(err)
		}
	}()

	addr := strings.Split(s.Listener.Addr().String(), ":")
	port, err := strconv.Atoi(addr[1])
	if err != nil {
		t.Fatalf("Failed to convert addr to port: %v", err)
	}

	return s, port
}
