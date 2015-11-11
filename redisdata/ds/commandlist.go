package ds
import (
	"github.com/garyburd/redigo/redis"
	"time"
	"github.com/Jumpscale/agentcontroller2/messages"
)

type CommandList struct {
	List List
}

func (list CommandList) BlockingLeftPop(connPool *redis.Pool, timeout time.Duration) (*messages.CommandMessage, error) {
	jsonData, err := list.List.BlockingLeftPop(connPool, timeout)
	if err != nil {
		return nil, err
	}
	return messages.CommandMessageFromJSON(jsonData)
}

func (list CommandList) LeftPush(connPool *redis.Pool, message *messages.CommandMessage) error {
	return list.List.LeftPush(connPool, message.Payload)
}

func (list CommandList) RightPush(connPool *redis.Pool, message *messages.CommandMessage) error {
	return list.List.RightPush(connPool, message.Payload)
}