package uuid

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testUUID = "6f003e33-7076-4e45-a270-87841b218ec7"
	testName = "Heathcliff26"

	noAccountUUID = "12345678-90ab-cdef-0000-123456789abc"
)

func TestFetchName(t *testing.T) {
	c := NewUUIDCache(time.Hour)

	result, err := c.GetNameFromUUID(testUUID)

	assert := assert.New(t)

	assert.NoError(err)
	assert.Equal(testName, result, "Should return the name")
	assert.Equal(testName, c.Items[testUUID].Name, "Should have saved the result in the cache")
}

func TestNoAccountForUUID(t *testing.T) {
	c := NewUUIDCache(time.Hour)

	result, err := c.GetNameFromUUID(noAccountUUID)

	assert := assert.New(t)

	assert.NoError(err)
	assert.Equal(noAccountUUID, result, "Should return the uuid")
	assert.Equal(noAccountUUID, c.Items[noAccountUUID].Name, "Should have saved the result in the cache")
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

	assert.NoError(err)
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

	assert.NoError(err)
	assert.Equal(testName, result)
}
