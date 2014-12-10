## rediscache [![GoDoc](https://godoc.org/github.com/guregu/rediscache?status.svg)](https://godoc.org/github.com/guregu/rediscache) 
`import "github.com/guregu/rediscache"`

rediscache is a small library for caching data in Redis, similar to how one might define a groupcache topic. When getting a value, it will automatically convert the stringly-typed Redis data to whatever you pass to it, kind of like how json.Unmarshal works. I'm still playing around with this, so consider it unstable. 

### Rationale
I found myself writing code that repeats these actions over and over:

* Turn some kind of an ID into a Redis key
* Try to get the value from Redis
* If missing from the cache, calculate the value and set it in Redis 
* Convert the result string into something usable 

This is a generic way to do the above, inspired by groupcache and the standard JSON package.

### Usage

```go
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
```

#### Example
```go
var redisClient *redis.Client
var thingCache = rediscache.Cache{
	Client: redisClient,
	Key:    "my:thing",
	Getter: fetchThing, // func() (string, error)

	TTL: 1 * time.Hour, // optional
}

// Redis is stringly-typed, so you have to give rediscache a string or an error.
func fetchThing() (string, error) {
	var thing int
	// do some expensive calculations
	return strconv.Itoa(thing), nil
}

func doSomethingWithTheData() {
	var thing int
	err := thingCache.Get(&thing) // works similar to the json package
	if err != nil {
		...
	}
	// use thing
}
```

#### Polymorphism
In the world of Redis, everything is a string, but that is not so true in the real world. Key and Field take an interface{} and try their best to turn it into a string, including checking for `fmt.Stringer` and `encoding.TextMarshaler`. You can also give them a `func() String`. 

```go
	dailyCache := rediscache.Cache{
		Client: client,
		Key:    func() string {
			date := time.Now().Format("2006-01-02")
			return "thing:" + date, nil
		},
		Getter: func() (string, error) {
			var todaysThing string
			// ...
			return todaysThing, nil
		},
	}
```

### Get
```go
var thing myCoolType
thingCache.Get(&thing)
```

Get works similar to package `json`'s Unmarshal. Pass it a pointer to the variable you want to be set and it will do its best to match up. 

Get works for `*string`, `*int`, `*int{64,32}` right now. I will add support for more basic types soon.

Get treats `encoding.TextUnmarshaler` and `json.Unmarshaler` properly, in that order of priority. 

As a last ditch effort, it will try calling json.Unmarshal() on the Redis data. You can take advantage of this by passing pointers to arbitrary objects if you store JSON data in Redis. 