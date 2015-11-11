package ds
import (
	"github.com/garyburd/redigo/redis"
	"time"
)

type Hash struct {
	Name string
}

// HSET to the hash
func (hash Hash) Set(connPool *redis.Pool, key string, data []byte) error {
	conn := connPool.Get()
	defer conn.Close()

	return conn.Send("HSET", hash.Name, key, data)
}

// EXPIRE thie hash
func (hash Hash) Expire(connPool *redis.Pool, duration time.Duration) error {
	conn := connPool.Get()
	defer conn.Close()

	return conn.Send("EXPIRE", hash.Name, duration.Seconds())
}