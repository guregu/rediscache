// Warning: Flushes DB #15 @ localhost:6379
package rediscache_test

import (
	"testing"
	// "time"

	"github.com/guregu/rediscache"
	"gopkg.in/redis.v2"
)

var client *redis.Client

func init() {
	client = redis.NewTCPClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   int64(15),
	})
	client.FlushDb()
}

func TestCache(t *testing.T) {
	cache := rediscache.New(client, "cachetest:1", func() (string, error) {
		return "hello world", nil
	})

	for i := 0; i < 3; i++ {
		var result string
		if err := cache.Get(&result); err != nil {
			t.Error("unexpected error:", err)
		}
		if result != "hello world" {
			t.Error("got:", result, "wanted:", "hello world")
		}
	}
}

type textunmarshaler string

func (t *textunmarshaler) UnmarshalText(data []byte) error {
	*t = textunmarshaler(data)
	return nil
}

func TestGetTextUnmarshaler(t *testing.T) {
	cache := rediscache.New(client, "cachetest:2", func() (string, error) {
		return "hello world", nil
	})

	for i := 0; i < 3; i++ {
		var result textunmarshaler
		if err := cache.Get(&result); err != nil {
			t.Error("unexpected error:", err)
		}
		if result != "hello world" {
			t.Error("got:", result, "wanted:", "hello world")
		}
	}
}
