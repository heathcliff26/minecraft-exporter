package uuid

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type UUIDCache struct {
	Items     map[string]UUIDCacheItem
	CacheTime time.Duration
}

type UUIDCacheItem struct {
	Name      string
	Timestamp time.Time
}

type MojanUUIDToProfileResponse struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// Returns a new UUID Cache
func NewUUIDCache(cacheTime time.Duration) *UUIDCache {
	return &UUIDCache{
		Items:     make(map[string]UUIDCacheItem),
		CacheTime: cacheTime,
	}
}

// Either return the name from cache or fetch from the server if the name is either
// not cached or the cache expired.
// If the uuid is not known to mojang, return the uuid instead
func (c *UUIDCache) GetNameFromUUID(uuid string) (string, error) {
	now := time.Now()

	item, ok := c.Items[uuid]
	if ok {
		if item.Timestamp.Add(c.CacheTime).After(now) {
			return item.Name, nil
		} else {
			delete(c.Items, uuid)
		}
	}

	res, err := http.Get("https://sessionserver.mojang.com/session/minecraft/profile/" + uuid)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var name string
	switch res.StatusCode {
	case http.StatusOK:
		var result MojanUUIDToProfileResponse
		err = json.NewDecoder(res.Body).Decode(&result)
		if err != nil {
			return "", err
		}
		name = result.Name
	case http.StatusNoContent:
		name = uuid
		slog.Warn("Found no minecraft account for the uuid, falling back to using uuid as name", slog.String("uuid", uuid))
	default:
		return "", NewErrHttpRequestFailed(res.StatusCode, res.Body)
	}

	c.Items[uuid] = UUIDCacheItem{
		Name:      name,
		Timestamp: now,
	}

	return name, nil
}
