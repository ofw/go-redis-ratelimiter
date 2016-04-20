// Package ratelimiter provides possibility to check
// rate limit usage by given resource with given allowed rate
// and time interval. It uses redis as backend so can be used
// to check ratelimit for distributed instances of your app.
package ratelimiter

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/garyburd/redigo/redis"
)

const expirationWindow = 60

type LimitCtx struct {
	// ported from here: http://flask.pocoo.org/snippets/70/
	ExpireAt  int64
	Key       string
	Limit     int
	Per       time.Duration
	Current   int
	Tries     int
	RedisPool *redis.Pool
}

// Returns how many times resource can be used
// before reaching limit
func (self *LimitCtx) Remaining() int {
	return self.Limit - self.Current
}

// Returns whether limit has been reached or not
func (self *LimitCtx) Reached() bool {
	return self.Current > self.Limit
}

// Calculates and sets ExpireAt
func (self *LimitCtx) setExpireAt() {
	perSeconds := self.Per.Seconds()
	now := float64(time.Now().Unix())
	expireAt := math.Floor(now/perSeconds)*perSeconds + perSeconds
	self.ExpireAt = int64(expireAt)
}

// Increments rate limit counter
func (self *LimitCtx) Incr() error {
	c := self.RedisPool.Get()
	defer c.Close()

	for self.Tries > 0 {
		log.Println("try to incr ratelimiter, tries:", self.Tries)
		self.Tries--

		key := fmt.Sprintf("rl:%v:%v", self.Key, self.ExpireAt)
		c.Send("MULTI")
		c.Send("INCR", key)
		c.Send("EXPIREAT", key, self.ExpireAt+expirationWindow)
		r, err := redis.Ints(c.Do("EXEC"))
		if err != nil {
			return err
		}
		self.Current = r[0]
		if !self.Reached() {
			return nil
		}
		sleep := self.ExpireAt - time.Now().Unix()
		log.Println("ratelimit reached, try to sleep for:", sleep)
		if sleep > 0 {
			time.Sleep(time.Duration(sleep) * time.Second)
		}
		self.setExpireAt()
	}
	return nil
}

// Initializes new LimiterCtx instance which then can be used
// to increment and check ratelimit usage
func BuildLimiter(redisPool *redis.Pool, key string, limit int, per time.Duration, tries int) *LimitCtx {
	ctx := &LimitCtx{
		Key:       key,
		Limit:     limit,
		Per:       per,
		RedisPool: redisPool,
		Tries:     tries,
	}
	ctx.setExpireAt()
	return ctx
}

// Shorthand function to increment resource usage
// and to get LimiterCtx back. Wrapper around BuildLimiter and LimiterCtx.Incr
func Incr(redisPool *redis.Pool, name string, limit int, period time.Duration, retries int) (*LimitCtx, error) {
	limitCtx := BuildLimiter(redisPool, name, limit, period, retries)
	err := limitCtx.Incr()
	return limitCtx, err
}
