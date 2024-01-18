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
		if !assert.Nil(err) {
			t.Logf("[Server] Failed to accept connection: %v", err)
			return
		}
		defer conn.Close()

		err = conn.AcceptLogin(pwd)
		if !assert.Nil(err) {
			t.Logf("[Server] Failed to accept login: %v", err)
			return
		}

		for i := range ch {
			if i == "Fail" {
				conn.Close()
				continue
			}

			cmd, err := conn.AcceptCmd()
			if !assert.Nil(err) {
				continue
			}
			var resp string
			if assert.Equal(i, cmd) {
				resp = "success"
			} else {
				resp = "fail"
			}
			err = conn.RespCmd(resp)
			assert.Nil(err)
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
	assert.Nil(err)
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
		if !assert.Nil(err) {
			t.Logf("[Server] Failed to accept connection: %v", err)
			return
		}
		defer conn.Close()

		err = conn.AcceptLogin(pwd)
		if !assert.Nil(err) {
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
