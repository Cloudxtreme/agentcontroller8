package ds
import (
	"github.com/garyburd/redigo/redis"
)

type Hash struct {
	Name string
}

// HSET the JSON-encoded value on this Hash.
func (hash Hash) Set(connPool *redis.Pool, key string, data []byte) error {
	conn := connPool.Get()
	defer conn.Close()

	return conn.Send("HSET", hash.Name, key, data)
}