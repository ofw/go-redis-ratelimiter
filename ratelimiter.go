package ratelimiter

import (
	"fmt"
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
	RedisPool *redis.Pool
}

func (self *LimitCtx) Remaining() int {
	return self.Limit - self.Current
}

func (self *LimitCtx) Reached() bool {
	return self.Current > self.Limit
}

func (self *LimitCtx) Incr() error {

	c := self.RedisPool.Get()
	defer c.Close()
	key := fmt.Sprintf("rl:%v:%v", self.Key, self.ExpireAt)
	c.Send("MULTI")
	c.Send("INCR", key)
	c.Send("EXPIREAT", key, self.ExpireAt+expirationWindow)
	r, err := redis.Ints(c.Do("EXEC"))
	self.Current = r[0]
	return err
}

func BuildLimiter(redisPool *redis.Pool, key string, limit int, per time.Duration) *LimitCtx {
	perSeconds := per.Seconds()
	now := float64(time.Now().Unix())
	expireAt := math.Floor(now/perSeconds)*perSeconds + perSeconds
	return &LimitCtx{
		Key:       key,
		Limit:     limit,
		Per:       per,
		RedisPool: redisPool,
		ExpireAt:  int64(expireAt),
	}
}

func CheckLimit(redisPool *redis.Pool, name string, limit int, period time.Duration) (*LimitCtx, error) {
	limitCtx := BuildLimiter(redisPool, name, limit, period)
	err := limitCtx.Incr()
	return limitCtx, err
}
