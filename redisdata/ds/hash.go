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

// HDEL the specified key on the hash
// Returns true if the key was found and deleted, false otherwise.
func (hash Hash) Delete(connPool *redis.Pool, key string) (bool, error) {
	conn := connPool.Get()
	defer conn.Close()

	intReply, err := redis.Int(conn.Do("HDEL", hash.Name, key))
	if intReply == 0 {
		return false, err
	}
	return true, err
}

// HGETALL this hash as a string -> string mapping
func (hash Hash) ToStringMap(connPool *redis.Pool) (map[string]string, error) {
	conn := connPool.Get()
	defer conn.Close()

	return redis.StringMap(conn.Do("HGETALL", hash.Name))
}