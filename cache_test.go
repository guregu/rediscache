// Tests use a gred server that runs at localhost:6389
package rediscache_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/guregu/gredlib"
	"github.com/guregu/rediscache"
	"gopkg.in/redis.v2"
)

var client *redis.Client

var listenAddr = "localhost:6389"

func init() {
	go gredlib.ListenAndServe("tcp", listenAddr)

	client = redis.NewTCPClient(&redis.Options{
		Addr: listenAddr,
		DB:   int64(15),
	})
}

func TestCache(t *testing.T) {
	defer client.Del("rediscache_test:1")
	cache := rediscache.Cache{
		Client: client,
		Key:    "rediscache_test:1",
		Getter: func() (string, error) {
			return "hello world", nil
		},
	}

	for i := 0; i < 3; i++ {
		var result string
		t.Logf("%#v", cache)
		if err := cache.Get(&result); err != nil {
			t.Error("unexpected error:", err)
		}
		if result != "hello world" {
			t.Error("got:", result, "wanted:", "hello world")
		}
	}
}

func TestHashCache(t *testing.T) {
	defer client.Del("rediscache_test")
	cache := rediscache.Cache{
		Client: client,
		Key:    "rediscache_test",
		Field:  "hash",
		Getter: func() (string, error) {
			return "hello world", nil
		},
	}

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
	cache := rediscache.Cache{
		Client: client,
		Key:    "rediscache_test:2",
		Getter: func() (string, error) {
			return "hello world", nil
		},
		TTL: time.Second,
	}

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

func TestGetInts(t *testing.T) {
	cache := rediscache.Cache{
		Client: client,
		Key:    "rediscache_test:3",
		Getter: func() (string, error) {
			return "12345", nil
		},
		TTL: time.Second,
	}

	var result int
	if err := cache.Get(&result); err != nil {
		t.Error("unexpected error:", err)
	}
	if result != 12345 {
		t.Error("got:", result, "wanted:", 12345)
	}

	var result64 int64
	if err := cache.Get(&result64); err != nil {
		t.Error("unexpected error:", err)
	}
	if result64 != 12345 {
		t.Error("got:", result64, "wanted:", 12345)
	}

	var result32 int32
	if err := cache.Get(&result32); err != nil {
		t.Error("unexpected error:", err)
	}
	if result32 != 12345 {
		t.Error("got:", result32, "wanted:", 12345)
	}
}

func TestKeyFunc(t *testing.T) {
	keyfunc := func() string {
		return "rediscache_test:4"
	}

	cache := rediscache.Cache{
		Client: client,
		Key:    keyfunc(),
		Getter: func() (string, error) {
			return "12345", nil
		},
		TTL: time.Second,
	}

	var result int
	if err := cache.Get(&result); err != nil {
		t.Error("unexpected error:", err)
	}
	if result != 12345 {
		t.Error("got:", result, "wanted:", 12345)
	}
}

func TestUnmarshalJSON(t *testing.T) {
	type testType struct {
		A int
		B string
	}
	expected := testType{1, "qqq"}

	cache := rediscache.Cache{
		Client: client,
		Key:    "rediscache_test:5",
		Getter: func() (string, error) {
			return `{"a": 1, "b": "qqq"}`, nil
		},
		TTL: time.Second,
	}

	var got testType
	cache.Get(&got)
	if !reflect.DeepEqual(expected, got) {
		t.Error("got:", got, "wanted:", expected)
	}
}
