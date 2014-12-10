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
	Key    interface{}
	Set    func() (string, error)
	Client *redis.Client

	TTL time.Duration
}

// New creates a new Cache with no expiry
// Key can be one of: string, []byte, fmt.Stringer, func() string
func New(client *redis.Client, key interface{}, setFunc func() (string, error)) Cache {
	return WithTTL(client, key, setFunc, 0)
}

// New creates a new Cache with the given TTL
// Key can be one of: string, []byte, fmt.Stringer, func() string
func WithTTL(client *redis.Client, key interface{}, setFunc func() (string, error), ttl time.Duration) Cache {
	return Cache{
		Key:    key,
		Set:    setFunc,
		Client: client,

		TTL: ttl,
	}
}

// Get will set the given pointer's value to the cached value.
// If the cached value has not been set yet, it will call the setFunc and set the returned value.
func (c Cache) Get(out interface{}) error {
	// special case: hash values
	if h, ok := c.Key.(Hash); ok {
		return c.getHash(h, out)
	}

	key := c.keyStr()
	value, err := c.Client.Get(key).Result()
	if err == nil {
		// our data is already in redis
		c.out(value, out)
		return nil
	}

	// we need to put the data in redis
	value, err = c.Set()
	if err != nil {
		return err
	}

	if c.TTL > 0 {
		if err := c.Client.SetEx(key, c.TTL, value).Err(); err != nil {
			return err
		}
	} else {
		if err := c.Client.Set(key, value).Err(); err != nil {
			return err
		}
	}

	return c.out(value, out)
}

func (c Cache) getHash(h Hash, out interface{}) error {
	value, err := c.Client.HGet(h.Key, h.Field).Result()
	if err == nil {
		// our data is already in redis
		c.out(value, out)
		return nil
	}

	// we need to put the data in redis
	value, err = c.Set()
	if err != nil {
		return err
	}

	// hashes don't support TTL
	// maybe warn for this?
	if err := c.Client.HSet(h.Key, h.Field, value).Err(); err != nil {
		return err
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
		return x.UnmarshalText([]byte(value))
	case json.Unmarshaler:
		return x.UnmarshalJSON([]byte(value))
	default:
		// hail mary
		if err := json.Unmarshal([]byte(value), out); err == nil {
			return nil
		}
	}
	return fmt.Errorf("rediscache: unsupported type %T (%#v) from %s", out, out, value)
}
