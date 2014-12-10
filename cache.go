// Package rediscache provies a simple way to cache values in Redis.
// It uses a somewhat similar interface to groupcache.
package rediscache

import (
	"encoding"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"gopkg.in/redis.v2"
)

// Cache represents one GET/SET Redis-cached value.
// It will try to get the value from Redis, setting the value with the given setFunc if necessary.
type Cache struct {
	Key string

	set    func() (string, error)
	ttl    time.Duration
	client *redis.Client
}

// New creates a new Cache with no expiry
func New(client *redis.Client, key string, setFunc func() (string, error)) *Cache {
	return NewWithTTL(client, key, setFunc, 0)
}

// New creates a new Cache with the given TTL
func NewWithTTL(client *redis.Client, key string, setFunc func() (string, error), ttl time.Duration) *Cache {
	return &Cache{
		Key: key,

		set:    setFunc,
		ttl:    ttl,
		client: client,
	}
}

// Get will set the given pointer's value to the cached value.
// If the cached value has not been set yet, it will call the setFunc and set the returned value.
func (c Cache) Get(out interface{}) error {
	value, err := c.client.Get(c.Key).Result()
	if err == nil {
		// our data is already in redis
		c.out(value, out)
		return nil
	}

	// we need to put the data in redis
	value, err = c.set()
	if err != nil {
		return err
	}

	if c.ttl > 0 {
		if err := c.client.SetEx(c.Key, c.ttl, value).Err(); err != nil {
			return err
		}
	} else {
		if err := c.client.Set(c.Key, value).Err(); err != nil {
			return err
		}
	}

	return c.out(value, out)
}

func (c Cache) out(value string, out interface{}) error {
	switch x := out.(type) {
	case *string:
		*x = value
		return nil
	case *[]byte:
		*x = []byte(value)
		return nil
	case *int64:
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		*x = n
		return nil
	case *int32:
		n, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return err
		}
		*x = int32(n)
		return nil
	case *int:
		n, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		*x = n
		return nil
	case encoding.TextUnmarshaler:
		x.UnmarshalText([]byte(value))
		return nil
	case json.Unmarshaler:
		x.UnmarshalJSON([]byte(value))
	default:
		// hail mary
		if err := json.Unmarshal([]byte(value), out); err == nil {
			return nil
		}
	}
	return fmt.Errorf("unknown type %T", out)
}
