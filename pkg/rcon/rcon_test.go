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
