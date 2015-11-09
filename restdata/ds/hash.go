package ds
import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
)

type Hash struct {
	Name string
}

// HSET the JSON-encoded value on this Hash.
func (hash *Hash) Set(connPool *redis.Pool, key string, value interface{}) error {

	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	conn := connPool.Get()
	defer conn.Close()

	err = conn.Send("HSET", hash.Name, key, []byte(jsonValue))
	return err
}