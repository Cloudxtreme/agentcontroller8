package ds
import (
	"time"
	"github.com/garyburd/redigo/redis"
	"github.com/Jumpscale/agentcontroller2/messages"
)

type CommandResultList struct {
	List List
}

func (list CommandResultList) BlockingLeftPop(connPool *redis.Pool,
	timeout time.Duration) (*messages.CommandResultMessage, error) {
	jsonData, err := list.List.BlockingLeftPop(connPool, timeout)
	if err != nil {
		return nil, err
	}
	return messages.CommandResultMessageFromJSON(jsonData)
}

func (list CommandResultList) LeftPush(connPool *redis.Pool, message *messages.CommandResultMessage) error {
	return list.List.LeftPush(connPool, message.Payload)
}

func (list CommandResultList) RightPush(connPool *redis.Pool, message *messages.CommandResultMessage) error {
	return list.List.RightPush(connPool, message.Payload)
}