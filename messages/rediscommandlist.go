package messages
import (
	"github.com/Jumpscale/agentcontroller2/redisdata/ds"
	"github.com/garyburd/redigo/redis"
	"time"
)

type RedisCommandList struct {
	List ds.List
}

func (list RedisCommandList) BlockingLeftPop(connPool *redis.Pool, timeout time.Duration) (*CommandMessage, error) {
	jsonData, err := list.List.BlockingLeftPop(connPool, timeout)
	if err != nil {
		return nil, err
	}
	return CommandMessageFromJSON(jsonData)
}

func (list RedisCommandList) LeftPush(connPool *redis.Pool, message *CommandMessage) error {
	return list.List.LeftPush(connPool, message.Payload)
}

func (list RedisCommandList) RightPush(connPool *redis.Pool, message *CommandMessage) error {
	return list.List.RightPush(connPool, message.Payload)
}