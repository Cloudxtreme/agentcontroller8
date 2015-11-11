package ds
import (
	"time"
	"github.com/garyburd/redigo/redis"
	"github.com/Jumpscale/agentcontroller2/core"
)

type CommandResultList struct {
	List List
}

func (list CommandResultList) BlockingLeftPop(connPool *redis.Pool,
	timeout time.Duration) (*core.CommandResult, error) {
	jsonData, err := list.List.BlockingLeftPop(connPool, timeout)
	if err != nil {
		return nil, err
	}
	return core.CommandResultFromJSON(jsonData)
}

func (list CommandResultList) LeftPush(connPool *redis.Pool, message *core.CommandResult) error {
	return list.List.LeftPush(connPool, message.JSON)
}

func (list CommandResultList) RightPush(connPool *redis.Pool, message *core.CommandResult) error {
	return list.List.RightPush(connPool, message.JSON)
}