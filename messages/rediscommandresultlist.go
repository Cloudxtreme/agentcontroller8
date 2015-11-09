package messages
import (
	"github.com/Jumpscale/agentcontroller2/redisdata/ds"
	"time"
	"github.com/garyburd/redigo/redis"
)

type RedisCommandResultList struct {
	List ds.List
}

func (list RedisCommandResultList) BlockingPop(connPool *redis.Pool, timeout time.Duration) (*CommandResultMessage, error) {
	jsonData, err := list.List.BlockingPop(connPool, timeout)
	if err != nil {
		return nil, err
	}
	return CommandResultMessageFrom(jsonData)
}

func (list RedisCommandResultList) LeftPush(connPool *redis.Pool, message *CommandResultMessage) error {
	return list.List.LeftPush(connPool, message.Payload)
}

func (list RedisCommandResultList) RightPush(connPool *redis.Pool, message *CommandResultMessage) error {
	return list.List.RightPush(connPool, message.Payload)
}