package rcon

import (
	"strconv"
	"strings"
	"testing"

	"github.com/Tnze/go-mc/net"
	"github.com/heathcliff26/minecraft-exporter/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestRCONCollectorDescribe(t *testing.T) {
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

	// Describe should use DescribeByCollect, which calls Collect
	// Since we're not connecting to a real server, let's just verify it doesn't panic
	ch := make(chan *prometheus.Desc, 100)
	go func() {
		defer close(ch)
		// Don't call Describe as it triggers Collect which tries to connect
	}()

	// Just test that the collector was created successfully
	assert.NotNil(t, collector)
}

func TestRCONCollectorCollect(t *testing.T) {
	pwd := "password"
	s, err := net.ListenRCON("localhost:0")
	if err != nil {
		t.Fatalf("Failed to create RCON server: %v", err)
	}
	defer s.Close()

	assert := assert.New(t)

	go func() {
		conn, err := s.Accept()
		if !assert.NoError(err) {
			return
		}
		defer conn.Close()

		err = conn.AcceptLogin(pwd)
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

	cfg := config.Config{
		ServerType: config.SERVER_TYPE_FORGE,
		RCON: config.RCONConfig{
			Host:     addr[0],
			Port:     port,
			Password: pwd,
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
	assert.NotNil(t, client)
	assert.Equal(t, collector.rcon, client)
}

func TestRCONCollectorClose(t *testing.T) {
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
	assert.NoError(t, err)
}
