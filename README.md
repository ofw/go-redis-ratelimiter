# go-redis-ratelimiter
Simple go ratelimiter library with redis backend

# Docs

https://godoc.org/github.com/oeegor/go-redis-ratelimiter

# Usage

```go

import (
	"github.com/garyburd/redigo/redis"
    "github.com/oeegor/go-redis-ratelimiter"
)

func main() {

    // initialize redis pool
    redisPool := redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", "your-redis-address")
		if err != nil {
			return nil, err
		}
		return c, err
	}, 100)  // also set max connections to 100

    // increment rate limit usage for given key that is allowed 10 requests per second
    limitCtx, err := ratelimiter.Incr(redisPool, "mykey", 10, time.Second)

    if err != nil {
        // do something
    }

    // limitCtx contains all necessary data for ratelimiter state
    if limitCtx.Reached() {
        // code to handle over the limit logic
    }
}
```
