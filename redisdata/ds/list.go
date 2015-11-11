package ds

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"time"
)

type List struct {
	Name string
}

// BLPOP from the list
func (list List) BlockingLeftPop(connPool *redis.Pool, timeout time.Duration) ([]byte, error) {
	conn := connPool.Get()
	defer conn.Close()

	reply, err := redis.Strings(conn.Do("BLPOP", list.Name, fmt.Sprintf("%d", timeout)))
	if err != nil {
		return nil, err
	}

	return []byte(reply[1]), nil
}

// LPUSH onto the list.
func (list List) LeftPush(connPool *redis.Pool, data []byte) error {
	conn := connPool.Get()
	defer conn.Close()

	return conn.Send("LPUSH", list.Name, data)
}

// RPUSH onto the list.
func (list List) RightPush(connPool *redis.Pool, data []byte) error {
	conn := connPool.Get()
	defer conn.Close()

	return conn.Send("RPUSH", list.Name, data)
}
