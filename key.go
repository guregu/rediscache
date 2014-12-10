package rediscache

import (
	"encoding"
	"fmt"
	"strconv"
)

// Hash is a Redis hash key-field pair. It can be used as a key for Caches.
// Since Redis hash values can't expire, TTL does nothing.
type Hash struct {
	Key   string
	Field string
}

func (c Cache) keyStr() string {
	switch x := c.Key.(type) {
	case string:
		return x
	case []byte:
		return string(x)
	case fmt.Stringer:
		return x.String()
	case func() string:
		return x()
	case int64:
		return strconv.FormatInt(x, 64)
	case int32:
		return strconv.FormatInt(int64(x), 32)
	case int:
		return strconv.Itoa(x)
	case encoding.TextMarshaler:
		data, err := x.MarshalText()
		if err != nil {
			panic(err)
		}
		return string(data)
	}
	panic(fmt.Errorf("rediscache: unsupported key type %T (%#v)", c.Key, c.Key))
}
