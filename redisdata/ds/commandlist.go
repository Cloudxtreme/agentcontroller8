package ds
import (
	"github.com/garyburd/redigo/redis"
	"time"
	"github.com/Jumpscale/agentcontroller8/core"
)

type CommandList struct {
	List List
}

func GetCommandList(name string) CommandList {
	return CommandList{GetList(name)}
}

func (list CommandList) BlockingLeftPop(connPool *redis.Pool, timeout time.Duration) (*core.Command, error) {
	jsonData, err := list.List.BlockingLeftPop(connPool, timeout)
	if err != nil {
		return nil, err
	}
	return core.CommandFromJSON(jsonData)
}

func (list CommandList) LeftPush(connPool *redis.Pool, command *core.Command) error {
	return list.List.LeftPush(connPool, command.JSON)
}

func (list CommandList) RightPush(connPool *redis.Pool, command *core.Command) error {
	return list.List.RightPush(connPool, command.JSON)
}