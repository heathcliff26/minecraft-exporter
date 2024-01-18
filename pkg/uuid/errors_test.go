package uuid

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewErrHttpRequestFailed(t *testing.T) {
	result := &ErrHttpRequestFailed{StatusCode: 400, Body: "testresult"}
	r := io.NopCloser(strings.NewReader(result.Body))
	defer r.Close()
	err := NewErrHttpRequestFailed(400, r)
	assert.Equal(t, result, err)
}
