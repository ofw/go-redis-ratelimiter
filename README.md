# go-redis-ratelimiter
Simple go ratelimiter library with redis backend

# Docs

https://godoc.org/github.com/oeegor/go-redis-ratelimiter

# Usage

```go

import (
    "github.com/oeegor/go-redis-ratelimiter"
	"github.com/garyburd/redigo/redis"
)

// increment rate limit usage for given key that is allowed 10 requests per second
limitCtx, err := ratelimiter.Incr(utils.Redis, "mykey", 10, time.Second)

if err != nil {
    // do something
}

// limitCtx contains all necessary data for ratelimiter state
if limitCtx.Reached() {
    // code to handle over the limit logic
}
```
