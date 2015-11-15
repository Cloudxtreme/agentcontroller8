package ds
import (
	"github.com/garyburd/redigo/redis"
)

type Hash struct {
	Value
}

func GetHash(name string) Hash {
	return Hash{Value{Name: name}}
}

// HSET to the hash
func (hash Hash) Set(connPool *redis.Pool, key string, data []byte) error {
	conn := connPool.Get()
	defer conn.Close()

	return conn.Send("HSET", hash.Name, key, data)
}

