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

// Cache represents one Redis-cached value.
// It will try to get the value from Redis, updating the cahe using the getter if necessary.
// If the Field is empty, it will use GET and SET operations.
// If the Field is not empty, it will use HGET and HSET. HGET/HSET don't support TTLs.
type Cache struct {
	Key    interface{}
	Field  interface{}
	Getter func() (string, error)
	Client *redis.Client

	TTL time.Duration // optional
}

// Get will set the given pointer's value to the cached value.
// If the cached value has not been set yet, it will call the setFunc and set the returned value.
func (c Cache) Get(out interface{}) error {
	// special case: hash values
	if c.Field != nil {
		return c.getHash(out)
	}

	key := keyStr(c.Key)
	value, err := c.Client.Get(key).Result()
	// if cached
	if err == nil {
		c.out(value, out)
		return nil
	}

	// otherwise get the data and put it in redis
	value, err = c.Getter()
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

func (c Cache) getHash(out interface{}) error {
	key, field := keyStr(c.Key), keyStr(c.Field)
	value, err := c.Client.HGet(key, field).Result()
	if err == nil {
		c.out(value, out)
		return nil
	}

	value, err = c.Getter()
	if err != nil {
		return err
	}

	// hashes don't support TTL
	// maybe warn for this?
	if err := c.Client.HSet(key, field, value).Err(); err != nil {
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
	case *json.RawMessage:
		*x = json.RawMessage([]byte(value))
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
