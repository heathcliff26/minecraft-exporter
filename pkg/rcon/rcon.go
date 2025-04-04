package rcon

import (
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/Tnze/go-mc/net"
	"github.com/hashicorp/go-version"
)

type RCONClient struct {
	addr     string
	password string
	conn     net.RCONClientConn

	version     string
	versionLock sync.RWMutex
}

// Creates an RCON client, does not create a connection immediatly
func NewRCONClient(host string, port int, password string) (*RCONClient, error) {
	if host == "" {
		return nil, ErrRCONMissingHost{}
	}
	if port <= 0 {
		return nil, ErrRCONMissingPort{}
	}
	if password == "" {
		return nil, ErrRCONMissingPassword{}
	}

	addr := host + ":" + strconv.Itoa(port)
	return &RCONClient{
		addr:     addr,
		password: password,
	}, nil
}

// Create a RCON Connection with the minecraft server
func (c *RCONClient) createConnection() error {
	slog.Debug("Creating new RCON connection")
	client, err := net.DialRCON(c.addr, c.password)
	if err != nil {
		return err
	}

	c.conn = client
	return nil
}

// Execute a remote command
func (c *RCONClient) cmd(cmd string) (string, error) {
	if c.conn == nil {
		err := c.createConnection()
		if err != nil {
			return "", err
		}
	}

	slog.Debug("RCON: Running command", "cmd", cmd)

	timeout := time.After(time.Second)
	done := make(chan bool)

	var err error
	var res string

	go func() {
		defer func() {
			done <- true
			close(done)
		}()

		err = c.conn.Cmd(cmd)
		if err != nil {
			_ = c.CloseConn()
			return
		}

		res, err = c.conn.Resp()
		if err != nil {
			_ = c.CloseConn()
			return
		}
		slog.Debug("RCON: Received response", "cmd", cmd, "res", res)
	}()

	select {
	case <-timeout:
		_ = c.CloseConn()
		done = nil
		return "", ErrRCONConnectionTimeout{}
	case <-done:
		return res, err
	}
}

// Return a list of all players currently online
func (c *RCONClient) GetPlayersOnline() []string {
	list, err := c.cmd("list")
	if err != nil {
		slog.Error("Failed to retrieve online players", "err", err)
		return []string{}
	}

	return parsePlayersOnline(list)
}

// Get the TPS statistics returned from forge
func (c *RCONClient) GetForgeTPS(variant string) ([]TPSStat, TPSStat, error) {
	res, err := c.cmd(variant + " tps")
	if err != nil {
		return nil, TPSStat{}, err
	}

	return parseForgeTPS(res)
}

// Get the count and name of all loaded forge entities
func (c *RCONClient) GetForgeEntities(variant string) ([]EntityCount, error) {
	list, err := c.cmd(variant + " entity list")
	if err != nil {
		return nil, err
	}

	return parseForgeEntities(list)
}

// Get the TPS statistics returned from paper
func (c *RCONClient) GetPaperTPS() ([]float64, error) {
	res, err := c.cmd("tps")
	if err != nil {
		return []float64{}, err
	}

	return parsePaperTPS(res)
}

// Get the render statistics returned from Dynmap
func (c *RCONClient) GetDynmapStats() ([]DynmapRenderStat, []DynmapChunkloadingStat, error) {
	res, err := c.cmd("dynmap stats")
	if err != nil {
		return nil, nil, err
	}

	return parseDynmapStats(res)
}

// Get the tick statistics returned from the "tick query" command.
// The command has been added to Minecraft Java Edition in 1.20.3
func (c *RCONClient) GetTickQuery() (TickStats, error) {
	res, err := c.cmd("tick query")
	if err != nil {
		return TickStats{}, err
	}

	return parseTickQuery(res)
}

// Update the minecraft server version
// Is concurrency safe
func (c *RCONClient) UpdateVersion(new string) {
	c.versionLock.Lock()
	defer c.versionLock.Unlock()

	c.version = new
}

func (c *RCONClient) Version() string {
	c.versionLock.RLock()
	defer c.versionLock.RUnlock()

	return c.version
}

// Returns if the version is greater or equal 1.20.3
func (c *RCONClient) V120() bool {
	versionStr := c.Version()

	if versionStr == "" {
		return false
	}

	v120 := version.Must(version.NewSemver("1.20.3"))

	v, err := version.NewSemver(versionStr)
	if err != nil {
		return false
	}

	return v120.LessThanOrEqual(v)
}

// Closes the RCON Connection and sets it to nil
func (c *RCONClient) CloseConn() error {
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

// Close the RCON connection if necessary
func (c *RCONClient) Close() error {
	return c.CloseConn()
}
