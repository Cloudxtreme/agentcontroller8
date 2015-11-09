package ds
import (
	"github.com/garyburd/redigo/redis"
	"time"
	"fmt"
	"encoding/json"
)

type List struct {
	Name	string
}

// BLPOP from the list, decode from JSON into the target pointer.
func (list *List) BlockingPop(connPool *redis.Pool, target interface{}, timeout time.Duration) error {
	conn := connPool.Get()
	defer conn.Close()

	reply, err := redis.Strings(conn.Do("BLPOP", list.Name, fmt.Sprintf("%d", timeout)))

	if err != nil {
		return err
	}

	popped := reply[1]

	err = json.Unmarshal([]byte(popped), target)
	if err != nil {
		return err
	}

	return nil
}

// LPUSH the JSON-encoded object onto the list.
func (list *List) LeftPush(connPool *redis.Pool, object interface{}) error {
	conn := connPool.Get()
	defer conn.Close()

	encoded, err := json.Marshal(object)
	if err != nil {
		return err
	}

	err = conn.Send("LPUSH", list.Name, []byte(encoded))
	return err
}

// RPUSH the JSON-encoded object onto the list.
func (list *List) RightPush(connPool *redis.Pool, object interface{}) error {
	conn := connPool.Get()
	defer conn.Close()

	encoded, err := json.Marshal(object)
	if err != nil {
		return err
	}

	err = conn.Send("RPUSH", list.Name, []byte(encoded))
	return err
}