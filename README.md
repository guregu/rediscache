## rediscache [![GoDoc](https://godoc.org/github.com/guregu/rediscache?status.svg)](https://godoc.org/github.com/guregu/rediscache) 
`import "github.com/guregu/rediscache"`

rediscache is a small library for caching data in Redis, similar to how one might define a groupcache topic. When getting a value, it will automatically convert the stringly-typed Redis data to whatever you pass to it, kind of like how json.Unmarshal works. I'm still playing around with this, so consider it unstable. 

### Usage

```go
var thingCache = rediscache.New(client, "my:thing:key", func() (string, error) {
		var result string
		// do some expensive calculations
		return result, nil
	})

func doSomethingWithTheData() {
	var thing string
	err := thingCache.get(&thing)
	if err != nil {
		...
	}
	// use thing
}
```