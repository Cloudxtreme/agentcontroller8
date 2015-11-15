package ds
import (
	"time"
	"github.com/garyburd/redigo/redis"
)

type Value struct {
	Name string
}

// EXPIRE this value
func (value Value) Expire(connPool *redis.Pool, duration time.Duration) error {
	conn := connPool.Get()
	defer conn.Close()

	return conn.Send("EXPIRE", value.Name, duration.Seconds())
}

// SET this value
func (value Value) Set(connPool *redis.Pool, content []byte) error {
	conn := connPool.Get()
	defer conn.Close()

	return conn.Send("SET", value.Name, content)
}

// GET this value
func (value Value) Get(connPool *redis.Pool) ([]byte, error) {
	conn := connPool.Get()
	defer conn.Close()

	return redis.Bytes(conn.Do("GET", value.Name))
}