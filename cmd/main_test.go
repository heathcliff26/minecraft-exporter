package main

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerRootHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err, "Should create request")

	rr := httptest.NewRecorder()

	ServerRootHandler(rr, req)

	assert := assert.New(t)

	assert.Equal(http.StatusOK, rr.Code)
	body := rr.Body.String()
	assert.Contains(body, "<html>")
	assert.Contains(body, "</html>")
	assert.Contains(body, "<a href='/metrics'>")
}

func TestShowVersion(t *testing.T) {
	if os.Getenv("RUN_CRASH_TEST") == "1" {
		_ = flag.CommandLine.Parse([]string{"-version"})
		main()
	}
	execExitTest(t, "TestShowVersion", false)
}

func TestErrorLoadConfiguration(t *testing.T) {
	if os.Getenv("RUN_CRASH_TEST") == "1" {
		_ = flag.CommandLine.Parse([]string{"-config", "/not/a/valid/path"})
		main()
	}
	execExitTest(t, "TestErrorLoadConfiguration", true)
}

func TestErrorCreateSaveCollector(t *testing.T) {
	if os.Getenv("RUN_CRASH_TEST") == "1" {
		_ = flag.CommandLine.Parse([]string{"-config", "testdata/invalid-world-dir.yaml"})
		main()
	}
	execExitTest(t, "TestErrorCreateSaveCollector", true)
}

func TestErrorCreateRCONCollector(t *testing.T) {
	if os.Getenv("RUN_CRASH_TEST") == "1" {
		_ = flag.CommandLine.Parse([]string{"-config", "testdata/invalid-rcon.yaml"})
		main()
	}
	execExitTest(t, "TestErrorCreateRCONCollector", true)
}

func TestErrorStartServer(t *testing.T) {
	if os.Getenv("RUN_CRASH_TEST") == "1" {
		_ = flag.CommandLine.Parse([]string{"-config", "testdata/invalid-port.yaml"})
		main()
	}
	execExitTest(t, "TestErrorStartServer", true)
}

func execExitTest(t *testing.T, test string, exitsError bool) {
	cmd := exec.Command(os.Args[0], "-test.run="+test)
	cmd.Env = append(os.Environ(), "RUN_CRASH_TEST=1")
	out, err := cmd.CombinedOutput()
	t.Log("Output:\n", string(out))
	if exitsError && err == nil {
		t.Fatal("Process exited without error")
	} else if !exitsError && err == nil {
		return
	}
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}
