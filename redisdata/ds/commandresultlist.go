package ds
import (
	"time"
	"github.com/garyburd/redigo/redis"
	"github.com/Jumpscale/agentcontroller8/core"
)

type CommandResultList struct {
	List List
}

func GetCommandResultList(name string) CommandResultList {
	return CommandResultList{GetList(name)}
}

func (list CommandResultList) BlockingLeftPop(connPool *redis.Pool,
	timeout time.Duration) (*core.CommandResponse, error) {
	jsonData, err := list.List.BlockingLeftPop(connPool, timeout)
	if err != nil {
		return nil, err
	}
	return core.CommandResponseFromJSON(jsonData)
}

func (list CommandResultList) LeftPush(connPool *redis.Pool, message *core.CommandResponse) error {
	return list.List.LeftPush(connPool, message.JSON)
}

func (list CommandResultList) RightPush(connPool *redis.Pool, message *core.CommandResponse) error {
	return list.List.RightPush(connPool, message.JSON)
}