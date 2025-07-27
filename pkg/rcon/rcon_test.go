package rcon

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Tnze/go-mc/net"
	"github.com/stretchr/testify/assert"
)

func TestNewRCONClient(t *testing.T) {
	tMatrix := []struct {
		Name     string
		Host     string
		Port     int
		Password string
		Error    error
	}{
		{"Success", "localhost", 25575, "password", nil},
		{"MissingHost", "", 25575, "password", ErrRCONMissingHost{}},
		{"MissingPort", "localhost", 0, "password", ErrRCONMissingPort{}},
		{"MissingPassword", "localhost", 25575, "", ErrRCONMissingPassword{}},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			c, err := NewRCONClient(tCase.Host, tCase.Port, tCase.Password)
			assert := assert.New(t)
			assert.Equal(tCase.Error, err)
			if tCase.Error == nil {
				assert.NotEmpty(c)
			} else {
				assert.Nil(c)
			}
		})
	}
}

func TestCloseConnectionOnError(t *testing.T) {
	pwd := "password"
	s, err := net.ListenRCON("localhost:0")
	if err != nil {
		t.Fatalf("Failed to create RCON server: %v", err)
	}

	ch := make(chan string, 1)

	assert := assert.New(t)

	go func() {
		conn, err := s.Accept()
		if !assert.NoError(err) {
			t.Logf("[Server] Failed to accept connection: %v", err)
			return
		}
		defer conn.Close()

		err = conn.AcceptLogin(pwd)
		if !assert.NoError(err) {
			t.Logf("[Server] Failed to accept login: %v", err)
			return
		}

		for i := range ch {
			if i == "Fail" {
				conn.Close()
				continue
			}

			cmd, err := conn.AcceptCmd()
			if !assert.NoError(err) {
				continue
			}
			var resp string
			if assert.Equal(i, cmd) {
				resp = "success"
			} else {
				resp = "fail"
			}
			err = conn.RespCmd(resp)
			assert.NoError(err)
		}
		t.Log("Closed the goroutine") // TODO: remove
	}()

	addr := strings.Split(s.Listener.Addr().String(), ":")
	port, err := strconv.Atoi(addr[1])
	if err != nil {
		t.Fatalf("Failed to convert addr to port: %v", err)
	}

	c, err := NewRCONClient(addr[0], port, pwd)
	if err != nil {
		t.Fatalf("Failed to create RCON client: %v", err)
	}

	assert.Nil(c.conn, "New Client should have no connection")

	ch <- "First cmd"
	res, err := c.cmd("First cmd")
	assert.NoError(err)
	assert.Equal("success", res)
	assert.NotNil(c.conn, "Should create connection if none exists")

	ch <- "Fail"
	res, err = c.cmd("")
	assert.Error(err)
	assert.Equal("", res)
	assert.Nil(c.conn, "Should close connection on error")
}

func TestTimeout(t *testing.T) {
	pwd := "password"
	s, err := net.ListenRCON("localhost:0")
	if err != nil {
		t.Fatalf("Failed to create RCON server: %v", err)
	}

	assert := assert.New(t)

	go func() {
		conn, err := s.Accept()
		if !assert.NoError(err) {
			t.Logf("[Server] Failed to accept connection: %v", err)
			return
		}
		defer conn.Close()

		err = conn.AcceptLogin(pwd)
		if !assert.NoError(err) {
			t.Logf("[Server] Failed to accept login: %v", err)
			return
		}

		done := time.After(3 * time.Second)
		// Wait for test timeout
		<-done
	}()

	addr := strings.Split(s.Listener.Addr().String(), ":")
	port, err := strconv.Atoi(addr[1])
	if err != nil {
		t.Fatalf("Failed to convert addr to port: %v", err)
	}

	c, err := NewRCONClient(addr[0], port, pwd)
	if err != nil {
		t.Fatalf("Failed to create RCON client: %v", err)
	}

	_, err = c.cmd("Test")
	assert.Equal("rcon.ErrRCONConnectionTimeout", reflect.TypeOf(err).String())
}

func TestUpdateVersion(t *testing.T) {
	c := &RCONClient{}

	c.UpdateVersion("1.21.0")
	assert.Equal(t, "1.21.0", c.version, "Should update version")

	c.versionLock.RLock()
	t.Cleanup(c.versionLock.RUnlock)

	ch := make(chan struct{}, 1)
	go func() {
		c.UpdateVersion("1.21.1")
		ch <- struct{}{}
	}()

	select {
	case <-ch:
		t.Fail()
	case <-time.After(time.Second):
	}
}

func TestV120(t *testing.T) {
	assert := assert.New(t)

	c := &RCONClient{}

	assert.False(c.V120(), "Should return false when version is empty")

	for _, version := range []string{"1.19.0", "1.20.0", "1.20.2", "1.20.3-pre2"} {
		c.UpdateVersion(version)
		assert.Falsef(c.V120(), "Should return false if version is %s", version)
	}
	for _, version := range []string{"1.20.3", "1.20.4-pre1", "1.20.4", "1.21.0"} {
		c.UpdateVersion(version)
		assert.Truef(c.V120(), "Should return true if version is %s", version)
	}
}

func TestGetPlayersOnline(t *testing.T) {
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

		// First request - return player list
		cmd, err := conn.AcceptCmd()
		if !assert.NoError(err) {
			return
		}
		assert.Equal("list", cmd)
		err = conn.RespCmd("There are 2/10 players online:Player1, Player2")
		assert.NoError(err)

		// Second request - return empty list
		cmd, err = conn.AcceptCmd()
		if !assert.NoError(err) {
			return
		}
		assert.Equal("list", cmd)
		err = conn.RespCmd("There are 0/10 players online:")
		assert.NoError(err)
	}()

	addr := strings.Split(s.Listener.Addr().String(), ":")
	port, err := strconv.Atoi(addr[1])
	if err != nil {
		t.Fatalf("Failed to convert addr to port: %v", err)
	}

	c, err := NewRCONClient(addr[0], port, pwd)
	if err != nil {
		t.Fatalf("Failed to create RCON client: %v", err)
	}

	// Test with players online
	players := c.GetPlayersOnline()
	assert.Equal([]string{"Player1", "Player2"}, players)

	// Test with no players online
	players = c.GetPlayersOnline()
	assert.Equal([]string{}, players)
}

func TestGetForgeTPS(t *testing.T) {
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

		cmd, err := conn.AcceptCmd()
		if !assert.NoError(err) {
			return
		}
		assert.Equal("forge tps", cmd)

		forgeOutput := "Dim  0 (DIM_0) : Mean tick time: 7.672 ms. Mean TPS: 20.000Overall : Mean tick time: 8.037 ms. Mean TPS: 20.000"
		err = conn.RespCmd(forgeOutput)
		assert.NoError(err)
	}()

	addr := strings.Split(s.Listener.Addr().String(), ":")
	port, err := strconv.Atoi(addr[1])
	if err != nil {
		t.Fatalf("Failed to convert addr to port: %v", err)
	}

	c, err := NewRCONClient(addr[0], port, pwd)
	if err != nil {
		t.Fatalf("Failed to create RCON client: %v", err)
	}

	dimStats, overallStat, err := c.GetForgeTPS("forge")
	assert.NoError(err)
	assert.Len(dimStats, 1)
	assert.Equal("0", dimStats[0].ID)
	assert.Equal("DIM_0", dimStats[0].Name)
	assert.Equal(7.672, dimStats[0].Ticktime)
	assert.Equal(20.0, dimStats[0].TPS)
	assert.Equal(8.037, overallStat.Ticktime)
	assert.Equal(20.0, overallStat.TPS)
}

func TestGetForgeEntities(t *testing.T) {
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

		cmd, err := conn.AcceptCmd()
		if !assert.NoError(err) {
			return
		}
		assert.Equal("forge entity list", cmd)

		entityOutput := "Total: 24  12: minecraft:chicken  5: minecraft:cow  2: minecraft:item"
		err = conn.RespCmd(entityOutput)
		assert.NoError(err)
	}()

	addr := strings.Split(s.Listener.Addr().String(), ":")
	port, err := strconv.Atoi(addr[1])
	if err != nil {
		t.Fatalf("Failed to convert addr to port: %v", err)
	}

	c, err := NewRCONClient(addr[0], port, pwd)
	if err != nil {
		t.Fatalf("Failed to create RCON client: %v", err)
	}

	entities, err := c.GetForgeEntities("forge")
	assert.NoError(err)
	assert.Len(entities, 3)
	assert.Equal("minecraft:chicken", entities[0].Name)
	assert.Equal(12, entities[0].Count)
	assert.Equal("minecraft:cow", entities[1].Name)
	assert.Equal(5, entities[1].Count)
}

func TestGetPaperTPS(t *testing.T) {
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

		cmd, err := conn.AcceptCmd()
		if !assert.NoError(err) {
			return
		}
		assert.Equal("tps", cmd)

		paperOutput := "§6TPS from last 1m, 5m, 15m: §a20.0§r, §a20.0§r, §a20.0\n"
		err = conn.RespCmd(paperOutput)
		assert.NoError(err)
	}()

	addr := strings.Split(s.Listener.Addr().String(), ":")
	port, err := strconv.Atoi(addr[1])
	if err != nil {
		t.Fatalf("Failed to convert addr to port: %v", err)
	}

	c, err := NewRCONClient(addr[0], port, pwd)
	if err != nil {
		t.Fatalf("Failed to create RCON client: %v", err)
	}

	tps, err := c.GetPaperTPS()
	assert.NoError(err)
	assert.Equal([]float64{20.0, 20.0, 20.0}, tps)
}

func TestGetDynmapStats(t *testing.T) {
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

		cmd, err := conn.AcceptCmd()
		if !assert.NoError(err) {
			return
		}
		assert.Equal("dynmap stats", cmd)

		dynmapOutput := "Tile Render Statistics:\n  world.cave: processed=50672, rendered=50672, updated=4757, transparent=0\nChunk Loading Statistics:\n  Chunks processed: Cached: count=3289892, 0.00 msec/chunk\n"
		err = conn.RespCmd(dynmapOutput)
		assert.NoError(err)
	}()

	addr := strings.Split(s.Listener.Addr().String(), ":")
	port, err := strconv.Atoi(addr[1])
	if err != nil {
		t.Fatalf("Failed to convert addr to port: %v", err)
	}

	c, err := NewRCONClient(addr[0], port, pwd)
	if err != nil {
		t.Fatalf("Failed to create RCON client: %v", err)
	}

	renderStats, chunkStats, err := c.GetDynmapStats()
	assert.NoError(err)
	assert.Len(renderStats, 1)
	assert.Equal("world.cave", renderStats[0].Dim)
	assert.Equal(50672, renderStats[0].Processed)
	assert.Len(chunkStats, 1)
	assert.Equal("Cached", chunkStats[0].State)
	assert.Equal(3289892, chunkStats[0].Count)
}

func TestGetTickQuery(t *testing.T) {
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

		cmd, err := conn.AcceptCmd()
		if !assert.NoError(err) {
			return
		}
		assert.Equal("tick query", cmd)

		tickOutput := "The game is running normallyTarget tick rate: 20.0 per second.\nAverage time per tick: 7.7ms (Target: 50.0ms)Percentiles: P50: 7.4ms P95: 9.9ms P99: 11.1ms, sample: 100"
		err = conn.RespCmd(tickOutput)
		assert.NoError(err)
	}()

	addr := strings.Split(s.Listener.Addr().String(), ":")
	port, err := strconv.Atoi(addr[1])
	if err != nil {
		t.Fatalf("Failed to convert addr to port: %v", err)
	}

	c, err := NewRCONClient(addr[0], port, pwd)
	if err != nil {
		t.Fatalf("Failed to create RCON client: %v", err)
	}

	tickStats, err := c.GetTickQuery()
	assert.NoError(err)
	assert.Equal(20.0, tickStats.Target)
	assert.Equal(7.7, tickStats.Average)
	assert.Equal(7.4, tickStats.P50)
	assert.Equal(9.9, tickStats.P95)
	assert.Equal(11.1, tickStats.P99)
}

func TestClose(t *testing.T) {
	c := &RCONClient{}
	err := c.Close()
	assert.NoError(t, err)

	// Test with existing connection
	pwd := "password"
	s, err := net.ListenRCON("localhost:0")
	if err != nil {
		t.Fatalf("Failed to create RCON server: %v", err)
	}
	defer s.Close()

	go func() {
		conn, err := s.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_ = conn.AcceptLogin(pwd)
	}()

	addr := strings.Split(s.Listener.Addr().String(), ":")
	port, err := strconv.Atoi(addr[1])
	if err != nil {
		t.Fatalf("Failed to convert addr to port: %v", err)
	}

	c, err = NewRCONClient(addr[0], port, pwd)
	if err != nil {
		t.Fatalf("Failed to create RCON client: %v", err)
	}

	// Create connection
	err = c.createConnection()
	assert.NoError(t, err)
	assert.NotNil(t, c.conn)

	// Test close
	err = c.Close()
	assert.NoError(t, err)
	assert.Nil(t, c.conn)
}
