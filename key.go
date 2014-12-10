package rediscache

import (
	"encoding"
	"fmt"
	"strconv"
)

// keyStr tries its best to turn a key or field interface{} into a string
func keyStr(v interface{}) string {
	switch x := v.(type) {
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
	panic(fmt.Errorf("rediscache: unsupported key type %T (%#v)", v, v))
}
