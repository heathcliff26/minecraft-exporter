package uuid

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testUUID = "6f003e33-7076-4e45-a270-87841b218ec7"
	testName = "Heathcliff26"
)

func TestFetchName(t *testing.T) {
	c := NewUUIDCache(time.Hour)

	result, err := c.GetNameFromUUID(testUUID)

	assert := assert.New(t)

	assert.Nil(err)
	assert.Equal(testName, result)
}

func TestCacheHit(t *testing.T) {
	name := "NotTheActualName"
	c := NewUUIDCache(time.Hour)
	c.Items[testUUID] = UUIDCacheItem{
		Name:      name,
		Timestamp: time.Now(),
	}

	result, err := c.GetNameFromUUID(testUUID)

	assert := assert.New(t)

	assert.Nil(err)
	assert.Equal(name, result)
}

func TestCacheExpired(t *testing.T) {
	c := NewUUIDCache(time.Hour)
	c.Items[testUUID] = UUIDCacheItem{
		Name:      "NotTheActualName",
		Timestamp: time.Now().Add(-2 * time.Hour),
	}

	result, err := c.GetNameFromUUID(testUUID)

	assert := assert.New(t)

	assert.Nil(err)
	assert.Equal(testName, result)
}
